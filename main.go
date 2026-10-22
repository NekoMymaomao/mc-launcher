package main

// main.go — 入口（普通构建）
//go:build !cshared

import "os"

func main() {
	baseDir = launcherDir()
	ensureDir(baseDir)
	initLogger()
	cfg = loadSettings()
	accounts = loadAccounts()

	for _, d := range []string{"versions", "libraries", "assets/indexes", "assets/objects",
		"java", "logs", "sessions", "mods"} {
		ensureDir(baseDir + "/" + d)
	}

	addr := cfg.ListenAddr
	if addr == "" { addr = "127.0.0.1:9876" }
	Serve(addr)
}
