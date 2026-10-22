//go:build !cshared

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// 全局变量
var (
	baseDir  string
	cfg      *Settings
	accounts *AccountStore
	curLang  string
)

func init() {
	curLang = "en"
	if v := os.Getenv("MC_LAUNCHER_LANG"); v == "zh" || v == "zh-CN" {
		curLang = "zh"
	}
	baseDir = getLauncherDir()
	cfg = loadSettings()
	accounts = loadAccounts()
}

func getLauncherDir() string {
	if v := os.Getenv("MC_LAUNCHER_DIR"); v != "" {
		return v
	}
	// 默认位置
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)
	
	// 尝试多个位置
	candidates := []string{
		filepath.Join(execDir, ".minecraft"),
		filepath.Join(execDir, "mc-launcher-data"),
		filepath.Join(os.Getenv("USERPROFILE")或 os.Getenv("HOME"), ".mc-launcher"),
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}
	return filepath.Join(os.Getenv("USERPROFILE")或 os.Getenv("HOME"), ".mc-launcher")
}

// ============ 路由注册 ============

func setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	
	// 健康检查
	mux.HandleFunc("/api/ping", handlePing)
	mux.HandleFunc("/api/info", handleInfo)
	
	// 版本管理
	mux.HandleFunc("/api/versions", handleVersions)
	mux.HandleFunc("/api/versions/install", handleInstallVersion)
	mux.HandleFunc("/api/versions/remove", handleRemoveVersion)
	mux.HandleFunc("/api/versions/verify", handleVerifyVersion)
	
	// 启动游戏
	mux.HandleFunc("/api/launch", handleLaunch)
	
	// 账号管理
	mux.HandleFunc("/api/accounts", handleAccounts)
	mux.HandleFunc("/api/accounts/offline", handleOfflineAuth)
	mux.HandleFunc("/api/accounts/microsoft", handleMicrosoftAuth)
	
	// Java 管理
	mux.HandleFunc("/api/java", handleJava)
	mux.HandleFunc("/api/java/install", handleInstallJava)
	mux.HandleFunc("/api/java/set", handleSetJava)
	
	// Modrinth
	mux.HandleFunc("/api/modrinth/search", handleModrinthSearch)
	mux.HandleFunc("/api/modrinth/info", handleModrinthInfo)
	mux.HandleFunc("/api/modrinth/install", handleModrinthInstall)
	
	// 配置
	mux.HandleFunc("/api/config", handleConfig)
	
	// 服务器
	mux.HandleFunc("/api/servers", handleServers)
	
	// 崩溃报告
	mux.HandleFunc("/api/crashes", handleCrashes)
	
	// 文档
	mux.HandleFunc("/api/docs/", handleDocs)
	mux.HandleFunc("/docs/", serveStaticDocs)
	
	return mux
}

// ============ 主要函数 ============

func main() {
	addr := ":9876"
	if v := os.Getenv("MC_LAUNCHER_ADDR"); v != "" {
		addr = v
	}
	
	// 设置基础目录
	os.Setenv("MC_LAUNCHER_DIR", baseDir)
	
	mux := setupRoutes()
	
	fmt.Printf("MC Launcher WebUI Server\n")
	fmt.Printf("Address: %s\n", addr)
	fmt.Printf("Base Dir: %s\n", baseDir)
	fmt.Printf("Language: %s\n", curLang)
	fmt.Println("================================")
	
	log.Fatal(http.ListenAndServe(addr, mux))
}

// ============ 响应助手 ============

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeResponse(w http.ResponseWriter, resp Response) {
	status := http.StatusOK
	if !resp.Ok {
		status = http.StatusBadRequest
	}
	writeJSON(w, status, resp)
}

func okResp(data any) Response {
	return Response{Ok: true, Data: data}
}

func errResp(msg string) Response {
	return Response{Ok: false, Error: msg}
}

// ============ API Handlers ============

func handlePing(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, okResp(map[string]string{"status": "ok", "version": "2.0.0"}))
}

func handleInfo(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, okResp(map[string]any{
		"version":   "2.0.0",
		"baseDir":   baseDir,
		"language":  curLang,
		"goos":      runtimeGOOS,
		"goarch":    runtimeGOARCH,
		"accounts":  len(accounts.Accounts),
		"versions":  listVersions(),
	}))
}

func handleVersions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		writeResponse(w, okResp(listVersions()))
	case "POST":
		// 重定向到 /api/versions/install
		http.Redirect(w, r, "/api/versions/install", http.StatusMethodNotAllowed)
	default:
		writeResponse(w, errResp("method not allowed"))
	}
}

func handleInstallVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeResponse(w, errResp("method not allowed"))
		return
	}
	
	var req struct {
		Version string `json:"version"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	
	if req.Version == "" {
		writeResponse(w, errResp("version required"))
		return
	}
	
	resp, err := installVersion(req.Version, "", func(p Progress) {})
	if err != nil {
		writeResponse(w, errResp(err.Error()))
		return
	}
	writeResponse(w, *resp)
}

func handleRemoveVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeResponse(w, errResp("method not allowed"))
		return
	}
	
	var req struct {
		Version string `json:"version"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	
	if req.Version == "" {
		writeResponse(w, errResp("version required"))
		return
	}
	
	err := removeVersion(req.Version)
	if err != nil {
		writeResponse(w, errResp(err.Error()))
		return
	}
	writeResponse(w, okResp(nil))
}

func handleVerifyVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeResponse(w, errResp("method not allowed"))
		return
	}
	
	var req struct {
		Version string `json:"version"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	
	if req.Version == "" {
		writeResponse(w, errResp("version required"))
		return
	}
	
	ok, msg := verifyVersion(req.Version)
	writeResponse(w, okResp(map[string]any{
		"valid": ok,
		"message": msg,
	}))
}

func handleLaunch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeResponse(w, errResp("method not allowed"))
		return
	}
	
	var req struct {
		Version  string `json:"version"`
		Username string `json:"username"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	
	if req.Version == "" {
		writeResponse(w, errResp("version required"))
		return
	}
	
	if req.Username == "" {
		req.Username = "Player"
	}
	
	resp, err := launchVersion(req.Version, req.Username)
	if err != nil {
		writeResponse(w, errResp(err.Error()))
		return
	}
	writeResponse(w, *resp)
}

func handleAccounts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		writeResponse(w, okResp(accounts.Accounts))
	case "DELETE":
		// 清除所有账号
		accounts.Accounts = nil
		saveAccounts()
		writeResponse(w, okResp(nil))
	default:
		writeResponse(w, errResp("method not allowed"))
	}
}

func handleOfflineAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeResponse(w, errResp("method not allowed"))
		return
	}
	
	var req struct {
		Name string `json:"name"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	
	if req.Name == "" {
		req.Name = "Player"
	}
	
	acc := loginOffline(req.Name)
	accounts.upsert(*acc)
	writeResponse(w, okResp(acc))
}

func handleMicrosoftAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeResponse(w, errResp("method not allowed"))
		return
	}
	
	acc, err := loginMicrosoft(false, "")
	if err != nil {
		writeResponse(w, errResp(err.Error()))
		return
	}
	accounts.upsert(*acc)
	writeResponse(w, okResp(acc))
}

func handleJava(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		writeResponse(w, okResp(scanJava()))
	default:
		writeResponse(w, errResp("method not allowed"))
	}
}

func handleInstallJava(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeResponse(w, errResp("method not allowed"))
		return
	}
	
	var req struct {
		Major int `json:"major"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	
	if req.Major == 0 {
		req.Major = 17
	}
	
	path, err := installJava(req.Major)
	if err != nil {
		writeResponse(w, errResp(err.Error()))
		return
	}
	writeResponse(w, okResp(map[string]string{"path": path}))
}

func handleSetJava(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		writeResponse(w, errResp("method not allowed"))
		return
	}
	
	var req struct {
		Path string `json:"path"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	
	if req.Path == "" {
		writeResponse(w, errResp("path required"))
		return
	}
	
	cfg.JavaPath = req.Path
	cfg.save()
	writeResponse(w, okResp(nil))
}

func handleModrinthSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	projectType := r.URL.Query().Get("type")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	
	resp := modrinthSearch(query, projectType, page)
	writeResponse(w, resp)
}

func handleModrinthInfo(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/api/modrinth/info/")
	if slug == "" {
		writeResponse(w, errResp("slug required"))
		return
	}
	
	resp := modrinthProject(slug)
	writeResponse(w, resp)
}

func handleModrinthInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeResponse(w, errResp("method not allowed"))
		return
	}
	
	var req struct {
		Slug         string `json:"slug"`
		MCVersion    string `json:"mc_version"`
		Type         string `json:"type"` // mod, modpack, shader, resourcepack, datapack
	}
	json.NewDecoder(r.Body).Decode(&req)
	
	if req.Slug == "" {
		writeResponse(w, errResp("slug required"))
		return
	}
	
	var resp Response
	switch req.Type {
	case "shader":
		resp = modrinthShaderInstall(req.Slug, req.MCVersion)
	case "resourcepack":
		resp = modrinthResourcePackInstall(req.Slug, req.MCVersion)
	case "datapack":
		resp = modrinthDataPackInstall(req.Slug, req.MCVersion)
	case "modpack":
		resp = modrinthPackInstall(req.Slug, req.MCVersion)
	default:
		resp = modrinthContentInstall(req.Slug, req.MCVersion, "mod")
	}
	
	writeResponse(w, resp)
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		writeResponse(w, okResp(cfg))
	case "PUT":
		var updated Settings
		json.NewDecoder(r.Body).Decode(&updated)
		if updated.MaxMemory > 0 {
			cfg.MaxMemory = updated.MaxMemory
		}
		if updated.MinMemory > 0 {
			cfg.MinMemory = updated.MinMemory
		}
		if updated.Threads > 0 {
			cfg.Threads = updated.Threads
		}
		if updated.Width > 0 {
			cfg.Width = updated.Width
		}
		if updated.Height > 0 {
			cfg.Height = updated.Height
		}
		cfg.save()
		writeResponse(w, okResp(cfg))
	default:
		writeResponse(w, errResp("method not allowed"))
	}
}

func handleServers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		writeResponse(w, okResp(listServers()))
	case "POST":
		var req struct {
			Address string `json:"address"`
			Name    string `json:"name"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		addServer(req.Address, req.Name)
		writeResponse(w, okResp(listServers()))
	case "DELETE":
		addr := strings.TrimPrefix(r.URL.Path, "/api/servers/")
		removeServer(addr)
		writeResponse(w, okResp(listServers()))
	default:
		writeResponse(w, errResp("method not allowed"))
	}
}

func handleCrashes(w http.ResponseWriter, r *http.Request) {
	version := strings.TrimPrefix(r.URL.Path, "/api/crashes/")
	if version == "" {
		writeResponse(w, errResp("version required"))
		return
	}
	
	crashes := listCrashes(version)
	writeResponse(w, okResp(crashes))
}

func handleDocs(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/docs/")
	switch path {
	case "api":
		serveFile(w, r, "frontend/docs/api.md")
	case "tutorial":
		serveFile(w, r, "frontend/docs/tutorial.md")
	case "faq":
		serveFile(w, r, "frontend/docs/faq.md")
	default:
		writeResponse(w, errResp("not found"))
	}
}

func serveStaticDocs(w http.ResponseWriter, r *http.Request) {
	path := "frontend" + r.URL.Path
	if strings.HasSuffix(r.URL.Path, "/") {
		path += "index.html"
	}
	serveFile(w, r, path)
}

func serveFile(w http.ResponseWriter, r *http.Request, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	
	ext := filepath.Ext(path)
	switch ext {
	case ".html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case ".css":
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	case ".md":
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	
	w.Write(data)
}
