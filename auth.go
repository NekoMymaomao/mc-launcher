package main

// auth.go — 认证（离线 / 微软设备码 / Authlib）
//
// 旧实现的设备码流程有两个致命问题：
//  1. 给 /devicecode 发的是 JSON 体 + Content-Type: application/json，
//     微软只接受 application/x-www-form-urlencoded；
//  2. 设备码返回的字段是 device_code / user_code / verification_uri / interval，
//     结构体没有 json tag，Go 匹配不到，全部是空值，直接报错退出。
import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

func offlineUuid(name string) string {
	h := md5.Sum([]byte("OfflinePlayer:" + name))
	b := h[:]
	b[6] = (b[6] & 0x0F) | 0x30
	b[8] = (b[8] & 0x3F) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func loginOffline(name string) *Account {
	if name == "" {
		name = "Player"
	}
	uuid := offlineUuid(name)
	return &Account{
		Type:        "offline",
		Name:        name,
		UUID:        uuid,
		AccessToken: strings.ReplaceAll(uuid, "-", ""),
	}
}

const (
	msDeviceCodeURL = "https://login.microsoftonline.com/consumers/oauth2/v2.0/devicecode"
	msTokenURL      = "https://login.microsoftonline.com/consumers/oauth2/v2.0/token"
	msScope         = "XboxLive.signin offline_access"
)

type deviceCodeResp struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
	Message         string `json:"message"`
}

type tokenResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
}

// loginMicrosoft 设备码登录
func loginMicrosoft(verbose bool, clientId string) (*Account, error) {
	if clientId == "" {
		clientId = os.Getenv("MC_MS_CLIENT_ID")
		if clientId == "" {
			// 官方启动器 client id，公开可用
			clientId = "00000000402b5328"
		}
	}

	form := url.Values{}
	form.Set("client_id", clientId)
	form.Set("scope", msScope)

	body, code, err := httpPostForm(msDeviceCodeURL, form.Encode())
	if err != nil {
		return nil, fmt.Errorf("请求设备码失败: %w", err)
	}
	if code != 200 {
		return nil, fmt.Errorf("请求设备码失败: HTTP %d %s", code, strings.TrimSpace(string(body)))
	}
	var dc deviceCodeResp
	if err := json.Unmarshal(body, &dc); err != nil {
		return nil, fmt.Errorf("解析设备码失败: %w", err)
	}
	if dc.DeviceCode == "" {
		return nil, fmt.Errorf("设备码响应缺少 device_code")
	}
	if verbose {
		fmt.Printf("请在浏览器打开: %s\n并输入代码: %s\n", dc.VerificationURI, dc.UserCode)
		fmt.Println("等待授权中...")
	}

	interval := dc.Interval
	if interval <= 0 {
		interval = 5
	}
	deadline := time.Now().Add(time.Duration(dc.ExpiresIn) * time.Second)
	if dc.ExpiresIn <= 0 {
		deadline = time.Now().Add(15 * time.Minute)
	}

	tokenForm := url.Values{}
	tokenForm.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	tokenForm.Set("client_id", clientId)
	tokenForm.Set("device_code", dc.DeviceCode)

	var msToken string
	for time.Now().Before(deadline) {
		time.Sleep(time.Duration(interval) * time.Second)
		b, code, err := httpPostForm(msTokenURL, tokenForm.Encode())
		if err != nil {
			continue
		}
		var tr tokenResp
		json.Unmarshal(b, &tr)
		if code == 200 && tr.AccessToken != "" {
			msToken = tr.AccessToken
			break
		}
		switch tr.Error {
		case "authorization_pending":
			continue
		case "slow_down":
			interval += 5
		case "authorization_declined":
			return nil, fmt.Errorf("用户在浏览器中拒绝了授权")
		case "expired_token":
			return nil, fmt.Errorf("设备码已过期，请重试")
		default:
			if code != 200 && tr.Error != "" {
				return nil, fmt.Errorf("授权失败: %s %s", tr.Error, tr.ErrorDesc)
			}
		}
	}
	if msToken == "" {
		return nil, fmt.Errorf("等待授权超时")
	}

	return exchangeForAccount(msToken, verbose)
}

// exchangeForAccount MSA token -> XBL -> XSTS -> Minecraft
func exchangeForAccount(msToken string, verbose bool) (*Account, error) {
	xblBody := `{"Properties":{"AuthMethod":"RPS","SiteName":"user.auth.xboxlive.com","RpsTicket":"d=` + msToken + `"},"RelyingParty":"http://auth.xboxlive.com","TokenType":"JWT"}`
	b, code, err := httpPostJSON("https://user.auth.xboxlive.com/user/authenticate", xblBody)
	if err != nil {
		return nil, fmt.Errorf("Xbox Live 认证失败: %w", err)
	}
	if code != 200 {
		return nil, fmt.Errorf("Xbox Live 认证失败: HTTP %d %s", code, truncate(string(b), 200))
	}
	var xbl struct {
		Token string `json:"Token"`
	}
	json.Unmarshal(b, &xbl)
	if xbl.Token == "" {
		return nil, fmt.Errorf("Xbox Live 认证失败: 没有拿到 Token")
	}
	if verbose {
		logInfo("Xbox Live 认证成功")
	}

	xstsBody := `{"Properties":{"SandboxId":"RETAIL","UserTokens":["` + xbl.Token + `"]},"RelyingParty":"rp://api.minecraftservices.com/","TokenType":"JWT"}`
	b, code, err = httpPostJSON("https://xsts.auth.xboxlive.com/xsts/authorize", xstsBody)
	if err != nil {
		return nil, fmt.Errorf("XSTS 授权失败: %w", err)
	}
	if code != 200 {
		return nil, fmt.Errorf("XSTS 授权失败: HTTP %d %s", code, truncate(string(b), 200))
	}
	var xsts struct {
		Token         string `json:"Token"`
		DisplayClaims struct {
			XUI []struct {
				UHS string `json:"uhs"`
			} `json:"xui"`
		} `json:"DisplayClaims"`
	}
	json.Unmarshal(b, &xsts)
	if xsts.Token == "" || len(xsts.DisplayClaims.XUI) == 0 || xsts.DisplayClaims.XUI[0].UHS == "" {
		return nil, fmt.Errorf("XSTS 授权失败: 缺少 Token 或 uhs")
	}
	uhs := xsts.DisplayClaims.XUI[0].UHS
	if verbose {
		logInfo("XSTS 授权成功")
	}

	mcBody := `{"identityToken":"XBL3.0 x=` + uhs + `;` + xsts.Token + `"}`
	b, code, err = httpPostJSON("https://api.minecraftservices.com/authentication/login_with_xbox", mcBody)
	if err != nil {
		return nil, fmt.Errorf("Minecraft 登录失败: %w", err)
	}
	if code != 200 {
		return nil, fmt.Errorf("Minecraft 登录失败: HTTP %d %s", code, truncate(string(b), 200))
	}
	var mc struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(b, &mc)
	if mc.AccessToken == "" {
		return nil, fmt.Errorf("Minecraft 登录失败: 没有拿到 access_token")
	}

	b, code, err = httpGet("https://api.minecraftservices.com/minecraft/profile", "Authorization", "Bearer "+mc.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("获取游戏档案失败: %w", err)
	}
	if code == 404 {
		return nil, fmt.Errorf("该账号没有购买 Minecraft")
	}
	if code != 200 {
		return nil, fmt.Errorf("获取游戏档案失败: HTTP %d", code)
	}
	var profile struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	}
	json.Unmarshal(b, &profile)
	if profile.Name == "" {
		return nil, fmt.Errorf("获取游戏档案失败: 没有返回用户名")
	}
	id := profile.ID
	if len(id) == 32 {
		id = id[:8] + "-" + id[8:12] + "-" + id[12:16] + "-" + id[16:20] + "-" + id[20:]
	}
	if verbose {
		logInfo("登录成功: " + profile.Name)
	}
	return &Account{
		Type:        "microsoft",
		Name:        profile.Name,
		UUID:        id,
		AccessToken: mc.AccessToken,
	}, nil
}

func loginAuthlib(server, username, password string) (*Account, error) {
	if server == "" || username == "" {
		return nil, fmt.Errorf("服务器与用户名不能为空")
	}
	server = strings.TrimRight(server, "/")
	body := `{"agent":{"name":"Minecraft","version":1},"username":"` + jsonEscape(username) +
		`","password":"` + jsonEscape(password) + `","requestUser":true}`
	b, code, err := httpPostJSON(server+"/authserver/authenticate", body)
	if err != nil {
		return nil, fmt.Errorf("Authlib 认证失败: %w", err)
	}
	if code != 200 {
		return nil, fmt.Errorf("Authlib 认证失败: HTTP %d %s", code, truncate(string(b), 200))
	}
	var resp struct {
		AccessToken    string `json:"accessToken"`
		ErrorMessage   string `json:"errorMessage"`
		SelectedProfile struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"selectedProfile"`
	}
	json.Unmarshal(b, &resp)
	if resp.AccessToken == "" {
		if resp.ErrorMessage != "" {
			return nil, fmt.Errorf("Authlib 认证失败: %s", resp.ErrorMessage)
		}
		return nil, fmt.Errorf("Authlib 认证失败: 没有拿到 accessToken")
	}
	name := resp.SelectedProfile.Name
	if name == "" {
		name = username
	}
	id := resp.SelectedProfile.ID
	if len(id) == 32 {
		id = id[:8] + "-" + id[8:12] + "-" + id[12:16] + "-" + id[16:20] + "-" + id[20:]
	}
	return &Account{Type: "authlib", Name: name, UUID: id, AccessToken: resp.AccessToken, Server: server}, nil
}

func jsonEscape(s string) string {
	b, _ := json.Marshal(s)
	// 去掉两端引号
	return strings.Trim(string(b), `"`)
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
