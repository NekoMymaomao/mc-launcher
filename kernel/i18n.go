//go:build !cshared



// i18n.go — 国际化支持 (-lang en/zh)
//go:build !cshared

import (
	"fmt"
	"sync"
)

var (
	i18nMu    sync.RWMutex
	i18nData  = make(map[string]map[string]string)
	i18nLoaded bool
)

func init() {
	// 预加载默认翻译
	ensureI18n()
}

func ensureI18n() {
	i18nMu.Lock()
	defer i18nMu.Unlock()
	if i18nLoaded {
		return
	}
	// 英文
	i18nData["en"] = map[string]string{
		// 通用
		"unknown_cmd":  "unknown command",
		"json_parse_error": "JSON parse error",
		"server_started": "Server started on %s",
		"server_error": "Server error: %v",

		// Java
		"java_not_found": "Java not found, please install Java 17+",
		"java_install":   "Installing Java %d...",
		"java_installed": "Java %d installed at %s",

		// 版本
		"version_missing": "Version not found",
		"version_exists":  "Version already exists",

		// Modrinth
		"modrinth_api_error": "Modrinth API error: %v",
		"modrinth_not_found": "Project not found: %s",
		"modrinth_parse_err": "Failed to parse Modrinth response",
		"modrinth_download":  "Downloading %s...",
		"modrinth_extract":   "Extracting %s...",
		"modrinth_done":      "Successfully installed %s",

		// 地图
		"map_install": "Installing map...",
		"map_done":    "Map installed",

		// 光影
		"shader_install": "Installing shader pack...",
		"shader_done":    "Shader pack installed",
	}

	// 中文
	i18nData["zh"] = map[string]string{
		// 通用
		"unknown_cmd":  "未知命令",
		"json_parse_error": "JSON 解析错误",
		"server_started": "服务器已启动于 %s",
		"server_error": "服务器错误: %v",

		// Java
		"java_not_found": "未找到 Java，请安装 Java 17+",
		"java_install":   "正在安装 Java %d...",
		"java_installed": "Java %d 已安装于 %s",

		// 版本
		"version_missing": "版本不存在",
		"version_exists":  "版本已存在",

		// Modrinth
		"modrinth_api_error": "Modrinth API 错误: %v",
		"modrinth_not_found": "项目不存在: %s",
		"modrinth_parse_err": "解析 Modrinth 响应失败",
		"modrinth_download":  "正在下载 %s...",
		"modrinth_extract":   "正在解压 %s...",
		"modrinth_done":      "安装成功: %s",

		// 地图
		"map_install": "正在安装地图...",
		"map_done":    "地图已安装",

		// 光影
		"shader_install": "正在安装光影包...",
		"shader_done":    "光影包已安装",
	}

	i18nLoaded = true
}

// i18n 获取翻译文本
func i18n(key string, args ...any) string {
	ensureI18n()
	i18nMu.RLock()
	defer i18nMu.RUnlock()

 langs := []string{curLang, "en"}
	for _, lang := range langs {
		if trans, ok := i18nData[lang][key]; ok {
			if len(args) > 0 {
				return fmt.Sprintf(trans, args...)
			}
			return trans
		}
	}
	return key // 返回 key 本身作为 fallback
}
