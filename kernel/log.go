package main

// log.go — 日志系统
import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var logMu sync.Mutex
var logW io.Writer = os.Stderr

func initLogger() {
	if baseDir != "" {
		logPath := filepath.Join(baseDir, "logs", "launcher.log")
		os.MkdirAll(filepath.Dir(logPath), 0755)
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			logW = io.MultiWriter(os.Stderr, f)
		}
	}
}
func logInfo(fmt2 string, args ...any) {
	logMu.Lock(); defer logMu.Unlock()
	fmt.Fprintf(logW, "[%s] [INF] "+fmt2+"\n", time.Now().Format("15:04:05"), args...)
}
func logDebug(fmt2 string, args ...any) {
	logMu.Lock(); defer logMu.Unlock()
	fmt.Fprintf(logW, "[%s] [DBG] "+fmt2+"\n", time.Now().Format("15:04:05"), args...)
}
func logError(fmt2 string, args ...any) {
	logMu.Lock(); defer logMu.Unlock()
	fmt.Fprintf(logW, "[%s] [ERR] "+fmt2+"\n", time.Now().Format("15:04:05"), args...)
}
func logWarn(fmt2 string, args ...any) {
	logMu.Lock(); defer logMu.Unlock()
	fmt.Fprintf(logW, "[%s] [WRN] "+fmt2+"\n", time.Now().Format("15:04:05"), args...)
}
