package main

// util.go — HTTPzip/tarSHA-1
import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha1"
	"hash"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ===  ===
func launcherDir() string {
	if v := os.Getenv("MC_LAUNCHER_DIR"); v != "" { return v }
	exe, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(exe)
		if filepath.Base(dir) == "bin" { dir = filepath.Dir(dir) }
		cand := filepath.Join(dir, "launcher")
		if info, _ := os.Stat(cand); info != nil { return cand }
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mc-launcher")
}
func ensureDir(p string) error { return os.MkdirAll(p, 0755) }
func pjoin(parts ...string) string { return filepath.Join(parts...) }
func plistDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil { return nil, err }
	var out []string
	for _, e := range entries { if !e.IsDir() { out = append(out, e.Name()) } }
	return out, nil
}
func listDirs(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil { return nil, err }
	var out []string
	for _, e := range entries { if e.IsDir() { out = append(out, e.Name()) } }
	return out, nil
}

// ===  ===
func readAll(path string) (string, error) {
	b, err := os.ReadFile(path)
	return string(b), err
}
func writeAll(path, data string) error {
	ensureDir(filepath.Dir(path))
	return os.WriteFile(path, []byte(data), 0644)
}
func writeBin(path string, data []byte) error {
	ensureDir(filepath.Dir(path))
	return os.WriteFile(path, data, 0644)
}
func fileExists(path string) bool {
	_, err := os.Stat(path); return err == nil
}
func fileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil { return 0, err }
	return info.Size(), nil
}

// === SHA-1 ===
func sha1Of(data []byte) string {
	h := sha1.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// sha1File 流式计算文件 SHA-1（旧实现把整个文件读进内存）
func sha1File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// === HTTP ===
var httpClient = &http.Client{Timeout: 120 * time.Second}

func httpGet(url string, headers ...string) ([]byte, int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, err
	}
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	req.Header.Set("User-Agent", "mc-launcher/2.0")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return body, resp.StatusCode, err
}

func httpPostJSON(url, body string, headers ...string) ([]byte, int, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	req.Header.Set("User-Agent", "mc-launcher/2.0")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	return b, resp.StatusCode, err
}

// httpPostForm 表单编码 POST（微软 OAuth 端点要求 form 而非 JSON）
func httpPostForm(url, form string) ([]byte, int, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(form))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "mc-launcher/2.0")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	return b, resp.StatusCode, err
}

// downloadFile 下载到目标路径。
//  - 有期望 SHA-1 时：先校验已有文件，命中直接返回；否则下载到 .tmp 校验后 rename（原子，不会留下半截文件）。
//  - 无 SHA-1 时：支持 Range 断点续传。
func downloadFile(url, dest, expectedSHA1 string, progress func(done, total int64)) error {
	if url == "" {
		return fmt.Errorf("下载地址为空")
	}
	if err := ensureDir(filepath.Dir(dest)); err != nil {
		return err
	}

	// 已有且校验通过 -> 跳过
	if expectedSHA1 != "" && fileExists(dest) {
		if cur, err := sha1File(dest); err == nil && strings.EqualFold(cur, expectedSHA1) {
			return nil
		}
	}

	if expectedSHA1 != "" {
		tmp := dest + ".tmp"
		if err := fetchToFile(url, tmp, 0, expectedSHA1, progress); err != nil {
			os.Remove(tmp)
			return err
		}
		os.Remove(dest)
		return os.Rename(tmp, dest)
	}

	var offset int64
	if fileExists(dest) {
		if s, err := fileSize(dest); err == nil {
			offset = s
		}
	}
	err := fetchToFile(url, dest, offset, "", progress)
	if err != nil && offset > 0 {
		// 断点续传失败：重头再来一次
		os.Remove(dest)
		return fetchToFile(url, dest, 0, "", progress)
	}
	return err
}

// fetchToFile 从 offset 起把 url 写入 dest。
// offset>0 时以追加方式打开，并在 416（区间无效）时从头重下。
func fetchToFile(url, dest string, offset int64, expectedSHA1 string, progress func(done, total int64)) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "mc-launcher/2.0")
	if offset > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", offset))
	}
	client := &http.Client{Timeout: 30 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	// 区间无效 -> 关掉本次响应，重新从头下载
	if resp.StatusCode == 416 && offset > 0 {
		resp.Body.Close()
		os.Remove(dest)
		return fetchToFile(url, dest, 0, expectedSHA1, progress)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 206 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, url)
	}
	// 服务器忽略 Range 返回 200 -> 不能追加，必须截断重写
	flag := os.O_CREATE | os.O_WRONLY
	if resp.StatusCode == 206 {
		flag |= os.O_APPEND
	} else {
		offset = 0
		flag |= os.O_TRUNC
	}
	out, err := os.OpenFile(dest, flag, 0644)
	if err != nil {
		return err
	}
	defer out.Close()

	var h hash.Hash
	if expectedSHA1 != "" {
		h = sha1.New()
	}
	buf := make([]byte, 256*1024)
	written := offset
	for {
		nr, er := resp.Body.Read(buf)
		if nr > 0 {
			nw, ew := out.Write(buf[:nr])
			if nw > 0 && h != nil {
				h.Write(buf[:nw])
			}
			written += int64(nw)
			if ew != nil {
				return ew
			}
			if nr != nw {
				return fmt.Errorf("写入不完整")
			}
		}
		if er != nil {
			if er != io.EOF {
				return er
			}
			break
		}
		if progress != nil {
			progress(written, resp.ContentLength)
		}
	}
	if err := out.Close(); err != nil {
		return err
	}
	if h != nil {
		if !strings.EqualFold(hex.EncodeToString(h.Sum(nil)), expectedSHA1) {
			os.Remove(dest)
			return fmt.Errorf("SHA-1 校验失败: %s", filepath.Base(dest))
		}
	}
	return nil
}

func mirrorUrl(url, source string) string {
	if source != "bmc" { return url }
	repls := map[string]string{
		"https://launchermeta.mojang.com": "https://bmclapi2.bangbang93.com",
		"https://launcher.mojang.com":     "https://bmclapi2.bangbang93.com",
		"https://resources.download.minecraft.net": "https://bmclapi2.bangbang93.com/assets",
		"https://libraries.minecraft.net": "https://bmclapi2.bangbang93.com/maven",
		"https://piston-meta.mojang.com":  "https://bmclapi2.bangbang93.com",
		"https://piston-data.mojang.com":  "https://bmclapi2.bangbang93.com",
	}
	for old, n := range repls {
		if strings.HasPrefix(url, old) { return n + url[len(old):] }
	}
	return url
}

// ===  ===
func unzip(src, dest string, exclude []string) error {
	r, err := zip.OpenReader(src)
	if err != nil { return err }
	defer r.Close()
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, "/") { continue }
		skip := false
		for _, e := range exclude { if strings.HasPrefix(f.Name, e) { skip = true; break } }
		if skip { continue }
		fp := filepath.Join(dest, f.Name)
		ensureDir(filepath.Dir(fp))
		rc, err := f.Open()
		if err != nil { continue }
		out, err := os.Create(fp)
		if err != nil { rc.Close(); continue }
		io.Copy(out, rc); out.Close(); rc.Close()
	}
	return nil
}
// readZipEntry 读取 zip 中某个条目的文本内容
func readZipEntry(archive, name string) (string, error) {
	r, err := zip.OpenReader(archive)
	if err != nil {
		return "", err
	}
	defer r.Close()
	for _, f := range r.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()
		b, err := io.ReadAll(rc)
		return string(b), err
	}
	return "", fmt.Errorf("zip 中没有 %s", name)
}

// unzipPrefix 只解压指定前缀下的条目，并去掉该前缀
func unzipPrefix(src, prefix, dest string) error {
	found := false
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		if !strings.HasPrefix(f.Name, prefix) || strings.HasSuffix(f.Name, "/") {
			continue
		}
		found = true
		rel := strings.TrimPrefix(f.Name, prefix)
		if rel == "" || strings.HasPrefix(rel, "..") {
			continue
		}
		fp := filepath.Join(dest, filepath.FromSlash(rel))
		if err := ensureDir(filepath.Dir(fp)); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		out, err := os.Create(fp)
		if err != nil {
			rc.Close()
			continue
		}
		_, err = io.Copy(out, rc)
		out.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	_ = found
	return nil
}

func untarGz(src, dest string) error {
	f, err := os.Open(src)
	if err != nil { return err }
	defer f.Close()
	gr, err := gzip.NewReader(f)
	if err != nil { return err }
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		h, err := tr.Next()
		if err == io.EOF { break }
		if err != nil { continue }
		fp := filepath.Join(dest, h.Name)
		switch h.Typeflag {
		case tar.TypeDir: ensureDir(fp)
		case tar.TypeReg:
			ensureDir(filepath.Dir(fp))
			out, _ := os.Create(fp)
			io.Copy(out, tr); out.Close()
		}
	}
	return nil
}
