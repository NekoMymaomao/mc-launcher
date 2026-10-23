package main

// proto.go — 数据结构定义
// Request / Response 为 CLI 与（可选）内核共用的简单协议
type Request struct {
	Cmd  string         `json:"cmd"`
	Args map[string]any `json:"args,omitempty"`
	Cb   string         `json:"cb,omitempty"`
}
type Response struct {
	Ok    bool   `json:"ok"`
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}
type Progress struct {
	ID    string `json:"id"`
	Done  int64  `json:"done"`
	Total int64  `json:"total"`
	Msg   string `json:"msg"`
}

// Settings 启动器配置（settings.json）
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
	Lang             string   `json:"lang"`
}

// Account 账号。字段名与 JSON 键必须显式声明，
// 否则不同工具写出的 accounts.json 互不兼容。
type Account struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	UUID        string `json:"uuid"`
	AccessToken string `json:"accessToken"`
	Server      string `json:"server,omitempty"`
}

type AccountStore struct {
	Selected string    `json:"selected"`
	Accounts []Account `json:"accounts"`
}

func (a *AccountStore) current() *Account {
	if a == nil {
		return nil
	}
	if a.Selected == "" && len(a.Accounts) > 0 {
		a.Selected = a.Accounts[0].Type
	}
	for i := range a.Accounts {
		if a.Accounts[i].Type == a.Selected {
			return &a.Accounts[i]
		}
	}
	return nil
}
func (a *AccountStore) upsert(acc Account) {
	for i := range a.Accounts {
		if a.Accounts[i].Type == acc.Type {
			a.Accounts[i] = acc
			a.Selected = acc.Type
			saveAccounts()
			return
		}
	}
	a.Accounts = append(a.Accounts, acc)
	a.Selected = acc.Type
	saveAccounts()
}
func (a *AccountStore) remove(typ string) bool {
	for i := range a.Accounts {
		if a.Accounts[i].Type == typ {
			a.Accounts = append(a.Accounts[:i], a.Accounts[i+1:]...)
			if a.Selected == typ {
				a.Selected = ""
				if len(a.Accounts) > 0 {
					a.Selected = a.Accounts[0].Type
				}
			}
			saveAccounts()
			return true
		}
	}
	return false
}

// 版本元数据
type Artifact struct {
	Path string `json:"path,omitempty"`
	URL  string `json:"url,omitempty"`
	SHA1 string `json:"sha1,omitempty"`
	Size int64  `json:"size,omitempty"`
}
type OSRule struct {
	Name string `json:"name,omitempty"`
	Arch string `json:"arch,omitempty"`
}
type Rule struct {
	Action   string          `json:"action"`
	OS       *OSRule         `json:"os,omitempty"`
	Features map[string]bool `json:"features,omitempty"`
}
type Downloads struct {
	Artifact    Artifact            `json:"artifact"`
	Classifiers map[string]Artifact `json:"classifiers"`
}
type Extract struct {
	Exclude []string `json:"exclude"`
}
type Library struct {
	Name       string     `json:"name"`
	Downloads  *Downloads `json:"downloads,omitempty"`
	Natives    map[string]string `json:"natives,omitempty"`
	Rules      []Rule     `json:"rules,omitempty"`
	Extract    *Extract   `json:"extract,omitempty"`
	ShouldUse  bool       `json:"-"`
	IsNative   bool       `json:"-"`
	MavenBase  string     `json:"-"`
	ResolvedPath string   `json:"-"`
	ResolvedURL  string   `json:"-"`
	ResolvedSHA1 string   `json:"-"`
	NativePath   string   `json:"-"`
	NativeURL    string   `json:"-"`
	NativeSHA1   string   `json:"-"`
}
type VersionDetail struct {
	Valid        bool      `json:"-"`
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	MainClass    string    `json:"mainClass"`
	Assets       string    `json:"assets"`
	AssetIndexID string    `json:"-"`
	AssetIndex   Artifact  `json:"assetIndex"`
	Client       Artifact  `json:"-"`
	Libraries    []Library `json:"libraries"`
	JVMArgs      []string  `json:"-"`
	GameArgs     []string  `json:"-"`
	LegacyArgs   string    `json:"-"`
	JavaMajor    int       `json:"-"`
	LogArgument  string    `json:"-"`
	LogFile      Artifact  `json:"-"`
	RawJSON      string    `json:"-"`
}
type VersionEntry struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	URL  string `json:"url"`
	Time string `json:"releaseTime"`
}
type VersionManifest struct {
	Latest struct {
		Release  string `json:"release"`
		Snapshot string `json:"snapshot"`
	} `json:"latest"`
	Versions []VersionEntry `json:"versions"`
}

// 服务器列表条目（servers.json）
type ServerInfo struct {
	Address string `json:"address"`
	Name    string `json:"name"`
	Icon    string `json:"icon,omitempty"`
}
