package main

// launch.go — 
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func launchVersion(versionId, username string) (*Response, error) {
	// find java
	java := pickJava(17)
	if java == nil {
		return &Response{Ok: false, Error: " Java 17+\n: java install"}, nil
	}
	// find account
	var accessToken, uuid string
	if acc := accounts.current(); acc != nil {
		accessToken = acc.AccessToken
		uuid = acc.UUID
	}
	if accessToken == "" {
		// offline
		acc := loginOffline(username)
		accessToken = acc.AccessToken
		uuid = acc.UUID
	}

	detail, err := fetchVersionDetail(versionId, cfg.DownloadSource)
	if err != nil { return &Response{Ok: false, Error: ": " + err.Error()}, nil }

	instDir := getInstalledDir(versionId)
	nativesDir := filepath.Join(instDir, "natives")

	// build classpath
	var cp []string
	cp = append(cp, filepath.Join(instDir, versionId+".jar"))
	for _, lib := range detail.Libraries {
		if lib.IsNative || lib.ResolvedPath == "" { continue }
		cp = append(cp, filepath.Join(getLibDir(versionId), lib.ResolvedPath))
	}
	// Linux  :Windows  ;
	sep := ":"
	if runtimeGOOS == "windows" { sep = ";" }
	cpStr := strings.Join(cp, sep)

	// build args
	var allArgs []string
	// jvm args
	for _, a := range detail.JVMArgs {
		allArgs = append(allArgs, substituteArgs(a, versionId, username, uuid, accessToken, instDir, nativesDir, cpStr))
	}
	allArgs = append(allArgs, "-Xmx"+itoa(cfg.MaxMemory)+"M")
	allArgs = append(allArgs, "-Xms"+itoa(cfg.MinMemory)+"M")
	for _, a := range cfg.JVMArgs { allArgs = append(allArgs, a) }
	allArgs = append(allArgs, "-cp", cpStr)
	allArgs = append(allArgs, detail.MainClass)
	// game args
	for _, a := range detail.GameArgs {
		allArgs = append(allArgs, substituteArgs(a, versionId, username, uuid, accessToken, instDir, nativesDir, cpStr))
	}
	// legacy args
	if detail.LegacyArgs != "" {
		for _, a := range strings.Fields(detail.LegacyArgs) {
			allArgs = append(allArgs, substituteArgs(a, versionId, username, uuid, accessToken, instDir, nativesDir, cpStr))
		}
	}
	// common args
	allArgs = append(allArgs, "--username", username)
	allArgs = append(allArgs, "--version", versionId)
	allArgs = append(allArgs, "--gameDir", instDir)
	allArgs = append(allArgs, "--assetsDir", filepath.Join(baseDir, "assets"))
	allArgs = append(allArgs, "--assetIndex", detail.AssetIndexID)
	allArgs = append(allArgs, "--uuid", uuid)
	allArgs = append(allArgs, "--accessToken", accessToken)
	allArgs = append(allArgs, "--userType", strings.Split(accounts.current().Type, ":")[0])
	if detail.LogArgument != "" { allArgs = append(allArgs, detail.LogArgument) }

	// build command
	var cmd *exec.Cmd
	cmd = exec.Command(java.Path, allArgs...)
	cmd.Dir = instDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	logInfo(": %s", strings.Join(cmd.Args, " "))
	err = cmd.Run()
	if err != nil {
		return &Response{Ok: false, Error: ": " + fmt.Sprintf("%v", err)}, nil
	}
	return &Response{Ok: true}, nil
}

func substituteArgs(arg, versionId, username, uuid, token, gameDir, nativesDir, cp string) string {
	repls := map[string]string{
		"${auth_player_name}": username,
		"${version_name}": versionId,
		"${game_directory}": gameDir,
		"${assets_root}": filepath.Join(baseDir, "assets"),
		"${assets_index_name}": "",
		"${auth_uuid}": uuid,
		"${auth_access_token}": token,
		"${user_type}": strings.Split(accounts.current().Type, ":")[0],
		"${resolution_width}": itoa(cfg.Width),
		"${resolution_height}": itoa(cfg.Height),
		"${natives_directory}": nativesDir,
		"${classpath}": cp,
	}
	for k, v := range repls {
		arg = strings.ReplaceAll(arg, k, v)
	}
	return arg
}
