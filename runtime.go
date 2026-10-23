package main

// runtime.go — 运行平台信息
// 之前用环境变量 MC_RUNTIME_GOOS 覆盖并默认写死 "linux"，
// 导致 Windows 构建拿到 linux：classpath 分隔符、natives 全部错。改用标准库。
import "runtime"

var (
	runtimeGOOS   = runtime.GOOS
	runtimeGOARCH = runtime.GOARCH
)
