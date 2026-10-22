//go:build !cshared



// globals.go — 全局变量声明

var (
	baseDir  string
	cfg      *Settings
	accounts *AccountStore
	curLang  string // "en" or "zh"
)

func init() {
	// 默认语言
	curLang = "en"
	if v := os.Getenv("MC_LAUNCHER_LANG"); v == "zh" || v == "zh-CN" {
		curLang = "zh"
	}
}
