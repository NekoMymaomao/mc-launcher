//go:build !cshared



// runtime.go — 
import "os"

var runtimeGOOS = func() string {
	if v := os.Getenv("MC_RUNTIME_GOOS"); v != "" { return v }
	return "linux"
}()
var runtimeGOARCH = func() string {
	if v := os.Getenv("MC_RUNTIME_GOARCH"); v != "" { return v }
	return "amd64"
}()
