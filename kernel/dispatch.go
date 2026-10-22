//go:build !cshared



// dispatch.go — API 命令分发
//go:build !cshared

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func dispatch(req Request) Response {
	logDebug("[%s] cmd=%s", curLang, req.Cmd)
	switch req.Cmd {
	case "ping":
		return okResp(map[string]any{
			"ok":      true,
			"lang":    curLang,
			"version": "2.0.0-dev",
		})

	case "info":
		return okResp(map[string]any{
			"launcherDir": baseDir,
			"version":     "2.0.0-dev",
			"lang":        curLang,
			"os":          runtimeGOOS,
			"arch":        runtimeGOARCH,
			"accounts":    len(accounts.Accounts),
			"versions":    listVersions(),
			"api": map[string]any{
				"modrinth": "enabled",
				"curseforge": "planned",
			},
		})

	case "settings.get":
		return okResp(cfg)

	case "settings.set":
		if v, ok := req.Args["key"].(string); ok {
			switch v {
			case "maxMemory":
				if n := toInt(req.Args["value"]); n > 0 {
					cfg.MaxMemory = n
				}
			case "minMemory":
				if n := toInt(req.Args["value"]); n > 0 {
					cfg.MinMemory = n
				}
			case "source":
				cfg.DownloadSource = req.Args["value"].(string)
			case "lang":
				if s, ok := req.Args["value"].(string); ok {
					if s == "zh" || s == "zh-CN" {
						curLang = "zh"
					} else {
						curLang = "en"
					}
				}
			}
			cfg.save()
		}
		return okResp(cfg)

	case "lang.get":
		return okResp(map[string]any{"lang": curLang})

	case "lang.set":
		if v, ok := req.Args["lang"].(string); ok {
			if v == "zh" || v == "zh-CN" {
				curLang = "zh"
			} else {
				curLang = "en"
			}
		}
		return okResp(map[string]any{"lang": curLang})

	case "lang.list":
		return okResp(map[string]any{
			"supported": []map[string]any{
				{"code": "en", "name": "English"},
				{"code": "zh", "name": "中文"},
			},
			"current": curLang,
		})

	case "versions.list":
		return okResp(listVersions())

	case "versions.available":
		m, err := fetchManifest(cfg.DownloadSource)
		if err != nil {
			return errResp(err.Error())
		}
		return okResp(m)

	case "version.detail":
		id, _ := req.Args["id"].(string)
		d, err := fetchVersionDetail(id, cfg.DownloadSource)
		if err != nil {
			return errResp(err.Error())
		}
		return okResp(d)

	case "install":
		id, _ := req.Args["id"].(string)
		cb, _ := req.Args["cb"].(string)
		resp, err := installVersion(id, cb, nil)
		if err != nil {
			return errResp(err.Error())
		}
		return *resp

	case "launch":
		id, _ := req.Args["id"].(string)
		user, _ := req.Args["username"].(string)
		return launchVersion(id, user)

	case "remove":
		id, _ := req.Args["id"].(string)
		removeVersion(id)
		return okResp(nil)

	case "java.list":
		return okResp(scanJava())

	case "java.install":
		major := toInt(req.Args["version"])
		if major == 0 {
			major = 17
		}
		path, err := installJava(major)
		if err != nil {
			return errResp(err.Error())
		}
		return okResp(map[string]any{"path": path, "major": major})

	case "java.pick":
		need := toInt(req.Args["need"])
		if need == 0 {
			need = 17
		}
		j := pickJava(need)
		if j == nil {
			return errResp(i18n("java_not_found"))
		}
		return okResp(j)

	case "auth.offline":
		name, _ := req.Args["name"].(string)
		acc := loginOffline(name)
		accounts.upsert(*acc)
		return okResp(acc)

	case "auth.microsoft":
		verbose, _ := req.Args["verbose"].(bool)
		clientId, _ := req.Args["clientId"].(string)
		acc, err := loginMicrosoft(verbose, clientId)
		if err != nil {
			return errResp(err.Error())
		}
		accounts.upsert(*acc)
		return okResp(acc)

	case "auth.authlib":
		server, _ := req.Args["server"].(string)
		user, _ := req.Args["user"].(string)
		pass, _ := req.Args["pass"].(string)
		acc, err := loginAuthlib(server, user, pass)
		if err != nil {
			return errResp(err.Error())
		}
		accounts.upsert(*acc)
		return okResp(acc)

	case "account.list":
		return okResp(accounts.Accounts)

	case "account.remove":
		typ, _ := req.Args["type"].(string)
		accounts.remove(typ)
		return okResp(nil)

	// ===== Modrinth API 命令 =====
	case "modrinth.search":
		query, _ := req.Args["query"].(string)
		projectType, _ := req.Args["type"].(string) // "mod", "modpack", "resourcepack", "shader", "datapack"
		page, _ := strconv.Atoi(fmt.Sprintf("%v", req.Args["page"]))
		if page <= 0 {
			page = 1
		}
		return modrinthSearch(query, projectType, page)

	case "modrinth.project":
		slug, _ := req.Args["slug"].(string)
		return modrinthProject(slug)

	case "modrinth.files":
		slug, _ := req.Args["slug"].(string)
		projectType, _ := req.Args["type"].(string)
		return modrinthFiles(slug, projectType)

	case "modrinth.versions":
		slug, _ := req.Args["slug"].(string)
		gameVersion, _ := req.Args["game_version"].(string)
		return modrinthVersions(slug, gameVersion)

	case "modrinth.featured":
		projectType, _ := req.Args["type"].(string)
		return modrinthFeatured(projectType)

	case "modrinth.categories":
		return modrinthCategories()

	case "mod.list":
		id, _ := req.Args["id"].(string)
		return okResp(listMods(id))

	case "mod.add":
		id, _ := req.Args["id"].(string)
		path, _ := req.Args["path"].(string)
		ok := addMod(id, path)
		return okResp(map[string]bool{"ok": ok})

	case "mod.remove":
		id, _ := req.Args["id"].(string)
		name, _ := req.Args["name"].(string)
		ok := removeMod(id, name)
		return okResp(map[string]bool{"ok": ok})

	case "shader.list":
		return okResp(listShaders())

	case "shader.install":
		slug, _ := req.Args["slug"].(string)
		mcVersion, _ := req.Args["mc_version"].(string)
		return modrinthShaderInstall(slug, mcVersion)

	case "resourcepack.list":
		return okResp(listResourcePacks())

	case "resourcepack.install":
		slug, _ := req.Args["slug"].(string)
		mcVersion, _ := req.Args["mc_version"].(string)
		return modrinthResourcePackInstall(slug, mcVersion)

	case "datapack.list":
		return okResp(listDataPacks())

	case "datapack.install":
		slug, _ := req.Args["slug"].(string)
		mcVersion, _ := req.Args["mc_version"].(string)
		return modrinthDataPackInstall(slug, mcVersion)

	case "mod.pack.install":
		slug, _ := req.Args["slug"].(string)
		mcVersion, _ := req.Args["mc_version"].(string)
		return modrinthPackInstall(slug, mcVersion)

	case "mod.resourcepack.install":
		slug, _ := req.Args["slug"].(string)
		mcVersion, _ := req.Args["mc_version"].(string)
		return modrinthResourcePackInstall(slug, mcVersion)

	case "mod.shader.install":
		slug, _ := req.Args["slug"].(string)
		mcVersion, _ := req.Args["mc_version"].(string)
		return modrinthShaderInstall(slug, mcVersion)

	case "mod.datapack.install":
		slug, _ := req.Args["slug"].(string)
		mcVersion, _ := req.Args["mc_version"].(string)
		return modrinthDataPackInstall(slug, mcVersion)

	case "map.list":
		id, _ := req.Args["id"].(string)
		return okResp(listSaves(id))

	case "map.install":
		slug, _ := req.Args["slug"].(string)
		mcVersion, _ := req.Args["mc_version"].(string)
		targetDir, _ := req.Args["target"].(string)
		return modrinthMapInstall(slug, mcVersion, targetDir)

	case "saves.list":
		id, _ := req.Args["id"].(string)
		return okResp(listSaves(id))

	case "saves.backup":
		id, _ := req.Args["id"].(string)
		name, _ := req.Args["name"].(string)
		return okResp(map[string]bool{"ok": backupSave(id, name)})

	case "saves.delete":
		id, _ := req.Args["id"].(string)
		name, _ := req.Args["name"].(string)
		return okResp(map[string]bool{"ok": deleteSave(id, name)})

	case "servers.list":
		return okResp(listServers())

	case "servers.add":
		addr, _ := req.Args["addr"].(string)
		name, _ := req.Args["name"].(string)
		addServer(addr, name)
		return okResp(nil)

	case "servers.remove":
		addr, _ := req.Args["addr"].(string)
		removeServer(addr)
		return okResp(nil)

	case "crash.list":
		id, _ := req.Args["id"].(string)
		return okResp(listCrashes(id))

	case "crash.analyze":
		id, _ := req.Args["id"].(string)
		file, _ := req.Args["file"].(string)
		return okResp(map[string]string{"content": analyzeCrash(id, file)})

	case "verify":
		id, _ := req.Args["id"].(string)
		ok, msg := verifyVersion(id)
		return okResp(map[string]any{"ok": ok, "message": msg})

	default:
		return errResp(i18n("unknown_cmd") + ": " + req.Cmd)
	}
}

func okResp(data any) Response {
	return Response{Ok: true, Data: data}
}

func errResp(msg string) Response {
	return Response{Ok: false, Error: msg}
}

func toInt(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case string:
		n, _ := strconv.Atoi(x)
		return n
	}
	return 0
}
