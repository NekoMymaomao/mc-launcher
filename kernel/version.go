package main

// version.go — 
import (
	"encoding/json"
	"strings"
)

func fetchManifest(source string) (*VersionManifest, error) {
	body, code, err := httpGet(mirrorUrl("https://launchermeta.mojang.com/mc/game/version_manifest_v2.json", source))
	if err != nil || code != 200 { return nil, err }
	var m VersionManifest
	if err := json.Unmarshal(body, &m); err != nil { return nil, err }
	return &m, nil
}

func fetchVersionDetail(id, source string) (*VersionDetail, error) {
	body, code, err := httpGet(mirrorUrl("https://launchermeta.mojang.com/mc/game/"+id+".json", source))
	if err != nil || code != 200 { return nil, err }
	return parseVersionJSON(string(body))
}

func parseVersionJSON(raw string) (*VersionDetail, error) {
	var v VersionDetail
	var j map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &j); err != nil { return nil, err }

	v.ID = stringValue(j, "id")
	v.Type = stringValue(j, "type")
	v.MainClass = stringValue(j, "mainClass")
	v.Assets = stringValue(j, "assets")
	if v.Assets == "" { v.Assets = v.ID }

	if raw, ok := j["assetIndex"]; ok {
		var ai struct{ ID, URL, SHA1 string; Size int64 }
		json.Unmarshal(raw, &ai)
		v.AssetIndex = Artifact{URL: ai.URL, SHA1: ai.SHA1, Size: ai.Size}
		v.AssetIndexID = ai.ID
		if v.Assets == "" { v.Assets = ai.ID }
	}
	if raw, ok := j["downloads"]; ok {
		var dl struct{ Client struct{ URL, SHA1 string; Size int64; Path string } }
		json.Unmarshal(raw, &dl)
		v.Client = Artifact{URL: dl.Client.URL, SHA1: dl.Client.SHA1, Size: dl.Client.Size}
	}
	if raw, ok := j["javaVersion"]; ok {
		var jv struct{ MajorVersion int }
		json.Unmarshal(raw, &jv)
		v.JavaMajor = jv.MajorVersion
		if v.JavaMajor <= 0 { v.JavaMajor = 8 }
	}
	if raw, ok := j["minecraftArguments"]; ok {
		json.Unmarshal(raw, &v.LegacyArgs)
	}
	if raw, ok := j["logging"]; ok {
		var lg struct{ Client struct{ Argument string; File struct{ URL, SHA1 string; Size int64; Path string } } }
		json.Unmarshal(raw, &lg)
		v.LogArgument = lg.Client.Argument
		v.LogFile = Artifact{URL: lg.Client.File.URL, SHA1: lg.Client.File.SHA1, Size: lg.Client.File.Size}
	}

	// libraries
	if raw, ok := j["libraries"]; ok {
		var libs []map[string]json.RawMessage
		json.Unmarshal(raw, &libs)
		plat := platformOS()
		for _, libj := range libs {
			name := stringValue(libj, "name")
			if name == "" { continue }

			lib := Library{Name: name, MavenBase: "https://libraries.minecraft.net/"}

			if raw, ok := libj["rules"]; ok {
				var rules []Rule
				json.Unmarshal(raw, &rules)
				lib.Rules = rules
			}
			if raw, ok := libj["downloads"]; ok {
				var dl struct {
					Artifact Artifact
					Classifiers map[string]Artifact
				}
				json.Unmarshal(raw, &dl)
				lib.Downloads = &Downloads{Artifact: dl.Artifact, Classifiers: dl.Classifiers}
				// resolve artifact path/url
				if dl.Artifact.URL != "" {
					lib.ResolvedURL = mirrorUrl(dl.Artifact.URL, cfg.DownloadSource)
					lib.ResolvedPath = dl.Artifact.Path
					if dl.Artifact.Path == "" {
						lib.ResolvedPath = mavenPath(name)
						lib.ResolvedURL = mirrorUrl(lib.MavenBase+lib.ResolvedPath, cfg.DownloadSource)
					}
				}
				// resolve natives
				nativesKey := "natives-" + plat
				if raw, ok := libj["natives"]; ok {
					var nat map[string]string
					json.Unmarshal(raw, &nat)
					lib.Natives = nat
				}
				if lib.Natives == nil && lib.Downloads != nil {
					lib.Natives = map[string]string{plat: nativesKey}
				}
				if lib.Downloads != nil && len(lib.Downloads.Classifiers) > 0 {
					// try platform-specific, then generic, then first
					try := []string{nativesKey, "natives-" + plat, "natives-linux", "natives-windows", "natives-osx"}
					for _, k := range try {
						if a, ok := lib.Downloads.Classifiers[k]; ok && a.URL != "" {
							lib.NativeURL = mirrorUrl(a.URL, cfg.DownloadSource)
							lib.NativePath = a.Path
							if lib.NativePath == "" { lib.NativePath = mavenPath(name + ":" + k) }
							lib.NativeURL = lib.MavenBase + lib.NativePath
							lib.IsNative = true
							break
						}
					}
					// fallback: try any classifier containing the platform name
					if !lib.IsNative {
						for k, a := range lib.Downloads.Classifiers {
							if strings.Contains(k, plat) && a.URL != "" {
								lib.NativeURL = mirrorUrl(a.URL, cfg.DownloadSource)
								lib.NativePath = a.Path
								if lib.NativePath == "" { lib.NativePath = mavenPath(name + ":" + k) }
								lib.NativeURL = lib.MavenBase + lib.NativePath
								lib.IsNative = true
								break
							}
						}
					}
				}
			}
			if raw, ok := libj["extract"]; ok {
				var ex struct{ Exclude []string }
				json.Unmarshal(raw, &ex)
				lib.Extract = &Extract{Exclude: ex.Exclude}
			}
			lib.ShouldUse = ruleAllowed(lib.Rules)
			if lib.ShouldUse { v.Libraries = append(v.Libraries, lib) }
		}
	}

	// arguments (modern)
	if raw, ok := j["arguments"]; ok {
		var args struct{ Game, JVM []json.RawMessage }
		json.Unmarshal(raw, &args)
		parseArgs(args.Game, &v.GameArgs)
		parseArgs(args.JVM, &v.JVMArgs)
	}

	v.Valid = true
	return &v, nil
}

func parseArgs(raws []json.RawMessage, dst *[]string) {
	for _, raw := range raws {
		var s string
		if err := json.Unmarshal(raw, &s); err == nil { *dst = append(*dst, s); continue }
		var el struct{ Rules []Rule; Value json.RawMessage }
		if err := json.Unmarshal(raw, &el); err == nil && ruleAllowed(el.Rules) {
			var ss []string
			json.Unmarshal(el.Value, &ss)
			*dst = append(*dst, ss...)
		}
	}
}

func mavenPath(name string) string {
	parts := strings.SplitN(name, ":", 4)
	if len(parts) < 3 { return name + ".jar" }
	group := strings.ReplaceAll(parts[0], ".", "/")
	art := parts[1]; ver := parts[2]; cls := ""
	if len(parts) > 3 { cls = ":" + parts[3] }
	return group + "/" + art + "/" + ver + "/" + art + "-" + ver + cls + ".jar"
}

func platformOS() string {
	switch runtimeGOOS {
	case "windows": return "windows"
	case "darwin": return "osx"
	default: return "linux"
	}
}

func ruleAllowed(rules []Rule) bool {
	if len(rules) == 0 { return true }
	allow := false
	for _, r := range rules {
		if r.matches() { allow = (r.Action == "allow") }
	}
	return allow
}

func (r Rule) matches() bool {
	if r.OS != nil {
		if r.OS.Name != "" && r.OS.Name != platformOS() { return false }
		// arch rarely used, skip
	}
	for k, v := range r.Features {
		switch k {
		case "is_demo_user": // we are not demo
			if v { return false }
		}
	}
	return true
}

func stringValue(m map[string]json.RawMessage, key string) string {
	b, ok := m[key]
	if !ok { return "" }
	var s string
	json.Unmarshal(b, &s)
	return s
}
