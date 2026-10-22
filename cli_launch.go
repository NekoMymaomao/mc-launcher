package main

// cli_launch.go — CLI 直接启动游戏
//go:build !cshared

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// launchVersionDirect 直接启动游戏（不通过 TCP 服务器）
func launchVersionDirect(versionId, username string) (*Response, error) {
	// 查找 Java
	java := pickJava(17)
	if java == nil {
		return nil, fmt.Errorf("未找到 Java 17+")
	}

	// 查找账号
	var accessToken, uuid string
	if acc := accounts.current(); acc != nil {
		accessToken = acc.AccessToken
		uuid = acc.UUID
	}
	if accessToken == "" {
		// 离线模式
		acc := loginOffline(username)
		accessToken = acc.AccessToken
		uuid = acc.UUID
	}

	detail, err := fetchVersionDetail(versionId, cfg.DownloadSource)
	if err != nil {
		return nil, fmt.Errorf("获取版本详情失败: %w", err)
	}

	instDir := getInstalledDir(versionId)
	nativesDir := filepath.Join(instDir, "natives")

	// 构建 classpath
	var cp []string
	cp = append(cp, filepath.Join(instDir, versionId+".jar"))
	for _, lib := range detail.Libraries {
		if lib.IsNative || lib.ResolvedPath == "" {
			continue
		}
		cp = append(cp, filepath.Join(getLibDir(versionId), lib.ResolvedPath))
	}

	// 分隔符
	sep := ":"
	if runtimeGOOS == "windows" {
		sep = ";"
	}
	cpStr := strings.Join(cp, sep)

	// 构建参数
	var allArgs []string
	for _, a := range detail.JVMArgs {
		allArgs = append(allArgs, substituteArgs(a, versionId, username, uuid, accessToken, instDir, nativesDir, cpStr))
	}
	allArgs = append(allArgs, "-Xmx"+itoa(cfg.MaxMemory)+"M")
	allArgs = append(allArgs, "-Xms"+itoa(cfg.MinMemory)+"M")
	for _, a := range cfg.JVMArgs {
		allArgs = append(allArgs, a)
	}
	allArgs = append(allArgs, "-cp", cpStr)
	allArgs = append(allArgs, detail.MainClass)

	for _, a := range detail.GameArgs {
		allArgs = append(allArgs, substituteArgs(a, versionId, username, uuid, accessToken, instDir, nativesDir, cpStr))
	}
	if detail.LegacyArgs != "" {
		for _, a := range strings.Fields(detail.LegacyArgs) {
			allArgs = append(allArgs, substituteArgs(a, versionId, username, uuid, accessToken, instDir, nativesDir, cpStr))
		}
	}

	allArgs = append(allArgs,
		"--username", username,
		"--version", versionId,
		"--gameDir", instDir,
		"--assetsDir", filepath.Join(baseDir, "assets"),
		"--assetIndex", detail.AssetIndexID,
		"--uuid", uuid,
		"--accessToken", accessToken,
		"--userType", strings.Split(accounts.current().Type, ":")[0],
	)
	if detail.LogArgument != "" {
		allArgs = append(allArgs, detail.LogArgument)
	}

	// 启动游戏
	cmd := exec.Command(java.Path, allArgs...)
	cmd.Dir = instDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logInfo("启动 %s 使用 %s", versionId, java.Path)
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("启动失败: %w", err)
	}

	return &Response{Ok: true}, nil
}
