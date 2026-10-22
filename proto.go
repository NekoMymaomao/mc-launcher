package main

// ============================================================
// 协议 + 数据结构（纯标准库）
// ============================================================
import (
	"encoding/json"
)

// 请求/响应
type Request struct {
	Cmd  string         `json:"cmd"`
	Args map[string]any `json:"args,omitempty"`
	Cb   string         `json:"cb,omitempty"`
}
type Response struct {
	Ok    bool     `json:"ok"`
	Data  any      `json:"data,omitempty"`
	Error string   `json:"error,omitempty"`
}
type Progress struct {
	ID    string `json:"id"`
	Done  int64  `json:"done"`
	Total int64  `json:"total"`
	Msg   string `json:"msg"`
}

// 配置
type Settings struct {
	MaxMemory        int      `json:"maxMemory"`
	MinMemory        int      `json:"minMemory"`
	JavaPath         string   `json:"javaPath"`
	JVMArgs          []string `json:"jvmArgs"`
	GameArgs         []string `json:"gameArgs"`
	DownloadSource   string   `json:"source"`
	Width            int      `json:"width"`
	Height           int      `json:"height"`
	VersionIsolation bool     `json:"versionIsolation"`
	Threads          int      `json:"threads"`
	ListenAddr       string   `json:"listenAddr"`
}
type Account struct {
	Type, Name, UUID, AccessToken, Server string
}
type AccountStore struct {
	Selected string     `json:"selected"`
	Accounts []Account  `json:"accounts"`
}
func (a *AccountStore) current() *Account {
	if a.Selected == "" && len(a.Accounts) > 0 { a.Selected = a.Accounts[0].Type }
	for i := range a.Accounts {
		if a.Accounts[i].Type == a.Selected { return &a.Accounts[i] }
	}
	return nil
}
func (a *AccountStore) upsert(acc Account) {
	for i := range a.Accounts {
		if a.Accounts[i].Type == acc.Type {
			a.Accounts[i] = acc; a.Selected = acc.Type; saveAccounts(); return
		}
	}
	a.Accounts = append(a.Accounts, acc); a.Selected = acc.Type; saveAccounts()
}
func (a *AccountStore) remove(typ string) bool {
	for i := range a.Accounts {
		if a.Accounts[i].Type == typ {
			a.Accounts = append(a.Accounts[:i], a.Accounts[i+1:]...)
			if a.Selected == typ {
				a.Selected = ""
				if len(a.Accounts) > 0 { a.Selected = a.Accounts[0].Type }
			}
			saveAccounts(); return true
		}
	}
	return false
}

// 版本/资源
type Artifact struct { Path, URL, SHA1 string; Size int64 }
type OSRule struct { Name, Arch, Version string }
type Rule struct { Action string; OS *OSRule; Features map[string]bool }
type Downloads struct { Artifact Artifact; Classifiers map[string]Artifact }
type Extract struct { Exclude []string }
type Library struct {
	Name string; Downloads *Downloads; Natives map[string]string
	Rules []Rule; Extract *Extract
	ShouldUse, IsNative bool
	MavenBase string
	ResolvedPath, ResolvedURL, NativePath, NativeURL string
}
type ArgEl struct { Rules []Rule; Value []string }
type VersionDetail struct {
	Valid bool; ID, Type, MainClass, Assets, AssetIndexID string
	AssetIndex Artifact; Client Artifact
	Libraries []Library
	JVMArgs, GameArgs []string
	LegacyArgs string
	JavaMajor int
	LogArgument string
	LogFile Artifact
}
type VersionEntry struct { ID, Type, URL, Time string }
type VersionManifest struct {
	Latest struct{ Release, Snapshot string }; Versions []VersionEntry
}
