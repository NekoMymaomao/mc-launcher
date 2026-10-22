//go:build !cshared


// auth.go — 认证系统
import (
	"crypto/md5"
	"encoding/json"
	"fmt"
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
	if name == "" { name = "Player" }
	return &Account{Type: "offline", Name: name, UUID: offlineUuid(name), AccessToken: strings.ReplaceAll(offlineUuid(name), "-", "")}
}
func loginMicrosoft(verbose bool, clientId string) (*Account, error) {
	if clientId == "" { clientId = os.Getenv("MC_MS_CLIENT_ID"); if clientId == "" { clientId = "00000000402b5328" } }
	body := `{"client_id":"` + clientId + `","scope":"XboxLive.signin offline_access"}`
	b, code, err := httpPostJSON("https://login.microsoftonline.com/consumers/oauth2/v2.0/devicecode", body)
	if err != nil || code != 200 { return nil, fmt.Errorf("") }
	var dc struct{ DeviceCode, UserCode, VerificationUri string; Interval int }
	json.Unmarshal(b, &dc)
	if dc.DeviceCode == "" { return nil, fmt.Errorf("") }
	if verbose { fmt.Printf(": %s\n: %s\n", dc.VerificationUri, dc.UserCode); fmt.Println("...") }
	tokenBody := `{"grant_type":"urn:ietf:params:oauth:grant-type:device_code","client_id":"` + clientId + `","device_code":"` + dc.DeviceCode + `"}`
	var msToken string
	for i := 0; i < 180; i++ {
		time.Sleep(time.Duration(dc.Interval)*time.Second)
		b, code, _ := httpPostJSON("https://login.microsoftonline.com/consumers/oauth2/v2.0/token", tokenBody)
		if code == 200 {
			var tj map[string]any; json.Unmarshal(b, &tj)
			if t, ok := tj["access_token"].(string); ok && t != "" { msToken = t; break }
		}
	}
	if msToken == "" { return nil, fmt.Errorf("") }
	if verbose { logInfo(" Microsoft ") }
	xblBody := `{"Properties":{"AuthMethod":"RPS","SiteName":"user.auth.xboxlive.com","RpsTicket":"d=` + msToken + `"},"RelyingParty":"http://auth.xboxlive.com","TokenType":"JWT"}`
	b, _, _ = httpPostJSON("https://user.auth.xboxlive.com/user/authenticate", xblBody)
	var xbl map[string]any; json.Unmarshal(b, &xbl)
	xblToken, _ := xbl["Token"].(string)
	if xblToken == "" { return nil, fmt.Errorf("Xbox Live ") }
	if verbose { logInfo(" Xbox Live ") }
	xstsBody := `{"Properties":{"SandboxId":"RETAIL","UserTokens":["` + xblToken + `"]},"RelyingParty":"rp://api.minecraftservices.com/","TokenType":"JWT"}`
	b, _, _ = httpPostJSON("https://xsts.auth.xboxlive.com/xsts/authorize", xstsBody)
	var sx map[string]any; json.Unmarshal(b, &sx)
	xstsTok, _ := sx["Token"].(string)
	uhs := ""
	if dc2, ok := sx["DisplayClaims"]; ok {
		if xui, ok := dc2.([]any); ok && len(xui) > 0 {
			if m, ok := xui[0].(map[string]any); ok { uhs, _ = m["uhs"].(string) }
		}
	}
	if xstsTok == "" || uhs == "" { return nil, fmt.Errorf("XSTS ") }
	if verbose { logInfo(" XSTS ") }
	mcBody := `{"identityToken":"XBL3.0 x=` + uhs + `;` + xstsTok + `"}`
	b, _, _ = httpPostJSON("https://api.minecraftservices.com/authentication/login_with_xbox", mcBody)
	var mj map[string]any; json.Unmarshal(b, &mj)
	mcToken, _ := mj["access_token"].(string)
	if mcToken == "" { return nil, fmt.Errorf("Minecraft ") }
	b, _, _ = httpGet("https://api.minecraftservices.com/minecraft/profile", "Authorization", "Bearer "+mcToken)
	var profile map[string]any; json.Unmarshal(b, &profile)
	name2, _ := profile["name"].(string)
	id2, _ := profile["id"].(string)
	if name2 == "" { name2 = "MicrosoftPlayer" }
	if id2 == "" { id2 = offlineUuid("MicrosoftPlayer") }
	if len(id2) == 32 { id2 = id2[:8] + "-" + id2[8:12] + "-" + id2[12:16] + "-" + id2[16:20] + "-" + id2[20:] }
	if verbose { logInfo(": " + name2) }
	return &Account{Type: "microsoft", Name: name2, UUID: id2, AccessToken: mcToken}, nil
}
func loginAuthlib(server, user, pass string) (*Account, error) {
	if server == "" || user == "" { return nil, fmt.Errorf("") }
	server = strings.TrimRight(server, "/")
	body := `{"agent":{"name":"Minecraft","version":1},"username":"` + user + `","password":"` + pass + `","requestUser":true}`
	b, code, err := httpPostJSON(server+"/authserver/authenticate", body)
	if err != nil || code != 200 { return nil, fmt.Errorf(": %v", err) }
	var mj map[string]any; json.Unmarshal(b, &mj)
	token, _ := mj["accessToken"].(string)
	if token == "" { msg, _ := mj["errorMessage"].(string); return nil, fmt.Errorf(": %s", msg) }
	selected, _ := mj["selectedProfile"].(map[string]any)
	uname, _ := selected["name"].(string)
	uid, _ := selected["id"].(string)
	if uname == "" { uname = user }
	if len(uid) == 32 { uid = uid[:8] + "-" + uid[8:12] + "-" + uid[12:16] + "-" + uid[16:20] + "-" + uid[20:] }
	return &Account{Type: "authlib", Name: uname, UUID: uid, AccessToken: token, Server: server}, nil
}
