package main

// install.go — 版本安装
//
// 旧的实现有几个致命问题：
//  1. client jar 的 goroutine 里没有 wg.Done()，wg.Wait() 永远返回不了（卡死）；
//  2. wg.Add(1) 写在 go 之后，和 Done() 竞争，会 panic("negative WaitGroup counter")；
//  3. 资源索引还没下载完就去读它，assets 永远装不上；
//  4. 安装开始时把 <version>.json 写成空字符串，下次启动直接坏掉。
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

func getInstalledDir(versionId string) string {
	return filepath.Join(baseDir, "versions", versionId)
}
func getMinecraftDir() string { return filepath.Join(baseDir, ".minecraft") }
func getLibDir(versionId string) string {
	return filepath.Join(baseDir, "libraries")
}

// defaultThreads 并发下载数
func defaultThreads() int {
	n := cfg.Threads
	if n <= 0 {
		n = 8
	}
	if n > 64 {
		n = 64
	}
	return n
}

// installVersion 安装指定版本；cb 非空时通过 sender 汇报进度
func installVersion(versionId string, cb string, sender func(Progress)) (Response, error) {
	if versionId == "" {
		return errResp("版本号为空"), nil
	}
	detail, err := fetchVersionDetail(versionId, cfg.DownloadSource)
	if err != nil {
		return errResp(err.Error()), nil
	}
	if !detail.Valid {
		return errResp("版本元数据无效"), nil
	}

	instDir := getInstalledDir(versionId)
	libDir := getLibDir(versionId)
	assetsDir := filepath.Join(baseDir, "assets")
	nativesDir := filepath.Join(instDir, "natives")
	for _, d := range []string{instDir, libDir, nativesDir, filepath.Join(assetsDir, "indexes"), filepath.Join(assetsDir, "objects")} {
		if err := ensureDir(d); err != nil {
			return errResp("创建目录失败: " + err.Error()), nil
		}
	}

	// 先把真正的版本 json 落盘（旧实现写空串，导致下次启动失败）
	if err := writeAll(filepath.Join(instDir, versionId+".json"), detail.RawJSON); err != nil {
		return errResp("写入版本文件失败: " + err.Error()), nil
	}

	var (
		mu       sync.Mutex
		done     int
		total    int
		errs     []string
		wg       sync.WaitGroup
		sem      = make(chan struct{}, defaultThreads())
	)

	report := func(msg string) {
		if cb == "" || sender == nil {
			return
		}
		mu.Lock()
		d, t := done, total
		mu.Unlock()
		sender(Progress{ID: cb, Done: int64(d), Total: int64(t), Msg: msg})
	}
	mark := func() {
		mu.Lock()
		done++
		mu.Unlock()
	}
	addTask := func(n int) {
		mu.Lock()
		total += n
		mu.Unlock()
	}
	addErr := func(scope string, err error) {
		if err == nil {
			return
		}
		mu.Lock()
		errs = append(errs, scope+": "+err.Error())
		mu.Unlock()
	}

	// run 提交一个并发下载任务（先占位再启动，避免 Add 与 Done 竞争）
	run := func(fn func()) {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer func() {
				<-sem
				mark()
				wg.Done()
			}()
			fn()
		}()
	}

	// ---------- 1. client jar / 日志配置 ----------
	addTask(1)
	run(func() {
		dest := filepath.Join(instDir, versionId+".jar")
		if err := downloadFile(detail.Client.URL, dest, detail.Client.SHA1, nil); err != nil {
			addErr("客户端 jar", err)
		}
		report("下载客户端 jar...")
	})

	if detail.LogFile.URL != "" {
		addTask(1)
		run(func() {
			dest := filepath.Join(instDir, logFileName(detail))
			if err := downloadFile(detail.LogFile.URL, dest, detail.LogFile.SHA1, nil); err != nil {
				addErr("日志配置", err)
			}
			report("下载日志配置...")
		})
	}

	// ---------- 2. libraries + natives ----------
	for _, lib := range detail.Libraries {
		l := lib
		if l.ResolvedURL != "" {
			addTask(1)
			run(func() {
				dest := filepath.Join(libDir, l.ResolvedPath)
				if err := downloadFile(l.ResolvedURL, dest, l.ResolvedSHA1, nil); err != nil {
					addErr("库 "+l.Name, err)
				}
				report("下载库: " + l.Name)
			})
		}
		if l.IsNative && l.NativeURL != "" {
			addTask(1)
			run(func() {
				dest := filepath.Join(libDir, l.NativePath)
				if err := downloadFile(l.NativeURL, dest, l.NativeSHA1, nil); err != nil {
					addErr("natives "+l.Name, err)
					return
				}
				exclude := []string{"META-INF/"}
				if l.Extract != nil && len(l.Extract.Exclude) > 0 {
					exclude = l.Extract.Exclude
				}
				if err := unzip(dest, nativesDir, exclude); err != nil {
					addErr("解压 natives "+l.Name, err)
				}
				report("解压 natives: " + l.Name)
			})
		}
	}

	// ---------- 3. 资源索引（必须先完成）----------
	indexPath := filepath.Join(assetsDir, "indexes", detail.AssetIndexID+".json")
	if detail.AssetIndex.URL != "" {
		addTask(1)
		run(func() {
			if err := downloadFile(detail.AssetIndex.URL, indexPath, detail.AssetIndex.SHA1, nil); err != nil {
				addErr("资源索引", err)
			}
			report("下载资源索引...")
		})
	}

	wg.Wait() // 等索引真正落盘，再读它

	// ---------- 4. 资源对象 ----------
	objects := loadAssetObjects(indexPath)
	if len(objects) > 0 {
		addTask(len(objects))
		for hash := range objects {
			h := hash
			run(func() {
				dest := filepath.Join(assetsDir, "objects", h[:2], h)
				url := mirrorUrl("https://resources.download.minecraft.net/"+h[:2]+"/"+h, cfg.DownloadSource)
				if err := downloadFile(url, dest, h, nil); err != nil {
					addErr("资源 "+h[:8], err)
				}
			})
		}
		wg.Wait()
		report("资源下载完成")
	}

	// ---------- 5. 收尾 ----------
	mu.Lock()
	failed := len(errs)
	firstErr := ""
	if failed > 0 {
		firstErr = errs[0]
	}
	mu.Unlock()

	if failed > 0 {
		return errResp(fmt.Sprintf("%d 个文件下载失败，首个错误: %s", failed, firstErr)), nil
	}
	return okResp(map[string]any{"id": versionId, "path": instDir}), nil
}

// loadAssetObjects 读取资源索引里的对象哈希集合
func loadAssetObjects(indexPath string) map[string]struct{} {
	body, err := readAll(indexPath)
	if err != nil || body == "" {
		return nil
	}
	var idx struct {
		Objects map[string]struct {
			Hash string `json:"hash"`
			Size int64  `json:"size"`
		} `json:"objects"`
	}
	if err := json.Unmarshal([]byte(body), &idx); err != nil {
		return nil
	}
	out := make(map[string]struct{}, len(idx.Objects))
	for _, obj := range idx.Objects {
		if len(obj.Hash) == 40 {
			out[obj.Hash] = struct{}{}
		}
	}
	return out
}

// logFileName 日志配置在版本目录下的文件名
func logFileName(d *VersionDetail) string {
	if d != nil && d.LogFile.Path != "" {
		return filepath.Base(d.LogFile.Path)
	}
	return "logging.xml"
}

func listVersions() []string {
	var out []string
	seen := map[string]bool{}
	checkDir := func(dir string) {
		dirs, err := listDirs(dir)
		if err != nil {
			return
		}
		for _, d := range dirs {
			if seen[d] {
				continue
			}
			if fileExists(filepath.Join(dir, d, d+".json")) {
				seen[d] = true
				out = append(out, d)
			}
		}
	}
	checkDir(filepath.Join(baseDir, "versions"))
	checkDir(filepath.Join(baseDir, ".minecraft", "versions"))
	return out
}

func removeVersion(versionId string) error {
	if versionId == "" {
		return fmt.Errorf("版本号为空")
	}
	// 防止 ../ 之类的路径穿越
	if filepath.Base(versionId) != versionId || versionId == "." || versionId == ".." {
		return fmt.Errorf("非法版本号: %s", versionId)
	}
	for _, base := range []string{
		filepath.Join(baseDir, "versions", versionId),
		filepath.Join(baseDir, ".minecraft", "versions", versionId),
	} {
		if info, err := os.Stat(base); err == nil && info.IsDir() {
			if err := os.RemoveAll(base); err != nil {
				return err
			}
		}
	}
	return nil
}
