//go:build !cshared

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/NekoMymaomao/mc-launcher-webui/core"
)

// 全局变量
var (
	baseDir  string
	cfg      *core.Settings
	accounts *core.AccountStore
	curLang  string
)

func init() {
	curLang = "en"
	if v := os.Getenv("MC_LAUNCHER_LANG"); v == "zh" || v == "zh-CN" {
		curLang = "zh"
	}
	baseDir = getLauncherDir()
	cfg = core.LoadSettings()
	accounts = core.LoadAccounts()
}

func getLauncherDir() string {
	if v := os.Getenv("MC_LAUNCHER_DIR"); v != "" {
		return v
	}
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)
	
	candidates := []string{
		filepath.Join(execDir, ".minecraft"),
		filepath.Join(execDir, "mc-launcher-data"),
		filepath.Join(os.Getenv("HOME"), ".mc-launcher"),
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}
	return filepath.Join(os.Getenv("HOME"), ".mc-launcher")
}

func setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	
	mux.HandleFunc("/api/ping", handlePing)
	mux.HandleFunc("/api/info", handleInfo)
	mux.HandleFunc("/api/versions", handleVersions)
	mux.HandleFunc("/api/versions/install", handleInstallVersion)
	mux.HandleFunc("/api/versions/remove", handleRemoveVersion)
	mux.HandleFunc("/api/versions/verify", handleVerifyVersion)
	mux.HandleFunc("/api/launch", handleLaunch)
	mux.HandleFunc("/api/accounts", handleAccounts)
	mux.HandleFunc("/api/accounts/offline", handleOfflineAuth)
	mux.HandleFunc("/api/accounts/microsoft", handleMicrosoftAuth)
	mux.HandleFunc("/api/java", handleJava)
	mux.HandleFunc("/api/java/install", handleInstallJava)
	mux.HandleFunc("/api/java/set", handleSetJava)
	mux.HandleFunc("/api/modrinth/search", handleModrinthSearch)
	mux.HandleFunc("/api/modrinth/info/", handleModrinthInfo)
	mux.HandleFunc("/api/modrinth/install", handleModrinthInstall)
	mux.HandleFunc("/api/config", handleConfig)
	mux.HandleFunc("/api/servers", handleServers)
	mux.HandleFunc("/api/crashes/", handleCrashes)
	mux.HandleFunc("/api/docs/", handleDocs)
	mux.HandleFunc("/docs/", serveStaticDocs)
	
	return mux
}

func main() {
	addr := ":9876"
	if v := os.Getenv("MC_LAUNCHER_ADDR"); v != "" {
		addr = v
	}
	
	os.Setenv("MC_LAUNCHER_DIR", baseDir)
	
	mux := setupRoutes()
	
	fmt.Printf("MC Launcher WebUI Server\n")
	fmt.Printf("Address: %s\n", addr)
	fmt.Printf("Base Dir: %s\n", baseDir)
	fmt.Printf("Language: %s\n", curLang)
	fmt.Println("================================")
	
	log.Fatal(http.ListenAndServe(addr, mux))
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeResponse(w http.ResponseWriter, resp core.Response) {
	status := http.StatusOK
	if !resp.Ok {
		status = http.StatusBadRequest
	}
	writeJSON(w, status, resp)
}

func okResp(data any) core.Response {
	return core.Response{Ok: true, Data: data}
}

func errResp(msg string) core.Response {
	return core.Response{Ok: false, Error: msg}
}

// ============ API Handlers ============

func handlePing(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, okResp(map[string]string{"status": "ok", "version": "2.0.0"}))
}

func handleInfo(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, okResp(map[string]any{
		"version":  "2.0.0",
		"baseDir":  baseDir,
		"language": curLang,
		"accounts": len(accounts.Accounts),
	}))
}

func handleVersions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		writeResponse(w, okResp(core.ListVersions()))
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
	
	resp, err := core.InstallVersion(req.Version, "", func(p core.Progress) {})
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
	
	err := core.RemoveVersion(req.Version)
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
	
	ok, msg := core.VerifyVersion(req.Version)
	writeResponse(w, okResp(map[string]any{"valid": ok, "message": msg}))
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
	
	resp, err := core.LaunchVersion(req.Version, req.Username)
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
		accounts.Accounts = nil
		accounts.Save()
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
	
	acc := core.LoginOffline(req.Name)
	accounts.Upsert(*acc)
	writeResponse(w, okResp(acc))
}

func handleMicrosoftAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeResponse(w, errResp("method not allowed"))
		return
	}
	
	acc, err := core.LoginMicrosoft(false, "")
	if err != nil {
		writeResponse(w, errResp(err.Error()))
		return
	}
	accounts.Upsert(*acc)
	writeResponse(w, okResp(acc))
}

func handleJava(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		writeResponse(w, okResp(core.ScanJava()))
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
	
	path, err := core.InstallJava(req.Major)
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
	
	cfg.JavaPath = req.Path
	cfg.Save()
	writeResponse(w, okResp(nil))
}

func handleModrinthSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	projectType := r.URL.Query().Get("type")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	
	resp := core.ModrinthSearch(query, projectType, page)
	writeResponse(w, resp)
}

func handleModrinthInfo(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/api/modrinth/info/")
	if slug == "" {
		writeResponse(w, errResp("slug required"))
		return
	}
	
	resp := core.ModrinthProject(slug)
	writeResponse(w, resp)
}

func handleModrinthInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeResponse(w, errResp("method not allowed"))
		return
	}
	
	var req struct {
		Slug      string `json:"slug"`
		MCVersion string `json:"mc_version"`
		Type      string `json:"type"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	
	if req.Slug == "" {
		writeResponse(w, errResp("slug required"))
		return
	}
	
	var resp core.Response
	switch req.Type {
	case "shader":
		resp = core.ModrinthShaderInstall(req.Slug, req.MCVersion)
	case "resourcepack":
		resp = core.ModrinthResourcePackInstall(req.Slug, req.MCVersion)
	case "datapack":
		resp = core.ModrinthDataPackInstall(req.Slug, req.MCVersion)
	case "modpack":
		resp = core.ModrinthPackInstall(req.Slug, req.MCVersion)
	default:
		resp = core.ModrinthContentInstall(req.Slug, req.MCVersion, "mod")
	}
	
	writeResponse(w, resp)
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		writeResponse(w, okResp(cfg))
	case "PUT":
		var updated core.Settings
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
		cfg.Save()
		writeResponse(w, okResp(cfg))
	default:
		writeResponse(w, errResp("method not allowed"))
	}
}

func handleServers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		writeResponse(w, okResp(core.ListServers()))
	case "POST":
		var req struct {
			Address string `json:"address"`
			Name    string `json:"name"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		core.AddServer(req.Address, req.Name)
		writeResponse(w, okResp(core.ListServers()))
	case "DELETE":
		addr := strings.TrimPrefix(r.URL.Path, "/api/servers/")
		core.RemoveServer(addr)
		writeResponse(w, okResp(core.ListServers()))
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
	
	crashes := core.ListCrashes(version)
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
