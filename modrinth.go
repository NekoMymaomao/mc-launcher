//go:build !cshared

package main
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	modrinthAPIBase = "https://api.modrinth.com/v2"
)

// ===== Modrinth API 数据结构 =====

type ModrinthProject struct {
	Slug         string           `json:"slug"`
	Title        string           `json:"title"`
	Description  string           `json:"description"`
	Type         string           `json:"type"` // mod, modpack, resourcepack, shader, datapack
	ClientSide   string           `json:"client_side"`
	ServerSide   string           `json:"server_side"`
	LicensedUnder string         `json:"licensed_under"`
	Downloads    int64            `json:"downloads"`
	Followers    int64            `json:"followers"`
	VersionCount int              `json:"version_count"`
	Categories   []string         `json:"categories"`
	Team         string           `json:"team"`
	Monitor      bool             `json:"monitor"`
	Requested    bool             `json:"requested"`
	Featured     bool             `json:"featured"`
	License      string           `json:"license"`
	URL          string           `json:"url"`
	ImageURL     string           `json:"image_url"`
	DateCreated  string           `json:"date_created"`
	DateUpdated  string           `json:"date_updated"`
}

type ModrinthFile struct {
	FileName string   `json:"filename"`
	Size     int64    `json:"size"`
	HASH     string   `json:"hash"`
	Files    []struct {
		Hash       string `json:"hash"`
		EntryName  string `json:"entry_name"`
		Path       []string `json:"path"`
		 fileType  string `json:"file_type"`
	} `json:"files"`
	PrimaryFile *ModrinthFileDetail `json:"primary_file"`
	DateEdited  string              `json:"date_edited"`
	Requested   bool                `json:"requested"`
	Monitored   bool                `json:"monitored"`
}

type ModrinthFileDetail struct {
	Path       []string `json:"path"`
	Primary    bool     `json:"primary"`
}

type ModrinthVersion struct {
	ID           string         `json:"id"`
	ProjectID    string         `json:"project_id"`
	Version      string         `json:"version"`
	Name         string         `json:"name"`
	VersionType  string         `json:"version_type"`
	GameVersions []string       `json:"game_versions"`
	Dependencies []ModrinthDep  `json:"dependencies"`
	DatePosted   string         `json:"date_posted"`
	DownloadURL  string         `json:"download_url,omitempty"`
	File         *ModrinthFile  `json:"file"`
}

type ModrinthDep struct {
	ProjectID string `json:"project_id"`
	VersionID string `json:"version_id"`
	FileType  string `json:"file_type"`
	DependencyType string `json:"dependency_type"`
}

type ModrinthSearchResponse struct {
	Hits     []ModrinthProject `json:"hits"`
	TotalHits int64            `json:"total_hits"`
	Page     int               `json:"page"`
	Clarity  int               `json:"clarity"`
}

type ModrinthCategory struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

// ===== Modrinth API 函数 =====

func modrinthSearch(query, projectType string, page int) Response {
	url := fmt.Sprintf("%s/search?query=%s&limit=20&page=%d", modrinthAPIBase, query, page)
	if projectType != "" {
		url += "&type=" + projectType
	}
	body, code, err := httpGet(url)
	if err != nil {
		return errResp(fmt.Sprintf(i18n("modrinth_api_error"), err))
	}
	if code != 200 {
		return errResp(fmt.Sprintf(i18n("modrinth_api_error"), fmt.Errorf("HTTP %d", code)))
	}

	var resp ModrinthSearchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return errResp(i18n("modrinth_parse_err"))
	}

	// 转换为简单格式
	hits := make([]map[string]any, len(resp.Hits))
	for i, p := range resp.Hits {
		hits[i] = map[string]any{
			"slug":        p.Slug,
			"title":       p.Title,
			"description": p.Description,
			"type":        p.Type,
			"downloads":   p.Downloads,
			"categories":  p.Categories,
			"url":         p.URL,
		}
	}

	return okResp(map[string]any{
		"hits":      hits,
		"total":     resp.TotalHits,
		"page":      resp.Page,
		"clarity":   resp.Clarity,
	})
}

func modrinthProject(slug string) Response {
	if slug == "" {
		return errResp("slug required")
	}
	url := fmt.Sprintf("%s/project/%s", modrinthAPIBase, slug)
	body, code, err := httpGet(url)
	if err != nil {
		return errResp(fmt.Sprintf(i18n("modrinth_api_error"), err))
	}
	if code == 404 {
		return errResp(fmt.Sprintf(i18n("modrinth_not_found"), slug))
	}
	if code != 200 {
		return errResp(fmt.Sprintf(i18n("modrinth_api_error"), fmt.Errorf("HTTP %d", code)))
	}

	var proj ModrinthProject
	if err := json.Unmarshal(body, &proj); err != nil {
		return errResp(i18n("modrinth_parse_err"))
	}

	return okResp(map[string]any{
		"slug":         proj.Slug,
		"title":        proj.Title,
		"description":  proj.Description,
		"type":         proj.Type,
		"downloads":    proj.Downloads,
		"followers":    proj.Followers,
		"categories":   proj.Categories,
		"url":          proj.URL,
		"image_url":    proj.ImageURL,
		"date_created": proj.DateCreated,
		"date_updated": proj.DateUpdated,
	})
}

func modrinthFiles(slug, projectType string) Response {
	if slug == "" {
		return errResp("slug required")
	}
	url := fmt.Sprintf("%s/project/%s/version", modrinthAPIBase, slug)
	body, code, err := httpGet(url)
	if err != nil {
		return errResp(fmt.Sprintf(i18n("modrinth_api_error"), err))
	}
	if code != 200 {
		return errResp(fmt.Sprintf(i18n("modrinth_api_error"), fmt.Errorf("HTTP %d", code)))
	}

	var versions []ModrinthVersion
	if err := json.Unmarshal(body, &versions); err != nil {
		return errResp(i18n("modrinth_parse_err"))
	}

	files := make([]map[string]any, len(versions))
	for i, v := range versions {
		files[i] = map[string]any{
			"id":           v.ID,
			"version":      v.Version,
			"name":         v.Name,
			"version_type": v.VersionType,
			"game_versions": v.GameVersions,
			"date_posted":  v.DatePosted,
			"download_url": getPrimaryFileURL(v),
		}
	}

	return okResp(files)
}

func modrinthVersions(slug, gameVersion string) Response {
	if slug == "" {
		return errResp("slug required")
	}
	url := fmt.Sprintf("%s/project/%s/version", modrinthAPIBase, slug)
	if gameVersion != "" {
		url += "?game_version=" + gameVersion
	}
	body, code, err := httpGet(url)
	if err != nil {
		return errResp(fmt.Sprintf(i18n("modrinth_api_error"), err))
	}
	if code != 200 {
		return errResp(fmt.Sprintf(i18n("modrinth_api_error"), fmt.Errorf("HTTP %d", code)))
	}

	var versions []ModrinthVersion
	if err := json.Unmarshal(body, &versions); err != nil {
		return errResp(i18n("modrinth_parse_err"))
	}

	result := make([]map[string]any, len(versions))
	for i, v := range versions {
		result[i] = map[string]any{
			"id":            v.ID,
			"version":       v.Version,
			"name":          v.Name,
			"version_type":  v.VersionType,
			"game_versions": v.GameVersions,
			"dependencies":  v.Dependencies,
			"date_posted":   v.DatePosted,
			"download_url":  getPrimaryFileURL(v),
		}
	}

	return okResp(result)
}

func modrinthFeatured(projectType string) Response {
	url := fmt.Sprintf("%s/search?featured=true&limit=20", modrinthAPIBase)
	if projectType != "" {
		url += "&type=" + projectType
	}
	body, code, err := httpGet(url)
	if err != nil {
		return errResp(fmt.Sprintf(i18n("modrinth_api_error"), err))
	}
	if code != 200 {
		return errResp(fmt.Sprintf(i18n("modrinth_api_error"), fmt.Errorf("HTTP %d", code)))
	}

	var resp ModrinthSearchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return errResp(i18n("modrinth_parse_err"))
	}

	hits := make([]map[string]any, len(resp.Hits))
	for i, p := range resp.Hits {
		hits[i] = map[string]any{
			"slug":        p.Slug,
			"title":       p.Title,
			"description": p.Description,
			"type":        p.Type,
			"downloads":   p.Downloads,
			"categories":  p.Categories,
			"url":         p.URL,
		}
	}

	return okResp(map[string]any{
		"hits":    hits,
		"total":   resp.TotalHits,
		"featured": true,
	})
}

func modrinthCategories() Response {
	// Modrinth 分类从项目类型推断
	types := []string{"mod", "modpack", "resourcepack", "shader", "datapack"}
	categories := make([]map[string]any, len(types))
	for i, t := range types {
		categories[i] = map[string]any{
			"id":   t,
			"name": i18nCategoryName(t),
		}
	}
	return okResp(categories)
}

// ===== Modrinth 安装函数 =====

func modrinthPackInstall(slug, mcVersion string) Response {
	// 获取项目信息
	proj := modrinthProject(slug)
	if !proj.Ok {
		return errResp(proj.Error)
	}

	pdata, _ := proj.Data.(map[string]any)
	if pdata["type"] != "modpack" {
		return errResp("这不是一个整合包项目")
	}

	// 获取最新版本
	files := modrinthVersions(slug, mcVersion)
	if !files.Ok {
		return errResp(files.Error)
	}

	fileList, _ := files.Data.([]map[string]any)
	if len(fileList) == 0 {
		return errResp("没有找到兼容版本")
	}

	version := fileList[0]
	downloadURL, _ := version["download_url"].(string)
	if downloadURL == "" {
		return errResp("无法获取下载链接")
	}

	// 安装到本地 modrinth 缓存
	cacheDir := filepath.Join(baseDir, "mods", "modrinth", "cache")
	ensureDir(cacheDir)

	filename := fmt.Sprintf("%s-%s.zip", slug, version["version"])
	dest := filepath.Join(cacheDir, filename)

	logInfo(i18n("modrinth_download"), filename)
	if err := downloadFile(downloadURL, dest, "", nil); err != nil {
		return errResp(fmt.Sprintf("下载失败: %v", err))
	}

	// 解压并安装
	targetDir := filepath.Join(baseDir, "mods", "modrinth", "installed", slug)
	ensureDir(targetDir)

	logInfo(i18n("modrinth_extract"), filename)
	if err := unzip(dest, targetDir, nil); err != nil {
		return errResp(fmt.Sprintf("解压失败: %v", err))
	}

	// 复制 mods 目录
	srcMods := filepath.Join(targetDir, "mods")
	if fileExists(srcMods) {
		dstMods := filepath.Join(baseDir, "mods")
		// 递归复制
		filepath.Walk(srcMods, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(srcMods, path)
			destPath := filepath.Join(dstMods, rel)
			ensureDir(filepath.Dir(destPath))
			data, _ := os.ReadFile(path)
			os.WriteFile(destPath, data, 0644)
			return nil
		})
	}

	return okResp(map[string]any{
		"slug":       slug,
		"title":      pdata["title"],
		"version":    version["version"],
		"target_dir": targetDir,
		"installed":  true,
	})
}

func modrinthResourcePackInstall(slug, mcVersion string) Response {
	return modrinthContentInstall(slug, mcVersion, "resourcepack")
}

func modrinthShaderInstall(slug, mcVersion string) Response {
	return modrinthContentInstall(slug, mcVersion, "shader")
}

func modrinthDataPackInstall(slug, mcVersion string) Response {
	return modrinthContentInstall(slug, mcVersion, "datapack")
}

func modrinthMapInstall(slug, mcVersion, targetDir string) Response {
	if targetDir == "" {
		targetDir = filepath.Join(baseDir, "mods", "modrinth", "maps", slug)
	}
	return modrinthContentInstall(slug, mcVersion, "datapack", targetDir)
}

func modrinthContentInstall(slug, mcVersion, contentType string, extraTargets ...string) Response {
	// 获取项目信息
	proj := modrinthProject(slug)
	if !proj.Ok {
		return errResp(proj.Error)
	}

	pdata, _ := proj.Data.(map[string]any)
	actualType := pdata["type"].(string)
	if contentType == "shader" && actualType != "shader" {
		return errResp("这不是一个光影项目")
	}
	if contentType == "resourcepack" && actualType != "resourcepack" {
		return errResp("这不是一个资源包项目")
	}

	// 获取版本
	files := modrinthVersions(slug, mcVersion)
	if !files.Ok {
		return errResp(files.Error)
	}

	fileList, _ := files.Data.([]map[string]any)
	if len(fileList) == 0 {
		return errResp("没有找到兼容版本")
	}

	version := fileList[0]
	downloadURL, _ := version["download_url"].(string)
	if downloadURL == "" {
		return errResp("无法获取下载链接")
	}

	// 确定目标目录
	var targetDir string
	switch contentType {
	case "shader":
		targetDir = filepath.Join(baseDir, "shaderpacks")
	case "resourcepack":
		targetDir = filepath.Join(baseDir, "resourcepacks")
	case "datapack":
		if len(extraTargets) > 0 {
			targetDir = extraTargets[0]
		} else {
			targetDir = filepath.Join(baseDir, "mods", "modrinth", "datapacks", slug)
		}
	default:
		targetDir = filepath.Join(baseDir, "mods", "modrinth", slug)
	}

	ensureDir(targetDir)

	// 下载
	cacheDir := filepath.Join(baseDir, "mods", "modrinth", "cache")
	ensureDir(cacheDir)
	filename := fmt.Sprintf("%s-%s.zip", slug, version["version"])
	dest := filepath.Join(cacheDir, filename)

	logInfo(i18n("modrinth_download"), filename)
	if err := downloadFile(downloadURL, dest, "", nil); err != nil {
		return errResp(fmt.Sprintf("下载失败: %v", err))
	}

	// 解压
	logInfo(i18n("modrinth_extract"), filename)
	if err := unzip(dest, targetDir, nil); err != nil {
		return errResp(fmt.Sprintf("解压失败: %v", err))
	}

	// 清理缓存
	os.Remove(dest)

	return okResp(map[string]any{
		"slug":       slug,
		"title":      pdata["title"],
		"version":    version["version"],
		"target_dir": targetDir,
		"content_type": contentType,
		"installed":  true,
	})
}

// ===== 辅助函数 =====

func getPrimaryFileURL(v ModrinthVersion) string {
	if v.File != nil && len(v.File.Files) > 0 {
		return v.File.Files[0].Hash // 实际需要构建完整 URL
	}
	if v.DownloadURL != "" {
		return v.DownloadURL
	}
	return ""
}

func i18nCategoryName(t string) string {
	switch t {
	case "mod":
		return "Mod"
	case "modpack":
		return "Modpack"
	case "resourcepack":
		return "Resource Pack"
	case "shader":
		return "Shader"
	case "datapack":
		return "Data Pack"
	default:
		return t
	}
}
