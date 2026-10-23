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

// ModrinthFile 对应 /version 响应里的 files[i]
// 旧结构体把 files 当成单个 file，且没有 url 字段，
// 导致 getPrimaryFileURL 只能返回 hash，所有下载 404。
type ModrinthFile struct {
	Hashes   map[string]string `json:"hashes"`
	URL      string            `json:"url"`
	Filename string            `json:"filename"`
	Primary  bool              `json:"primary"`
	Size     int64             `json:"size"`
	FileType string            `json:"file_type,omitempty"`
}

type ModrinthVersion struct {
	ID           string         `json:"id"`
	ProjectID    string         `json:"project_id"`
	Version      string         `json:"version"`
	Name         string         `json:"name"`
	VersionType  string         `json:"version_type"`
	GameVersions []string       `json:"game_versions"`
	Loaders      []string       `json:"loaders"`
	Dependencies []ModrinthDep  `json:"dependencies"`
	DatePosted   string         `json:"date_posted"`
	Files        []ModrinthFile `json:"files"`
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
	proj := modrinthProject(slug)
	if !proj.Ok {
		return errResp(proj.Error)
	}
	pdata, _ := proj.Data.(map[string]any)
	if pdata["type"] != "modpack" {
		return errResp("这不是一个整合包项目")
	}

	v, err := pickModrinthVersion(slug, mcVersion, "modpack")
	if err != nil {
		return errResp(err.Error())
	}
	fileURL := getPrimaryFileURL(*v)
	if fileURL == "" {
		return errResp("无法获取下载链接")
	}

	cacheDir := modrinthCacheDir()
	if err := ensureDir(cacheDir); err != nil {
		return errResp("创建缓存目录失败: " + err.Error())
	}
	name := getPrimaryFileName(*v)
	if name == "" {
		name = slug + "-" + v.Version + ".mrpack"
	}
	archive := filepath.Join(cacheDir, name)

	logInfo(i18n("modrinth_download"), name)
	if err := downloadFile(fileURL, archive, "", nil); err != nil {
		return errResp("下载失败: " + err.Error())
	}

	targetDir := modrinthInstalledDir(slug)
	if err := ensureDir(targetDir); err != nil {
		return errResp("创建目录失败: " + err.Error())
	}

	// 整合包是 .mrpack：modrinth.index.json + overrides/
	installed, err := installMrpack(archive, targetDir)
	if err != nil {
		return errResp("安装整合包失败: " + err.Error())
	}
	os.Remove(archive)

	return okResp(map[string]any{
		"slug":       slug,
		"title":      pdata["title"],
		"version":    v.Version,
		"target_dir": targetDir,
		"files":      installed,
		"installed":  true,
	})
}

// pickModrinthVersion 选一个与 mcVersion/loader 兼容的版本（没有兼容则退回最新）
func pickModrinthVersion(slug, mcVersion, loader string) (*ModrinthVersion, error) {
	url := modrinthAPIBase + "/project/" + slug + "/version"
	body, code, err := httpGet(url)
	if err != nil {
		return nil, fmt.Errorf("获取版本列表失败: %w", err)
	}
	if code != 200 {
		return nil, fmt.Errorf("获取版本列表失败: HTTP %d", code)
	}
	var versions []ModrinthVersion
	if err := json.Unmarshal(body, &versions); err != nil {
		return nil, fmt.Errorf("解析版本列表失败: %w", err)
	}
	if len(versions) == 0 {
		return nil, fmt.Errorf("没有可用版本")
	}
	if mcVersion == "" && loader == "" {
		return &versions[0], nil
	}
	for i := range versions {
		v := &versions[i]
		if mcVersion != "" && !containsStr(v.GameVersions, mcVersion) {
			continue
		}
		if loader != "" && loader != "modpack" && !containsStr(v.Loaders, loader) {
			continue
		}
		return v, nil
	}
	// 没有精确匹配时退回最新，避免直接失败
	return &versions[0], nil
}

func containsStr(list []string, want string) bool {
	for _, s := range list {
		if strings.EqualFold(s, want) {
			return true
		}
	}
	return false
}

// mrpackIndex 是 modrinth.index.json 的结构（只取用得到的字段）
type mrpackIndex struct {
	FormatVersion int    `json:"formatVersion"`
	Name          string `json:"name"`
	VersionID     string `json:"versionId"`
	Files         []struct {
		Path      string            `json:"path"`
		Hashes    map[string]string `json:"hashes"`
		Downloads []string          `json:"downloads"`
		FileSize  int64             `json:"fileSize"`
	} `json:"files"`
}

// installMrpack 解包 .mrpack：下载 index 里列出的文件 + 解压 overrides/
// 返回成功写入的文件数。旧实现直接把整合包当普通 zip 解压，
// 但 mrpack 里根本没有 mods/ 目录，于是“安装成功”却什么都没装上。
func installMrpack(archive, targetDir string) (int, error) {
	indexRaw, err := readZipEntry(archive, "modrinth.index.json")
	if err != nil {
		return 0, fmt.Errorf("不是有效的整合包（缺少 modrinth.index.json）")
	}
	var idx mrpackIndex
	if err := json.Unmarshal([]byte(indexRaw), &idx); err != nil {
		return 0, fmt.Errorf("解析 modrinth.index.json 失败: %w", err)
	}
	if len(idx.Files) == 0 {
		return 0, fmt.Errorf("整合包里没有文件清单")
	}

	count := 0
	var failed []string
	for _, f := range idx.Files {
		rel := filepath.Clean(f.Path)
		if strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
			continue // 防路径穿越
		}
		if len(f.Downloads) == 0 {
			continue
		}
		dest := filepath.Join(targetDir, rel)
		sha := f.Hashes["sha1"]
		if err := downloadFile(f.Downloads[0], dest, sha, nil); err != nil {
			failed = append(failed, rel)
			continue
		}
		count++
	}

	// overrides/ 与 client-overrides/ 直接覆盖到实例目录
	for _, prefix := range []string{"overrides/", "client-overrides/"} {
		if err := unzipPrefix(archive, prefix, targetDir); err != nil {
			return count, err
		}
	}

	if len(failed) > 0 {
		return count, fmt.Errorf("%d 个文件下载失败，例如 %s", len(failed), failed[0])
	}
	return count, nil
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
	proj := modrinthProject(slug)
	if !proj.Ok {
		return errResp(proj.Error)
	}
	pdata, _ := proj.Data.(map[string]any)
	actualType, _ := pdata["type"].(string)
	switch contentType {
	case "shader":
		if actualType != "shader" {
			return errResp("这不是一个光影项目")
		}
	case "resourcepack":
		if actualType != "resourcepack" {
			return errResp("这不是一个资源包项目")
		}
	case "datapack":
		if actualType != "datapack" {
			return errResp("这不是一个数据包项目")
		}
	}

	v, err := pickModrinthVersion(slug, mcVersion, "")
	if err != nil {
		return errResp(err.Error())
	}
	fileURL := getPrimaryFileURL(*v)
	if fileURL == "" {
		return errResp("无法获取下载链接")
	}
	name := getPrimaryFileName(*v)
	if name == "" {
		name = slug + "-" + v.Version + ".zip"
	}

	// 目标目录：资源包/光影/数据包都是“丢一个文件进去”，不解压
	var targetDir string
	switch contentType {
	case "shader":
		targetDir = getShaderDir()
	case "resourcepack":
		targetDir = getResourcePackDir()
	case "datapack":
		if len(extraTargets) > 0 && extraTargets[0] != "" {
			targetDir = extraTargets[0]
		} else {
			targetDir = filepath.Join(baseDir, "mods", "modrinth", "datapacks", slug)
		}
	default:
		targetDir = filepath.Join(baseDir, "mods")
	}
	if err := ensureDir(targetDir); err != nil {
		return errResp("创建目录失败: " + err.Error())
	}

	dest := filepath.Join(targetDir, name)
	logInfo(i18n("modrinth_download"), name)
	if err := downloadFile(fileURL, dest, "", nil); err != nil {
		return errResp("下载失败: " + err.Error())
	}

	return okResp(map[string]any{
		"slug":         slug,
		"title":        pdata["title"],
		"version":      v.Version,
		"path":         dest,
		"content_type": contentType,
		"installed":    true,
	})
}

// ===== 辅助函数 =====

// getPrimaryFileURL 返回该版本主文件的真实下载地址
func getPrimaryFileURL(v ModrinthVersion) string {
	for _, f := range v.Files {
		if f.Primary && f.URL != "" {
			return f.URL
		}
	}
	for _, f := range v.Files {
		if f.URL != "" {
			return f.URL
		}
	}
	return ""
}

// getPrimaryFileName 返回主文件名（用于落盘命名）
func getPrimaryFileName(v ModrinthVersion) string {
	for _, f := range v.Files {
		if f.Primary && f.Filename != "" {
			return f.Filename
		}
	}
	if len(v.Files) > 0 {
		return v.Files[0].Filename
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
