package main

// launch.go — 启动游戏
//
// 旧的实现在三个地方必然崩溃/失效：
//  1. accounts.current() 在没有账号时返回 nil，第 78 行直接 .Type 解引用 panic；
//  2. pickJava(17) 写死，1.12 这类只需要 Java 8 的版本反而启动不了；
//  3. 版本 JSON 走了一个不存在的 URL，拿不到 mainClass。
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// launchVersionDirect 组装命令行并拉起游戏进程
func launchVersionDirect(versionId, username string) (*Response, error) {
	if versionId == "" {
		return nil, fmt.Errorf("版本号为空")
	}

	detail, err := fetchVersionDetail(versionId, cfg.DownloadSource)
	if err != nil {
		return nil, err
	}

	// ---- Java ----
	need := detail.JavaMajor
	if need <= 0 {
		need = 8
	}
	java := pickJava(need)
	if java == nil {
		return nil, fmt.Errorf("未找到 Java %d+，可用 `mc-launcher java install %d` 安装，或用 `java set <path>` 指定", need, need)
	}
	if cfg.JavaPath != java.Path {
		cfg.JavaPath = java.Path
		cfg.save()
	}

	// ---- 账号 ----
	acc := accounts.current()
	if acc == nil || acc.AccessToken == "" {
		acc = loginOffline(username)
	}
	if acc.Name != "" {
		username = acc.Name
	}
	if username == "" {
		username = "Player"
	}
	userType := userTypeOf(acc)

	instDir := getInstalledDir(versionId)
	libDir := getLibDir(versionId)
	nativesDir := filepath.Join(instDir, "natives")
	assetsDir := filepath.Join(baseDir, "assets")

	// ---- classpath ----
	var cp []string
	cp = append(cp, filepath.Join(instDir, versionId+".jar"))
	for _, lib := range detail.Libraries {
		if lib.IsNative || lib.ResolvedPath == "" {
			continue
		}
		cp = append(cp, filepath.Join(libDir, lib.ResolvedPath))
	}
	sep := ":"
	if runtimeGOOS == "windows" {
		sep = ";"
	}
	cpStr := strings.Join(cp, sep)

	sub := func(a string) string {
		return substituteArgs(a, versionId, username, acc.UUID, acc.AccessToken, userType, instDir, nativesDir, assetsDir, detail.AssetIndexID, cpStr)
	}

	// ---- JVM 参数 ----
	var allArgs []string
	for _, a := range detail.JVMArgs {
		allArgs = append(allArgs, sub(a))
	}
	for _, a := range cfg.JVMArgs {
		allArgs = append(allArgs, a)
	}
	allArgs = append(allArgs, "-Xmx"+itoa(cfg.MaxMemory)+"M")
	if cfg.MinMemory > 0 {
		allArgs = append(allArgs, "-Xms"+itoa(cfg.MinMemory)+"M")
	}
	allArgs = append(allArgs, "-Djava.library.path="+nativesDir)

	// 日志配置
	if detail.LogArgument != "" {
		logPath := filepath.Join(instDir, logFileName(detail))
		allArgs = append(allArgs, strings.ReplaceAll(detail.LogArgument, "${path}", logPath))
	}

	allArgs = append(allArgs, "-cp", cpStr, detail.MainClass)

	// ---- 游戏参数 ----
	var gameArgs []string
	for _, a := range detail.GameArgs {
		gameArgs = append(gameArgs, sub(a))
	}
	if detail.LegacyArgs != "" {
		for _, a := range splitArgs(detail.LegacyArgs) {
			gameArgs = append(gameArgs, sub(a))
		}
	}
	for _, a := range cfg.GameArgs {
		gameArgs = append(gameArgs, a)
	}

	// 现代版本 JSON 自带 --username/--uuid 等；老版本没有时补齐
	if !containsFlag(gameArgs, "--username") {
		gameArgs = append(gameArgs,
			"--username", username,
			"--version", versionId,
			"--gameDir", instDir,
			"--assetsDir", assetsDir,
			"--assetIndex", detail.AssetIndexID,
			"--uuid", acc.UUID,
			"--accessToken", acc.AccessToken,
			"--userType", userType,
		)
	}
	allArgs = append(allArgs, gameArgs...)

	// ---- 启动 ----
	cmd := exec.Command(java.Path, allArgs...)
	cmd.Dir = instDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	logInfo("使用 Java %d (%s) 启动 %s", java.Major, java.Path, versionId)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("游戏进程退出: %w", err)
	}
	return &Response{Ok: true, Data: map[string]any{"id": versionId}}, nil
}

// userTypeOf 映射 Minecraft 需要的 userType
func userTypeOf(acc *Account) string {
	if acc == nil {
		return "legacy"
	}
	switch acc.Type {
	case "microsoft", "msa":
		return "msa"
	default:
		return "legacy"
	}
}

func substituteArgs(arg, versionId, username, uuid, token, userType, gameDir, nativesDir, assetsDir, assetIndex, cp string) string {
	repls := map[string]string{
		"${auth_player_name}":    username,
		"${version_name}":        versionId,
		"${game_directory}":      gameDir,
		"${assets_root}":         assetsDir,
		"${assets_index_name}":   assetIndex,
		"${auth_uuid}":           uuid,
		"${auth_access_token}":   token,
		"${auth_session}":        token,
		"${user_type}":           userType,
		"${version_type}":        "",
		"${resolution_width}":    itoa(cfg.Width),
		"${resolution_height}":   itoa(cfg.Height),
		"${natives_directory}":   nativesDir,
		"${launcher_name}":       "mc-launcher",
		"${launcher_version}":    "2.0.0",
		"${classpath}":           cp,
		"${classpath_separator}": separator(),
		"${library_directory}":   filepath.Join(baseDir, "libraries"),
	}
	for k, v := range repls {
		arg = strings.ReplaceAll(arg, k, v)
	}
	return arg
}

func separator() string {
	if runtimeGOOS == "windows" {
		return ";"
	}
	return ":"
}

// splitArgs 按空格切分旧版 minecraftArguments，保留引号内的空格
func splitArgs(s string) []string {
	var out []string
	var cur strings.Builder
	inQuote := false
	for _, r := range s {
		switch {
		case r == '"':
			inQuote = !inQuote
		case r == ' ' && !inQuote:
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

func containsFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}
