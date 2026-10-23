package main

// version.go — 版本清单解析
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	versionManifestURL = "https://launchermeta.mojang.com/mc/game/version_manifest_v2.json"
	mavenBaseURL       = "https://libraries.minecraft.net/"
)

func fetchManifest(source string) (*VersionManifest, error) {
	body, code, err := httpGet(mirrorUrl(versionManifestURL, source))
	if err != nil {
		return nil, fmt.Errorf("获取版本清单失败: %w", err)
	}
	if code != 200 {
		return nil, fmt.Errorf("获取版本清单失败: HTTP %d", code)
	}
	var m VersionManifest
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, fmt.Errorf("解析版本清单失败: %w", err)
	}
	if len(m.Versions) == 0 {
		return nil, fmt.Errorf("版本清单为空")
	}
	return &m, nil
}

// fetchVersionDetail 通过版本清单拿到该版本的元数据 URL 再下载。
// 之前直接拼 /mc/game/<id>.json（该地址不存在），安装/启动必然失败。
func fetchVersionDetail(id, source string) (*VersionDetail, error) {
	if id == "" {
		return nil, fmt.Errorf("版本号为空")
	}
	if m, err := fetchManifest(source); err == nil {
		for _, v := range m.Versions {
			if v.ID != id || v.URL == "" {
				continue
			}
			body, code, err := httpGet(mirrorUrl(v.URL, source))
			if err != nil {
				return nil, fmt.Errorf("获取版本 %s 元数据失败: %w", id, err)
			}
			if code != 200 {
				return nil, fmt.Errorf("获取版本 %s 元数据失败: HTTP %d", id, code)
			}
			return parseVersionJSON(string(body))
		}
	}
	// 清单里没有：回退到本地已安装的 json（离线启动 / 特殊版本）
	if local, err := readLocalVersionJSON(id); err == nil {
		return parseVersionJSON(local)
	}
	return nil, fmt.Errorf("版本 %s 不在 Mojang 清单中，且本地未安装", id)
}

func readLocalVersionJSON(id string) (string, error) {
	for _, dir := range []string{
		filepath.Join(baseDir, "versions", id),
		filepath.Join(baseDir, ".minecraft", "versions", id),
	} {
		b, err := os.ReadFile(filepath.Join(dir, id+".json"))
		if err == nil && len(b) > 0 {
			return string(b), nil
		}
	}
	return "", fmt.Errorf("本地没有版本 %s", id)
}

// parseVersionJSON 解析官方版本 JSON（可离线复用）
func parseVersionJSON(raw string) (*VersionDetail, error) {
	var v VersionDetail
	var j map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &j); err != nil {
		return nil, fmt.Errorf("解析版本 JSON 失败: %w", err)
	}
	v.RawJSON = raw

	v.ID = stringValue(j, "id")
	v.Type = stringValue(j, "type")
	v.MainClass = stringValue(j, "mainClass")
	v.Assets = stringValue(j, "assets")
	if v.Assets == "" {
		v.Assets = v.ID
	}

	if r, ok := j["assetIndex"]; ok {
		var ai struct {
			ID   string `json:"id"`
			URL  string `json:"url"`
			SHA1 string `json:"sha1"`
			Size int64  `json:"size"`
		}
		json.Unmarshal(r, &ai)
		v.AssetIndex = Artifact{URL: ai.URL, SHA1: ai.SHA1, Size: ai.Size}
		v.AssetIndexID = ai.ID
		if v.AssetIndexID == "" {
			v.AssetIndexID = v.Assets
		}
	}
	if r, ok := j["downloads"]; ok {
		var dl struct {
			Client struct {
				URL  string `json:"url"`
				SHA1 string `json:"sha1"`
				Size int64  `json:"size"`
				Path string `json:"path"`
			} `json:"client"`
		}
		json.Unmarshal(r, &dl)
		v.Client = Artifact{
			URL:  dl.Client.URL,
			SHA1: dl.Client.SHA1,
			Size: dl.Client.Size,
			Path: dl.Client.Path,
		}
	}
	if r, ok := j["javaVersion"]; ok {
		var jv struct {
			MajorVersion int `json:"majorVersion"`
		}
		json.Unmarshal(r, &jv)
		v.JavaMajor = jv.MajorVersion
	}
	if v.JavaMajor <= 0 {
		v.JavaMajor = majorFromID(v.ID)
	}
	if r, ok := j["minecraftArguments"]; ok {
		json.Unmarshal(r, &v.LegacyArgs)
	}
	if r, ok := j["logging"]; ok {
		var lg struct {
			Client struct {
				Argument string `json:"argument"`
				File     struct {
					ID   string `json:"id"`
					URL  string `json:"url"`
					SHA1 string `json:"sha1"`
					Size int64  `json:"size"`
				} `json:"file"`
			} `json:"client"`
		}
		json.Unmarshal(r, &lg)
		v.LogArgument = lg.Client.Argument
		v.LogFile = Artifact{
			Path: lg.Client.File.ID,
			URL:  lg.Client.File.URL,
			SHA1: lg.Client.File.SHA1,
			Size: lg.Client.File.Size,
		}
	}

	// ---- libraries ----
	plat := platformOS()
	if r, ok := j["libraries"]; ok {
		var libs []map[string]json.RawMessage
		json.Unmarshal(r, &libs)
		for _, libj := range libs {
			name := stringValue(libj, "name")
			if name == "" {
				continue
			}
			lib := Library{Name: name, MavenBase: mavenBaseURL}

			if r, ok := libj["rules"]; ok {
				var rules []Rule
				json.Unmarshal(r, &rules)
				lib.Rules = rules
			}

			if r, ok := libj["downloads"]; ok {
				var dl struct {
					Artifact    Artifact            `json:"artifact"`
					Classifiers map[string]Artifact `json:"classifiers"`
				}
				json.Unmarshal(r, &dl)
				lib.Downloads = &Downloads{Artifact: dl.Artifact, Classifiers: dl.Classifiers}

				if dl.Artifact.URL != "" {
					lib.ResolvedPath = dl.Artifact.Path
					if lib.ResolvedPath == "" {
						lib.ResolvedPath = mavenPath(name)
					}
					lib.ResolvedURL = mirrorUrl(dl.Artifact.URL, cfg.DownloadSource)
					lib.ResolvedSHA1 = dl.Artifact.SHA1
				}

				// natives：按平台挑 classifier
				key := nativeClassifier(libj, plat)
				if a, ok := dl.Classifiers[key]; ok && a.URL != "" {
					lib.NativeURL = mirrorUrl(a.URL, cfg.DownloadSource)
					lib.NativePath = a.Path
					if lib.NativePath == "" {
						lib.NativePath = mavenPathWithClassifier(name, key)
					}
					lib.NativeSHA1 = a.SHA1
					lib.IsNative = true
				}
			}

			if r, ok := libj["extract"]; ok {
				var ex struct {
					Exclude []string `json:"exclude"`
				}
				json.Unmarshal(r, &ex)
				lib.Extract = &Extract{Exclude: ex.Exclude}
			}

			lib.ShouldUse = ruleAllowed(lib.Rules)
			if lib.ShouldUse {
				v.Libraries = append(v.Libraries, lib)
			}
		}
	}

	// ---- arguments ----
	if r, ok := j["arguments"]; ok {
		var args struct {
			Game []json.RawMessage `json:"game"`
			JVM  []json.RawMessage `json:"jvm"`
		}
		json.Unmarshal(r, &args)
		parseArgs(args.Game, &v.GameArgs)
		parseArgs(args.JVM, &v.JVMArgs)
	}

	v.Valid = v.MainClass != ""
	if !v.Valid {
		return nil, fmt.Errorf("版本 JSON 缺少 mainClass")
	}
	return &v, nil
}

// nativeClassifier 读取 libraries[].natives 里当前平台对应的 classifier
func nativeClassifier(libj map[string]json.RawMessage, plat string) string {
	r, ok := libj["natives"]
	if !ok {
		return "natives-" + plat
	}
	var nat map[string]string
	json.Unmarshal(r, &nat)
	if c, ok := nat[plat]; ok && c != "" {
		return strings.ReplaceAll(c, "${arch}", archSuffix())
	}
	return "natives-" + plat
}

func archSuffix() string {
	if runtimeGOARCH == "386" || runtimeGOARCH == "arm" {
		return "32"
	}
	return "64"
}

// majorFromID 从版本号粗略推断需要的 Java 主版本
func majorFromID(id string) int {
	if strings.HasPrefix(id, "1.") {
		parts := strings.SplitN(id, ".", 3)
		if len(parts) >= 2 {
			if parseIntSuffix(parts[1]) >= 17 {
				return 17
			}
		}
		return 8
	}
	// 快照（24w14a）等：按较新的 Java 处理
	if y := parseIntSuffix(id); y >= 23 {
		return 21
	}
	return 17
}

func parseIntSuffix(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}

// parseArgs 解析 arguments.game / arguments.jvm，支持规则与字符串/数组两种 value
func parseArgs(raws []json.RawMessage, dst *[]string) {
	for _, raw := range raws {
		var s string
		if err := json.Unmarshal(raw, &s); err == nil {
			*dst = append(*dst, s)
			continue
		}
		var el struct {
			Rules []Rule          `json:"rules"`
			Value json.RawMessage `json:"value"`
		}
		if err := json.Unmarshal(raw, &el); err != nil {
			continue
		}
		if !ruleAllowed(el.Rules) {
			continue
		}
		var one string
		if err := json.Unmarshal(el.Value, &one); err == nil {
			*dst = append(*dst, one)
			continue
		}
		var many []string
		if err := json.Unmarshal(el.Value, &many); err == nil {
			*dst = append(*dst, many...)
		}
	}
}

func mavenPath(name string) string {
	parts := strings.SplitN(name, ":", 4)
	if len(parts) < 3 {
		return name + ".jar"
	}
	group := strings.ReplaceAll(parts[0], ".", "/")
	art, ver := parts[1], parts[2]
	cls := ""
	if len(parts) > 3 && parts[3] != "" {
		cls = "-" + parts[3]
	}
	return group + "/" + art + "/" + ver + "/" + art + "-" + ver + cls + ".jar"
}

// mavenPathWithClassifier 显式指定 classifier 的 maven 路径
func mavenPathWithClassifier(name, classifier string) string {
	parts := strings.SplitN(name, ":", 4)
	if len(parts) < 3 {
		return name + ".jar"
	}
	group := strings.ReplaceAll(parts[0], ".", "/")
	art, ver := parts[1], parts[2]
	return group + "/" + art + "/" + ver + "/" + art + "-" + ver + "-" + classifier + ".jar"
}

func platformOS() string {
	switch runtimeGOOS {
	case "windows":
		return "windows"
	default:
		return "linux"
	}
}

func ruleAllowed(rules []Rule) bool {
	if len(rules) == 0 {
		return true
	}
	allow := false
	for _, r := range rules {
		if ruleMatches(r) {
			allow = r.Action == "allow"
		}
	}
	return allow
}

func ruleMatches(r Rule) bool {
	if r.OS != nil {
		if r.OS.Name != "" && r.OS.Name != platformOS() {
			return false
		}
		if r.OS.Arch != "" {
			want, have := r.OS.Arch, runtimeGOARCH
			switch want {
			case "x86":
				if have != "386" && have != "amd64" {
					return false
				}
			case "x86_64":
				if have != "amd64" {
					return false
				}
			case "arm64":
				if have != "arm64" {
					return false
				}
			}
		}
	}
	for k, want := range r.Features {
		switch k {
		case "has_custom_resolution":
			have := cfg != nil && cfg.Width > 0 && cfg.Height > 0
			if want != have {
				return false
			}
		default:
			// is_demo_user / quick play 等特性我们都不支持
			if want {
				return false
			}
		}
	}
	return true
}

func stringValue(m map[string]json.RawMessage, key string) string {
	b, ok := m[key]
	if !ok {
		return ""
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return ""
	}
	return s
}
