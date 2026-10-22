//go:build !cshared



// util.go — HTTPzip/tarSHA-1
import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha1"
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
	h := sha1.New(); h.Write(data); return hex.EncodeToString(h.Sum(nil))
}
func sha1File(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil { return "", err }
	return sha1Of(b), nil
}

// === HTTP ===
func httpGet(url string, headers ...string) ([]byte, int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil { return nil, 0, err }
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	req.Header.Set("User-Agent", "MC-Launcher/1.0")
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil { return nil, 0, err }
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return body, resp.StatusCode, err
}
func httpPostJSON(url, body string, headers ...string) ([]byte, int, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil { return nil, 0, err }
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	req.Header.Set("User-Agent", "MC-Launcher/1.0")
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil { return nil, 0, err }
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	return b, resp.StatusCode, err
}

// ===  + SHA1 +  ===
func downloadFile(url, dest, expectedSHA1 string, progress func(done, total int64)) error {
	if expectedSHA1 != "" && fileExists(dest) {
		cur, err := sha1File(dest)
		if err == nil && cur == expectedSHA1 { return nil }
	}
	offset := int64(0)
	if fileExists(dest) {
		if s, err := fileSize(dest); err == nil { offset = s }
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil { return err }
	if offset > 0 { req.Header.Set("Range", fmt.Sprintf("bytes=%d-", offset)) }
	client := &http.Client{Timeout: 600 * time.Second}
	resp, err := client.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()
	if resp.StatusCode == 416 {
		os.Remove(dest); offset = 0
		req.Header.Del("Range")
		resp, err = client.Do(req)
		if err != nil { return err }
		defer resp.Body.Close()
	}
	if resp.StatusCode != 200 && resp.StatusCode != 206 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	out, err := os.Create(dest)
	if err != nil { return err }
	defer out.Close()
	var h sha1.Hash
	if expectedSHA1 != "" { h = sha1.New() }
	buf := make([]byte, 128*1024)
	var written int64
	for {
		nr, er := resp.Body.Read(buf)
		if nr > 0 {
			nw, ew := out.Write(buf[0:nr])
			if nw > 0 && h != nil { h.Write(buf[0:nw]) }
			written += int64(nw)
			if ew != nil { out.Close(); return ew }
			if nr != nw { out.Close(); return fmt.Errorf("short write") }
		}
		if er != nil {
			if er != io.EOF { out.Close(); return er }
			break
		}
		if progress != nil { progress(written, resp.ContentLength) }
	}
	out.Close()
	if h != nil && expectedSHA1 != "" {
		if hex.EncodeToString(h.Sum(nil)) != expectedSHA1 {
			os.Remove(dest); return fmt.Errorf("sha1 mismatch")
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
