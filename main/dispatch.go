package main

// dispatch.go — 
import (
	"encoding/json"
	"net"
	"os"
	"strconv"
	"strings"
)

func dispatch(req Request) Response {
	logDebug(": %s", req.Cmd)
	switch req.Cmd {
	case "ping":
		return okResp(nil)
	case "info":
		return okResp(map[string]any{
			"launcherDir": baseDir,
			"version":     "2.0.0-dev",
			"os":          runtimeGOOS,
			"arch":        runtimeGOARCH,
			"accounts":    len(accounts.Accounts),
			"versions":    listVersions(),
		})
	case "settings.get":
		return okResp(cfg)
	case "settings.set":
		if v, ok := req.Args["key"].(string); ok {
			switch v {
			case "maxMemory":
				if n := toInt(req.Args["value"]); n > 0 { cfg.MaxMemory = n }
			case "minMemory":
				if n := toInt(req.Args["value"]); n > 0 { cfg.MinMemory = n }
			case "source":
				cfg.DownloadSource = req.Args["value"].(string)
			}
			cfg.save()
		}
		return okResp(cfg)
	case "versions.list":
		return okResp(listVersions())
	case "versions.available":
		m, err := fetchManifest(cfg.DownloadSource)
		if err != nil { return errResp(err.Error()) }
		return okResp(m)
	case "version.detail":
		id, _ := req.Args["id"].(string)
		d, err := fetchVersionDetail(id, cfg.DownloadSource)
		if err != nil { return errResp(err.Error()) }
		return okResp(d)
	case "install":
		id, _ := req.Args["id"].(string)
		cb, _ := req.Args["cb"].(string)
		resp, err := installVersion(id, cb, nil)
		if err != nil { return errResp(err.Error()) }
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
		if major == 0 { major = 17 }
		path, err := installJava(major)
		if err != nil { return errResp(err.Error()) }
		return okResp(map[string]any{"path": path, "major": major})
	case "java.pick":
		need := toInt(req.Args["need"])
		if need == 0 { need = 17 }
		j := pickJava(need)
		if j == nil { return errResp(" Java") }
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
		if err != nil { return errResp(err.Error()) }
		accounts.upsert(*acc)
		return okResp(acc)
	case "auth.authlib":
		server, _ := req.Args["server"].(string)
		user, _ := req.Args["user"].(string)
		pass, _ := req.Args["pass"].(string)
		acc, err := loginAuthlib(server, user, pass)
		if err != nil { return errResp(err.Error()) }
		accounts.upsert(*acc)
		return okResp(acc)
	case "account.list":
		return okResp(accounts.Accounts)
	case "account.remove":
		typ, _ := req.Args["type"].(string)
		accounts.remove(typ)
		return okResp(nil)
	case "mods.list":
		id, _ := req.Args["id"].(string)
		return okResp(listMods(id))
	case "mods.add":
		id, _ := req.Args["id"].(string)
		path, _ := req.Args["path"].(string)
		ok := addMod(id, path)
		return okResp(map[string]bool{"ok": ok})
	case "mods.remove":
		id, _ := req.Args["id"].(string)
		name, _ := req.Args["name"].(string)
		ok := removeMod(id, name)
		return okResp(map[string]bool{"ok": ok})
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
		return errResp(": " + req.Cmd)
	}
}

func okResp(data any) Response { return Response{Ok: true, Data: data} }
func errResp(msg string) Response { return Response{Ok: false, Error: msg} }

func toInt(v any) int {
	switch x := v.(type) {
	case float64: return int(x)
	case int: return x
	case string:
		n, _ := strconv.Atoi(x)
		return n
	}
	return 0
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 65536)
	for {
		n, err := conn.Read(buf)
		if err != nil { return }
		var req Request
		if err := json.Unmarshal(buf[:n], &req); err != nil {
			writeResp(conn, Response{Ok: false, Error: "JSON "})
			continue
		}
		resp := dispatch(req)
		data, _ := json.Marshal(resp)
		data = append(data, '\n')
		conn.Write(data)
	}
}

func writeResp(conn net.Conn, resp Response) {
	data, _ := json.Marshal(resp)
	data = append(data, '\n')
	conn.Write(data)
}
