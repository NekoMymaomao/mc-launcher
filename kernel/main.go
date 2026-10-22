//go:build !cshared



// main.go — Kernel 服务器入口
//go:build !cshared

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// 命令行参数
	addr := flag.String("addr", "", "TCP 监听地址 (默认: 127.0.0.1:9876)")
	lng := flag.String("lang", "en", "语言: en 或 zh")
	version := flag.Bool("version", false, "显示版本信息")
	flag.Parse()

	// 版本信息
	if *version {
		fmt.Println("MC Launcher Kernel v2.0.0-dev")
		fmt.Println("Build: Go", runtimeGOOS, "/", runtimeGOARCH)
		fmt.Println("API: JSON over TCP, port configurable")
		return
	}

	// 设置语言
	if *lng == "zh" || *lng == "zh-CN" {
		curLang = "zh"
	} else {
		curLang = "en"
	}

	// 初始化
	baseDir = launcherDir()
	ensureDir(baseDir)
	initLogger()
	cfg = loadSettings()
	accounts = loadAccounts()

	// 创建目录
	for _, d := range []string{
		"versions", "libraries",
		"assets/indexes", "assets/objects",
		"java", "logs", "sessions", "mods",
		"mods/modrinth", "mods curseforge",
		"saves", "screenshot",
	} {
		ensureDir(baseDir + "/" + d)
	}

	// 启动服务器
	serverAddr := *addr
	if serverAddr == "" {
		serverAddr = cfg.ListenAddr
	}
	if serverAddr == "" {
		serverAddr = "127.0.0.1:9876"
	}

	Serve(serverAddr)
}
