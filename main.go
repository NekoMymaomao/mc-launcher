//go:build !cshared

package main
import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

// parseInt 安全解析整数
func parseInt(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

func printUsage() {
	fmt.Println(`MC Launcher - Minecraft 启动器 (CLI 版本)

用法:
  mc-launcher <命令> [参数]

命令:
  install <version>           安装 Minecraft 版本
  launch <version> [选项]     启动游戏
  list                        列出已安装版本
  remove <version>            删除版本
  verify <version>            验证版本完整性

  auth offline [name]         离线登录
  auth ms                     Microsoft 设备码登录
  auth authlib <server> <user> <pass>
                              Authlib 登录
  accounts                    列出账号
  accounts use <type>         切换当前账号
  accounts rm <type>          删除账号

  java list                   列出已安装 Java
  java install [major]        安装 Java (默认 17)
  java set <path>             设置 Java 路径

  mods <version> list         列出模组
  mods <version> add <path>   添加模组
  mods <version> rm <name>    删除模组

  modrinth search <query> [type]
                              搜索 Modrinth 项目
  modrinth info <slug>        查看项目详情
  modrinth install <slug> [mc_version]
                              安装整合包/模组
  modrinth shader <slug> [mc_version]
                              安装光影
  modrinth resourcepack <slug> [mc_version]
                              安装资源包
  modrinth datapack <slug> [mc_version]
                              安装数据包

  saves <version> list        列出存档
  saves <version> backup <name>
                              备份存档
  saves <version> delete <name>
                              删除存档

  servers list                列出服务器
  servers add <addr> [name]   添加服务器
  servers rm <addr>           删除服务器

  crashes list                列出崩溃报告
  crashes analyze <file>      分析崩溃

  config show                 显示配置
  config set <key> <value>    设置配置
  config mem <MB>             设置内存
  config jvmarg <arg>         添加 JVM 参数

  lang [en|zh]               切换语言
  version                     显示版本信息

选项:
  -lang en|zh     设置语言
  -v, --version   显示版本
`)
}

func main() {
	// 解析全局标志
	lng := flag.String("lang", "", "语言: en 或 zh")
	versionFlag := flag.Bool("version", false, "显示版本信息")
	flag.Parse()

	// 设置语言
	if *lng == "zh" || *lng == "zh-CN" {
		curLang = "zh"
	} else if *lng == "en" {
		curLang = "en"
	}

	// 版本信息
	if *versionFlag {
		fmt.Printf("MC Launcher v2.0.0-dev\n")
		fmt.Printf("Build: Go %s/%s\n", runtimeGOOS, runtimeGOARCH)
		fmt.Printf("Language: %s\n", curLang)
		return
	}

	// 无参数显示帮助
	if len(flag.Args()) == 0 {
		printUsage()
		return
	}

	// 初始化
	baseDir = launcherDir()
	ensureDir(baseDir)
	initLogger()
	cfg = loadSettings()
	accounts = loadAccounts()
	if cfg.Lang != "" && *lng == "" {
		curLang = cfg.Lang
	}

	// 创建目录
	for _, d := range []string{
		"versions", "libraries",
		"assets/indexes", "assets/objects",
		"java", "logs", "sessions", "mods",
		"mods/modrinth", "mods/curseforge",
		"saves", "shaderpacks", "resourcepacks", "data_packs",
	} {
		ensureDir(baseDir + "/" + d)
	}

	// 命令路由
	cmd := flag.Arg(0)
	args := flag.Args()[1:]

	switch cmd {
	// ===== 版本管理 =====
	case "install":
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "用法: mc-launcher install <version>\n")
			os.Exit(1)
		}
		resp, err := installVersion(args[0], "", func(p Progress) {
			fmt.Printf("[%d/%d] %s\n", p.Done, p.Total, p.Msg)
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}
		if resp.Ok {
			fmt.Printf("已安装版本: %s\n", args[0])
		} else {
			fmt.Fprintf(os.Stderr, "错误: %s\n", resp.Error)
			os.Exit(1)
		}

	case "launch":
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "用法: mc-launcher launch <version> [username]\n")
			os.Exit(1)
		}
		username := "Player"
		if len(args) > 1 {
			username = args[1]
		}
		resp, err := launchVersionDirect(args[0], username)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}
		if resp.Ok {
			fmt.Printf("已启动游戏: %s\n", args[0])
		} else {
			fmt.Fprintf(os.Stderr, "错误: %s\n", resp.Error)
			os.Exit(1)
		}

	case "list", "ls":
		versions := listVersions()
		if len(versions) == 0 {
			fmt.Println("没有已安装的版本")
		} else {
			fmt.Println("已安装版本:")
			for _, v := range versions {
				fmt.Printf("  %s\n", v)
			}
		}

	case "remove", "rm":
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "用法: mc-launcher remove <version>\n")
			os.Exit(1)
		}
		removeVersion(args[0])
		fmt.Printf("已删除版本: %s\n", args[0])

	case "verify":
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "用法: mc-launcher verify <version>\n")
			os.Exit(1)
		}
		ok, msg := verifyVersion(args[0])
		if ok {
			fmt.Printf("版本 %s 验证通过\n", args[0])
		} else {
			fmt.Fprintf(os.Stderr, "验证失败: %s\n", msg)
			os.Exit(1)
		}

	// ===== 认证 =====
	case "auth":
		if len(args) < 1 {
			fmt.Println("认证命令:")
			fmt.Println("  offline [name]          - 离线登录")
			fmt.Println("  ms                      - Microsoft 设备码登录")
			fmt.Println("  authlib <server> <user> <pass> - Authlib 登录")
			return
		}
		switch args[0] {
		case "offline":
			name := "Player"
			if len(args) > 1 {
				name = args[1]
			}
			acc := loginOffline(name)
			accounts.upsert(*acc)
			fmt.Printf("已登录: %s (%s)\n", acc.Name, acc.UUID)

		case "ms":
			acc, err := loginMicrosoft(true, "")
			if err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				os.Exit(1)
			}
			accounts.upsert(*acc)
			fmt.Printf("已登录: %s (%s)\n", acc.Name, acc.UUID)

		case "authlib":
			if len(args) < 4 {
				fmt.Fprintf(os.Stderr, "用法: mc-launcher auth authlib <server> <user> <pass>\n")
				os.Exit(1)
			}
			acc, err := loginAuthlib(args[1], args[2], args[3])
			if err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				os.Exit(1)
			}
			accounts.upsert(*acc)
			fmt.Printf("已登录: %s (%s)\n", acc.Name, acc.UUID)

		default:
			fmt.Fprintf(os.Stderr, "未知命令: %s\n", args[0])
			os.Exit(1)
		}

	case "accounts", "account":
		if len(args) > 0 {
			switch args[0] {
			case "use", "select", "switch":
				if len(args) < 2 {
					fmt.Fprintf(os.Stderr, "用法: mc-launcher accounts use <type>\n")
					os.Exit(1)
				}
				found := false
				for _, a := range accounts.Accounts {
					if a.Type == args[1] {
						found = true
						break
					}
				}
				if !found {
					fmt.Fprintf(os.Stderr, "没有类型为 %s 的账号\n", args[1])
					os.Exit(1)
				}
				accounts.Selected = args[1]
				saveAccounts()
				fmt.Printf("已切换账号: %s\n", args[1])
				return
			case "rm", "remove", "delete":
				if len(args) < 2 {
					fmt.Fprintf(os.Stderr, "用法: mc-launcher accounts rm <type>\n")
					os.Exit(1)
				}
				if !accounts.remove(args[1]) {
					fmt.Fprintf(os.Stderr, "没有类型为 %s 的账号\n", args[1])
					os.Exit(1)
				}
				fmt.Printf("已删除账号: %s\n", args[1])
				return
			}
		}
		if len(accounts.Accounts) == 0 {
			fmt.Println("没有已登录的账号")
		} else {
			fmt.Println("已登录账号:")
			for _, a := range accounts.Accounts {
				selected := ""
				if a.Type == accounts.Selected {
					selected = " *"
				}
				fmt.Printf("  [%s] %s (%s)%s\n", a.Type, a.Name, a.UUID, selected)
			}
		}

	// ===== Java 管理 =====
	case "java":
		if len(args) < 1 {
			fmt.Println("Java 管理命令:")
			fmt.Println("  list                    - 列出已安装 Java")
			fmt.Println("  install [major]         - 安装 Java (默认 17)")
			fmt.Println("  set <path>              - 设置 Java 路径")
			return
		}
		switch args[0] {
		case "list", "ls":
			javas := scanJava()
			if len(javas) == 0 {
				fmt.Println("未找到 Java")
			} else {
				fmt.Println("已安装 Java:")
				for _, j := range javas {
					fmt.Printf("  %s (Java %d)\n", j.Path, j.Major)
				}
			}

		case "install":
			major := 17
			if len(args) > 1 {
				major = parseInt(args[1])
			}
			path, err := installJava(major)
			if err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Java %d 已安装: %s\n", major, path)

		case "set":
			if len(args) < 2 {
				fmt.Fprintf(os.Stderr, "用法: mc-launcher java set <path>\n")
				os.Exit(1)
			}
			cfg.JavaPath = args[1]
			cfg.save()
			fmt.Printf("已设置 Java 路径: %s\n", args[1])

		default:
			fmt.Fprintf(os.Stderr, "未知命令: %s\n", args[0])
			os.Exit(1)
		}

	// ===== 模组管理 =====
	case "mods":
		if len(args) < 2 {
			fmt.Println("模组管理命令:")
			fmt.Println("  <version> list          - 列出模组")
			fmt.Println("  <version> add <path>    - 添加模组")
			fmt.Println("  <version> rm <name>     - 删除模组")
			return
		}
		version := args[0]
		switch args[1] {
		case "list", "ls":
			mods := listMods(version)
			if len(mods) == 0 {
				fmt.Printf("版本 %s 没有模组\n", version)
			} else {
				fmt.Printf("版本 %s 的模组 (%d 个):\n", version, len(mods))
				for _, m := range mods {
					fmt.Printf("  %s\n", m)
				}
			}

		case "add", "install":
			if len(args) < 3 {
				fmt.Fprintf(os.Stderr, "用法: mc-launcher mods <version> add <path>\n")
				os.Exit(1)
			}
			ok := addMod(version, args[2])
			if ok {
				fmt.Printf("已添加模组: %s\n", args[2])
			} else {
				fmt.Fprintf(os.Stderr, "添加失败: %s\n", args[2])
				os.Exit(1)
			}

		case "rm", "remove", "delete":
			if len(args) < 3 {
				fmt.Fprintf(os.Stderr, "用法: mc-launcher mods <version> rm <name>\n")
				os.Exit(1)
			}
			ok := removeMod(version, args[2])
			if ok {
				fmt.Printf("已删除模组: %s\n", args[2])
			} else {
				fmt.Fprintf(os.Stderr, "删除失败: %s\n", args[2])
			}

		default:
			fmt.Fprintf(os.Stderr, "未知命令: %s\n", args[1])
			os.Exit(1)
		}

	// ===== Modrinth =====
	case "modrinth":
		if len(args) < 1 {
			fmt.Println("Modrinth 命令:")
			fmt.Println("  search <query> [type]   - 搜索项目")
			fmt.Println("  info <slug>             - 项目详情")
			fmt.Println("  install <slug> [mc_ver] - 安装整合包")
			fmt.Println("  shader <slug> [mc_ver]  - 安装光影")
			fmt.Println("  resourcepack <slug> [mc_ver] - 安装资源包")
			fmt.Println("  datapack <slug> [mc_ver] - 安装数据包")
			return
		}
		switch args[0] {
		case "search":
			if len(args) < 2 {
				fmt.Fprintf(os.Stderr, "用法: mc-launcher modrinth search <query> [type]\n")
				os.Exit(1)
			}
			q := args[1]
			ptype := ""
			if len(args) > 2 {
				ptype = args[2]
			}
			resp := modrinthSearch(q, ptype, 1)
			if !resp.Ok {
				fmt.Fprintf(os.Stderr, "错误: %s\n", resp.Error)
				os.Exit(1)
			}
			hits := resp.Data.([]map[string]any)
			if len(hits) == 0 {
				fmt.Printf("未找到匹配 '%s' 的项目\n", q)
			} else {
				fmt.Printf("搜索 '%s' 的结果 (%d 个):\n", q, len(hits))
				for _, h := range hits {
					title := h["title"].(string)
					slug := h["slug"].(string)
					typ := h["type"].(string)
					dl := h["downloads"].(int64)
					fmt.Printf("  %-20s [%s] %d downloads\n", slug, typ, dl)
					fmt.Printf("    %s\n", title)
				}
			}

		case "info":
			if len(args) < 2 {
				fmt.Fprintf(os.Stderr, "用法: mc-launcher modrinth info <slug>\n")
				os.Exit(1)
			}
			resp := modrinthProject(args[1])
			if !resp.Ok {
				fmt.Fprintf(os.Stderr, "错误: %s\n", resp.Error)
				os.Exit(1)
			}
			p := resp.Data.(map[string]any)
			fmt.Printf("项目: %s\n", p["title"])
			fmt.Printf("  Slug:    %s\n", p["slug"])
			fmt.Printf("  Type:    %s\n", p["type"])
			fmt.Printf("  Downloads: %d\n", p["downloads"])
			fmt.Printf("  URL:     %s\n", p["url"])

		case "install":
			if len(args) < 2 {
				fmt.Fprintf(os.Stderr, "用法: mc-launcher modrinth install <slug> [mc_version]\n")
				os.Exit(1)
			}
			mcVer := ""
			if len(args) > 2 {
				mcVer = args[2]
			}
			resp := modrinthPackInstall(args[1], mcVer)
			if !resp.Ok {
				fmt.Fprintf(os.Stderr, "错误: %s\n", resp.Error)
				os.Exit(1)
			}
			fmt.Printf("已安装: %s\n", resp.Data.(map[string]any)["title"])

		case "shader":
			if len(args) < 2 {
				fmt.Fprintf(os.Stderr, "用法: mc-launcher modrinth shader <slug> [mc_version]\n")
				os.Exit(1)
			}
			mcVer := ""
			if len(args) > 2 {
				mcVer = args[2]
			}
			resp := modrinthShaderInstall(args[1], mcVer)
			if !resp.Ok {
				fmt.Fprintf(os.Stderr, "错误: %s\n", resp.Error)
				os.Exit(1)
			}
			fmt.Printf("已安装光影: %s\n", resp.Data.(map[string]any)["title"])

		case "resourcepack":
			if len(args) < 2 {
				fmt.Fprintf(os.Stderr, "用法: mc-launcher modrinth resourcepack <slug> [mc_version]\n")
				os.Exit(1)
			}
			mcVer := ""
			if len(args) > 2 {
				mcVer = args[2]
			}
			resp := modrinthResourcePackInstall(args[1], mcVer)
			if !resp.Ok {
				fmt.Fprintf(os.Stderr, "错误: %s\n", resp.Error)
				os.Exit(1)
			}
			fmt.Printf("已安装资源包: %s\n", resp.Data.(map[string]any)["title"])

		case "datapack":
			if len(args) < 2 {
				fmt.Fprintf(os.Stderr, "用法: mc-launcher modrinth datapack <slug> [mc_version]\n")
				os.Exit(1)
			}
			mcVer := ""
			if len(args) > 2 {
				mcVer = args[2]
			}
			resp := modrinthDataPackInstall(args[1], mcVer)
			if !resp.Ok {
				fmt.Fprintf(os.Stderr, "错误: %s\n", resp.Error)
				os.Exit(1)
			}
			fmt.Printf("已安装数据包: %s\n", resp.Data.(map[string]any)["title"])

		default:
			fmt.Fprintf(os.Stderr, "未知命令: %s\n", args[0])
			os.Exit(1)
		}

	// ===== 存档管理 =====
	case "saves":
		if len(args) < 2 {
			fmt.Println("存档管理命令:")
			fmt.Println("  <version> list          - 列出存档")
			fmt.Println("  <version> backup <name> - 备份存档")
			fmt.Println("  <version> delete <name> - 删除存档")
			return
		}
		version := args[0]
		switch args[1] {
		case "list", "ls":
			saves := listSaves(version)
			if len(saves) == 0 {
				fmt.Printf("版本 %s 没有存档\n", version)
			} else {
				fmt.Printf("版本 %s 的存档 (%d 个):\n", version, len(saves))
				for _, s := range saves {
					fmt.Printf("  %s\n", s)
				}
			}

		case "backup":
			if len(args) < 3 {
				fmt.Fprintf(os.Stderr, "用法: mc-launcher saves <version> backup <name>\n")
				os.Exit(1)
			}
			backupSave(version, args[2])
			fmt.Printf("已备份存档: %s\n", args[2])

		case "delete", "del", "rm":
			if len(args) < 3 {
				fmt.Fprintf(os.Stderr, "用法: mc-launcher saves <version> delete <name>\n")
				os.Exit(1)
			}
			deleteSave(version, args[2])
			fmt.Printf("已删除存档: %s\n", args[2])

		default:
			fmt.Fprintf(os.Stderr, "未知命令: %s\n", args[1])
			os.Exit(1)
		}

	// ===== 服务器管理 =====
	case "servers":
		if len(args) < 1 {
			fmt.Println("服务器管理命令:")
			fmt.Println("  list                    - 列出服务器")
			fmt.Println("  add <addr> [name]       - 添加服务器")
			fmt.Println("  rm <addr>               - 删除服务器")
			return
		}
		switch args[0] {
		case "list", "ls":
			servers := listServers()
			if len(servers) == 0 {
				fmt.Println("没有已保存的服务器")
			} else {
				fmt.Println("已保存的服务器:")
				for _, s := range servers {
					fmt.Printf("  %-30s %s\n", s.Address, s.Name)
				}
			}

		case "add":
			if len(args) < 2 {
				fmt.Fprintf(os.Stderr, "用法: mc-launcher servers add <addr> [name]\n")
				os.Exit(1)
			}
			name := args[1]
			if len(args) > 2 {
				name = args[2]
			}
			addServer(args[1], name)
			fmt.Printf("已添加服务器: %s\n", args[1])

		case "rm", "remove", "delete":
			if len(args) < 2 {
				fmt.Fprintf(os.Stderr, "用法: mc-launcher servers rm <addr>\n")
				os.Exit(1)
			}
			removeServer(args[1])
			fmt.Printf("已删除服务器: %s\n", args[1])

		default:
			fmt.Fprintf(os.Stderr, "未知命令: %s\n", args[0])
			os.Exit(1)
		}

	// ===== 崩溃分析 =====
	case "crashes", "crash":
		if len(args) < 1 {
			fmt.Println("崩溃分析命令:")
			fmt.Println("  list                    - 列出崩溃报告")
			fmt.Println("  analyze <file>          - 分析崩溃")
			return
		}
		switch args[0] {
		case "list", "ls":
			versions := listVersions()
			found := false
			for _, v := range versions {
				crashes := listCrashes(v)
				if len(crashes) > 0 {
					found = true
					fmt.Printf("版本 %s 的崩溃报告 (%d 个):\n", v, len(crashes))
					for _, c := range crashes {
						fmt.Printf("  %s\n", c)
					}
				}
			}
			if !found {
				fmt.Println("没有崩溃报告")
			}

		case "analyze":
			if len(args) < 2 {
				fmt.Fprintf(os.Stderr, "用法: mc-launcher crashes analyze <file>\n")
				os.Exit(1)
			}
			content := analyzeCrash("", args[1])
			if content == "" {
				fmt.Println("无法分析崩溃报告")
			} else {
				fmt.Println("崩溃分析结果:")
				fmt.Println(content)
			}

		default:
			fmt.Fprintf(os.Stderr, "未知命令: %s\n", args[0])
			os.Exit(1)
		}

	// ===== 配置 =====
	case "config":
		if len(args) < 1 {
			fmt.Println("配置命令:")
			fmt.Println("  show                    - 显示配置")
			fmt.Println("  set <key> <value>       - 设置配置")
			fmt.Println("  mem <MB>                - 设置内存")
			fmt.Println("  jvmarg <arg>            - 添加 JVM 参数")
			return
		}
		switch args[0] {
		case "show":
			fmt.Println("当前配置:")
			fmt.Printf("  maxMemory:   %d MB\n", cfg.MaxMemory)
			fmt.Printf("  minMemory:   %d MB\n", cfg.MinMemory)
			fmt.Printf("  javaPath:    %s\n", cfg.JavaPath)
			fmt.Printf("  source:      %s\n", cfg.DownloadSource)
			fmt.Printf("  width:       %d\n", cfg.Width)
			fmt.Printf("  height:      %d\n", cfg.Height)
			fmt.Printf("  threads:     %d\n", cfg.Threads)

		case "set":
			if len(args) < 3 {
				fmt.Fprintf(os.Stderr, "用法: mc-launcher config set <key> <value>\n")
				os.Exit(1)
			}
			key := args[1]
			value := args[2]
			switch key {
			case "maxMemory":
				cfg.MaxMemory = parseInt(value)
			case "minMemory":
				cfg.MinMemory = parseInt(value)
			case "source":
				cfg.DownloadSource = value
			case "width":
				cfg.Width = parseInt(value)
			case "height":
				cfg.Height = parseInt(value)
			default:
				fmt.Fprintf(os.Stderr, "未知配置项: %s\n", key)
				os.Exit(1)
			}
			cfg.save()
			fmt.Printf("已设置 %s = %s\n", key, value)

		case "mem":
			if len(args) < 2 {
				fmt.Fprintf(os.Stderr, "用法: mc-launcher config mem <MB>\n")
				os.Exit(1)
			}
			cfg.MaxMemory = parseInt(args[1])
			cfg.save()
			fmt.Printf("已设置内存: %d MB\n", cfg.MaxMemory)

		case "jvmarg":
			if len(args) < 2 {
				fmt.Fprintf(os.Stderr, "用法: mc-launcher config jvmarg <arg>\n")
				os.Exit(1)
			}
			cfg.JVMArgs = append(cfg.JVMArgs, args[1])
			cfg.save()
			fmt.Printf("已添加 JVM 参数: %s\n", args[1])

		default:
			fmt.Fprintf(os.Stderr, "未知命令: %s\n", args[0])
			os.Exit(1)
		}

	// ===== 语言 =====
	case "lang":
		if len(args) < 1 {
			fmt.Printf("当前语言: %s\n", curLang)
			fmt.Println("可用语言:")
			fmt.Println("  en - English")
			fmt.Println("  zh - Chinese")
			return
		}
		switch args[0] {
		case "en", "english":
			curLang = "en"
			cfg.Lang = "en"
			cfg.save()
			fmt.Println("语言已切换: English")
		case "zh", "chinese", "zh-cn":
			curLang = "zh"
			cfg.Lang = "zh"
			cfg.save()
			fmt.Println("语言已切换: 中文")
		default:
			fmt.Fprintf(os.Stderr, "未知语言: %s\n", args[0])
			os.Exit(1)
		}

	// ===== 帮助 =====
	case "help", "-h", "--help":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "未知命令: %s\n", cmd)
		fmt.Println()
		printUsage()
		os.Exit(1)
	}
}
