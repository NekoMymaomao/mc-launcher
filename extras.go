package main

// extras.go — ////
import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ---  ---
func getModsDir(versionId string) string {
	return filepath.Join(getInstalledDir(versionId), "mods")
}
func listMods(versionId string) []string {
	dir := getModsDir(versionId)
	entries, err := os.ReadDir(dir)
	if err != nil { return nil }
	var out []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".jar") { out = append(out, e.Name()) }
	}
	return out
}
func addMod(versionId, modPath string) bool {
	dir := getModsDir(versionId)
	ensureDir(dir)
	dest := filepath.Join(dir, filepath.Base(modPath))
	src, err := os.Open(modPath)
	if err != nil { return false }
	defer src.Close()
	dst, err := os.Create(dest)
	if err != nil { return false }
	defer dst.Close()
	_, err = dst.ReadFrom(src)
	return err == nil
}
func removeMod(versionId, modName string) bool {
	path := filepath.Join(getModsDir(versionId), modName)
	return os.Remove(path) == nil
}

// --- Modrinth 缓存 ---
func modrinthCacheDir() string { return filepath.Join(baseDir, "mods", "modrinth", "cache") }
func modrinthInstalledDir(slug string) string {
	return filepath.Join(baseDir, "mods", "modrinth", "installed", slug)
}
func modrinthProjectDir(slug string) string {
	return filepath.Join(baseDir, "mods", "modrinth", "projects", slug)
}

// --- 光影 ---
func getShaderDir() string { return filepath.Join(baseDir, "shaderpacks") }
func listShaders() []string {
	dir := getShaderDir()
	entries, err := os.ReadDir(dir)
	if err != nil { return nil }
	var out []string
	for _, e := range entries { if e.IsDir() { out = append(out, e.Name()) } }
	return out
}

// --- 资源包 ---
func getResourcePackDir() string { return filepath.Join(baseDir, "resourcepacks") }
func listResourcePacks() []string {
	dir := getResourcePackDir()
	entries, err := os.ReadDir(dir)
	if err != nil { return nil }
	var out []string
	for _, e := range entries { if e.IsDir() { out = append(out, e.Name()) } }
	return out
}

// --- 数据包 ---
func getDataPackDir() string { return filepath.Join(baseDir, "data_packs") }
func listDataPacks() []string {
	dir := getDataPackDir()
	entries, err := os.ReadDir(dir)
	if err != nil { return nil }
	var out []string
	for _, e := range entries { if e.IsDir() { out = append(out, e.Name()) } }
	return out
}

// ---  ---
func getSaveDir(versionId string) string {
	return filepath.Join(getInstalledDir(versionId), "saves")
}
func listSaves(versionId string) []string {
	dir := getSaveDir(versionId)
	entries, err := os.ReadDir(dir)
	if err != nil { return nil }
	var out []string
	for _, e := range entries { if e.IsDir() { out = append(out, e.Name()) } }
	return out
}
func backupSave(versionId, saveName string) bool {
	src := filepath.Join(getSaveDir(versionId), saveName)
	dest := filepath.Join(getSaveDir(versionId), "..", "backups", saveName+"-"+timeNow())
	ensureDir(filepath.Dir(dest))
	// recursive copy
	filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil { return nil }
		rel, _ := filepath.Rel(src, path)
		newPath := filepath.Join(dest, rel)
		if info.IsDir() { ensureDir(newPath); return nil }
		data, _ := os.ReadFile(path)
		writeBin(newPath, data)
		return nil
	})
	return true
}
func timeNow() string { return strings.ReplaceAll(timeNowStr(), ":", "-") }
func timeNowStr() string {
	return time.Now().Format("2006-01-02_15-04-05")
}
func deleteSave(versionId, saveName string) bool {
	return os.RemoveAll(filepath.Join(getSaveDir(versionId), saveName)) == nil
}

// ---  ---
func serversPath() string { return filepath.Join(baseDir, "servers.json") }

func listServers() []ServerInfo {
	b, err := os.ReadFile(serversPath())
	if err != nil { return nil }
	var out []ServerInfo; json.Unmarshal(b, &out); return out
}
func addServer(addr, name string) {
	if addr == "" {
		return
	}
	servers := listServers()
	for i := range servers {
		if servers[i].Address == addr {
			servers[i].Name = name
			b, _ := json.MarshalIndent(servers, "", "  ")
			os.WriteFile(serversPath(), b, 0644)
			return
		}
	}
	servers = append(servers, ServerInfo{Address: addr, Name: name})
	b, _ := json.MarshalIndent(servers, "", "  ")
	os.WriteFile(serversPath(), b, 0644)
}
func removeServer(addr string) {
	servers := listServers()
	for i, s := range servers {
		if s.Address == addr {
			servers = append(servers[:i], servers[i+1:]...)
			b, _ := json.MarshalIndent(servers, "", "  ")
			os.WriteFile(serversPath(), b, 0644)
			return
		}
	}
}

// ---  ---
func listCrashes(versionId string) []string {
	dir := filepath.Join(getInstalledDir(versionId), "crash-reports")
	entries, err := os.ReadDir(dir)
	if err != nil { return nil }
	var out []string
	for _, e := range entries { if !e.IsDir() { out = append(out, e.Name()) } }
	return out
}
func analyzeCrash(versionId, filename string) string {
	path := filepath.Join(getInstalledDir(versionId), "crash-reports", filename)
	b, err := os.ReadFile(path)
	if err != nil { return "" }
	lines := strings.Split(string(b), "\n")
	var result []string
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "exception") || strings.Contains(lower, "error") ||
			strings.Contains(lower, "fatal") || strings.Contains(lower, "caused by") {
			result = append(result, line)
		}
	}
	if len(result) == 0 { return "" }
	return strings.Join(result[:min(len(result), 10)], "\n")
}
func min(a, b int) int { if a < b { return a }; return b }

// ---  ---
func verifyVersion(versionId string) (bool, string) {
	dir := getInstalledDir(versionId)
	if !fileExists(dir) {
		return false, "版本目录不存在: " + versionId
	}
	jar := filepath.Join(dir, versionId+".jar")
	if !fileExists(jar) {
		return false, "缺少客户端 jar: " + versionId + ".jar"
	}
	if size, err := fileSize(jar); err != nil || size < 1024 {
		return false, "客户端 jar 不完整"
	}
	manifest := filepath.Join(dir, versionId+".json")
	if !fileExists(manifest) {
		return false, "缺少版本 manifest"
	}
	body, err := readAll(manifest)
	if err != nil || len(body) < 64 {
		return false, "版本 manifest 为空或损坏"
	}
	return true, ""
}
