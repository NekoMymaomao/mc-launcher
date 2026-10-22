package main

// install.go — 
import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func getInstalledDir(versionId string) string {
	if cfg.VersionIsolation {
		return filepath.Join(baseDir, "versions", versionId)
	}
	return filepath.Join(baseDir, ".minecraft", "versions", versionId)
}
func getMinecraftDir() string {
	if cfg.VersionIsolation {
		return filepath.Join(baseDir, ".minecraft")
	}
	return filepath.Join(baseDir, ".minecraft")
}
func getLibDir(versionId string) string {
	if cfg.VersionIsolation {
		return filepath.Join(baseDir, "libraries")
	}
	return filepath.Join(getMinecraftDir(), "libraries")
}

func installVersion(versionId string, cb string, sender func(Progress)) (*Response, error) {
	var wg sync.WaitGroup
	detail, err := fetchVersionDetail(versionId, cfg.DownloadSource)
	if err != nil {
		return &Response{Ok: false, Error: ": " + err.Error()}, nil
	}
	if !detail.Valid {
		return &Response{Ok: false, Error: ""}, nil
	}
	if cb != "" { sender(Progress{ID: cb, Msg: "", Done: 1, Total: 100}) }

	instDir := getInstalledDir(versionId)
	libDir := getLibDir(versionId)
	ensureDir(instDir)
	ensureDir(libDir)

	// save manifest
	writeAll(filepath.Join(instDir, versionId+".json"), "")

	var totalTasks int
	var downloaded int64
	var mu sync.Mutex

	updateProgress := func(done int64, msg string) {
		mu.Lock(); downloaded = done; mu.Unlock()
		if cb != "" { sender(Progress{ID: cb, Done: downloaded, Total: totalTasks, Msg: msg}) }
	}

	// 1. client jar
	totalTasks++
	go func() {
		dest := filepath.Join(instDir, versionId+".jar")
		downloadFile(detail.Client.URL, dest, detail.Client.SHA1, nil)
		mu.Lock(); downloaded++; mu.Unlock()
	}()
	wg.Add(1)

	// 2. libraries (parallel)
	sem := make(chan struct{}, cfg.Threads)
	for _, lib := range detail.Libraries {
		if lib.IsNative || lib.ResolvedURL == "" { continue }
		totalTasks++
		sem <- struct{}{}
		wg.Add(1)
		go func(l Library) {
			defer func() { <-sem; wg.Done() }()
			dest := filepath.Join(libDir, l.ResolvedPath)
			downloadFile(l.ResolvedURL, dest, "", nil)
			mu.Lock(); downloaded++; mu.Unlock()
		}(lib)
	}

	// 3. asset index
	totalTasks++
	go func() {
		ai := detail.AssetIndex
		if ai.URL != "" {
			dest := filepath.Join(baseDir, "assets", "indexes", ai.ID+".json")
			downloadFile(ai.URL, dest, ai.SHA1, nil)
			updateProgress(downloaded+1, "")
		}
		mu.Lock(); downloaded++; mu.Unlock()
	}()
	wg.Add(1)

	// 4. asset objects (parallel)
	if detail.AssetIndex.URL != "" {
		indexPath := filepath.Join(baseDir, "assets", "indexes", detail.AssetIndex.ID+".json")
		if idxBody, err := readAll(indexPath); err == nil {
			var idx struct { Objects map[string]struct{ Hash string; Size int64 } }
			if json.Unmarshal([]byte(idxBody), &idx) == nil {
				for _, obj := range idx.Objects {
					hash := obj.Hash
					if len(hash) == 40 {
						totalTasks++
						sem <- struct{}{}
						wg.Add(1)
						go func(h string) {
							defer func() { <-sem; wg.Done() }()
							subdir := filepath.Join(baseDir, "assets", "objects", h[:2])
							dest := filepath.Join(subdir, h)
							url := "https://resources.download.minecraft.net/" + h[:2] + "/" + h
							downloadFile(mirrorUrl(url, cfg.DownloadSource), dest, h, nil)
							mu.Lock(); downloaded++; mu.Unlock()
						}(hash)
					}
				}
			}
		}
	}

	// 5. natives (parallel)
	for _, lib := range detail.Libraries {
		if !lib.IsNative || lib.NativeURL == "" { continue }
		totalTasks++
		sem <- struct{}{}
		wg.Add(1)
		go func(l Library) {
			defer func() { <-sem; wg.Done() }()
			dest := filepath.Join(libDir, l.NativePath)
			if downloadFile(l.NativeURL, dest, "", nil) == nil {
				// extract natives
				nativesDir := filepath.Join(instDir, "natives")
				exclude := []string{"META-INF/"}
				if l.Extract != nil { exclude = l.Extract.Exclude }
				unzip(dest, nativesDir, exclude)
			}
			mu.Lock(); downloaded++; mu.Unlock()
		}(lib)
	}

	wg.Wait()
	updateProgress(int64(totalTasks), "")
	return &Response{Ok: true, Data: map[string]string{"id": versionId, "path": instDir}}, nil
}

func listVersions() []string {
	//  installed versions  + minecraft 
	var out []string
	checkDir := func(dir string) {
		dirs, err := listDirs(dir)
		if err != nil { return }
		for _, d := range dirs {
			if fileExists(filepath.Join(dir, d, d+".json")) {
				out = append(out, d)
			}
		}
	}
	checkDir(filepath.Join(baseDir, "versions"))
	checkDir(filepath.Join(baseDir, ".minecraft", "versions"))
	return out
}

func removeVersion(versionId string) error {
	for _, base := range []string{
		filepath.Join(baseDir, "versions", versionId),
		filepath.Join(baseDir, ".minecraft", "versions", versionId),
	} {
		if info, err := os.Stat(base); err == nil && info.IsDir() {
			os.RemoveAll(base)
		}
	}
	return nil
}
