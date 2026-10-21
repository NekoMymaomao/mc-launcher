#include "core/auth.hpp"
#include "core/http.hpp"
#include "core/utils.hpp"
#include <nlohmann/json.hpp>
#include <iomanip>
#include <regex>

namespace mc {

// ============================================================
// AuthManager 实现
// ============================================================

AuthManager::AuthManager() {
    // 加载保存的 session
    ConfigManager& cfg = ConfigManager::instance();
    const auto& tokens = cfg.getConfig().authTokens;
    for (const auto& [type, token] : tokens) {
        // 简单恢复，实际应该解析完整的 session
        if (type == "offline") {
            AuthSession session;
            session.tokenType = "Offline";
            session.accessToken = token;
            session.selected = true;
            sessions_["offline"] = session;
        } else if (type == "microsoft") {
            AuthSession session;
            session.tokenType = "Microsoft";
            session.accessToken = token;
            session.selected = true;
            sessions_["microsoft"] = session;
        } else if (type.substr(0, 7) == "authlib") {
            AuthSession session;
            session.tokenType = type;
            session.accessToken = token;
            session.selected = true;
            sessions_[type] = session;
        }
    }
}

std::optional<AuthSession> AuthManager::loginOffline(const std::string& username) {
    AuthSession session;
    session.userName = username;
    session.UUID = utils::Uuid::generateV4();
    session.UUIDClean = utils::Uuid::stripDashes(session.UUID);
    session.accessToken = session.UUIDClean;
    session.tokenType = "Offline";
    session.selected = true;
    
    saveSession("offline", session);
    return session;
}

std::optional<AuthSession> AuthManager::loginMicrosoft(ProgressCallback progress) {
    HttpClient http;
    
    // 1. 获取设备码
    progress("正在获取 Microsoft 登录链接...");
    std::string deviceCodeUrl = "https://login.microsoftonline.com/consumers/oauth2/v2.0/devicecode";
    
    auto resp = http.post(deviceCodeUrl, R"({
        "client_id": "5b8cf6a0-6f0a-4d51-9f8a-7b5c8e8c8f8a",
        "scope": "XboxLive.signin offline_access"
    })");
    
    if (resp.empty()) {
        MC_LOG_ERROR << "Failed to get device code";
        return std::nullopt;
    }
    
    try {
        auto j = nlohmann::json::parse(resp);
        std::string userCode = j["user_code"].get<std::string>();
        std::string deviceCode = j["device_code"].get<std::string>();
        std::string verificationUri = j["verification_uri"].get<std::string>();
        
        MC_LOG_INFO << "请在浏览器中访问: " << verificationUri;
        MC_LOG_INFO << "输入验证码: " << userCode;
        progress("请在浏览器中登录 Microsoft 账户");
        
        // 2. 轮询获取 token
        std::string tokenUrl = "https://login.microsoftonline.com/consumers/oauth2/v2.0/token";
        for (int i = 0; i < 30; i++) {
            std::this_thread::sleep_for(std::chrono::seconds(5));
            
            auto tokenResp = http.post(tokenUrl, R"({
                "grant_type": "device_code",
                "client_id": "5b8cf6a0-6f0a-4d51-9f8a-7b5c8e8c8f8a",
                "device_code": ")" + deviceCode + "\""
            })");
            
            if (tokenResp.empty()) continue;
            
            try {
                auto tj = nlohmann::json::parse(tokenResp);
                if (tj.contains("access_token")) {
                    progress("登录成功！");
                    
                    // 3. 获取 Xbox token
                    auto xboxResp = http.post(
                        "https://user.auth.xboxlive.com/user/authenticate",
                        R"({"Properties":{"AuthMethod":"RPS","SiteName":"user.auth.xboxlive.com","RpsTicket":"d=" + j["access_token"]}})"
                    );
                    
                    // 4. 获取 Minecraft token
                    auto mcResp = http.post(
                        "https://api.minecraftservices.com/authentication/login_with_xbox",
                        R"({"identityToken": ")" + xboxResp + R"("})"
                    );
                    
                    if (mcResp.empty()) return std::nullopt;
                    
                    auto mj = nlohmann::json::parse(mcResp);
                    if (!mj.contains("access_token")) return std::nullopt;
                    
                    AuthSession session;
                    session.accessToken = mj["access_token"].get<std::string>();
                    session.tokenType = "Microsoft";
                    session.selected = true;
                    
                    // 解析用户信息
                    if (mj.contains("user_info")) {
                        auto ui = mj["user_info"].get<nlohmann::json>();
                        session.userId = ui.value("子", "");
                        session.userName = ui.value("邮件地址", ui.value("名称", "Unknown"));
                        
                        // 生成 UUID（基于名称）
                        std::string nameHash = utils::Sha1::compute(session.userName);
                        session.UUID = nameHash.substr(0,8) + "-" + 
                                      nameHash.substr(8,4) + "-" +
                                      nameHash.substr(12,4) + "-" +
                                      nameHash.substr(16,4) + "-" +
                                      nameHash.substr(20);
                        session.UUIDClean = utils::Uuid::stripDashes(session.UUID);
                    }
                    
                    saveSession("microsoft", session);
                    return session;
                }
            } catch (...) {
                continue;
            }
        }
        
        progress("登录超时");
        return std::nullopt;
        
    } catch (const std::exception& e) {
        MC_LOG_ERROR << "Microsoft login failed: " << e.what();
        return std::nullopt;
    }
}

std::optional<AuthSession> AuthManager::loginAuthlib(const std::string& server,
                                                      const std::string& username,
                                                      const std::string& password) {
    HttpClient http;
    
    // Authlib Injector 认证端点
    std::string url = server + "/authserver/authenticate";
    
    auto resp = http.post(url, R"({
        "agent": {"name": "Minecraft", "version": 1},
        "username": ")" + username + R"(",
        "password": ")" + password + R"(",
        "requestUser": true
    })");
    
    if (resp.empty()) return std::nullopt;
    
    try {
        auto j = nlohmann::json::parse(resp);
        if (!j.contains("accessToken")) return std::nullopt;
        
        AuthSession session;
        session.accessToken = j["accessToken"].get<std::string>();
        session.userName = j.value("username", username);
        session.UUID = j.value("id", utils::Uuid::generateV4());
        session.UUIDClean = utils::Uuid::stripDashes(session.UUID);
        session.tokenType = "authlib:" + server;
        session.selected = true;
        
        saveSession(session.tokenType, session);
        return session;
        
    } catch (...) {
        return std::nullopt;
    }
}

bool AuthManager::saveSession(const std::string& accountType, const AuthSession& session) {
    sessions_[accountType] = session;
    
    ConfigManager& cfg = ConfigManager::instance();
    cfg.getConfig().authTokens[accountType] = session.accessToken;
    return cfg.save();
}

std::optional<AuthSession> AuthManager::loadSession(const std::string& accountType) {
    auto it = sessions_.find(accountType);
    if (it != sessions_.end()) {
        return it->second;
    }
    return std::nullopt;
}

std::vector<std::string> AuthManager::listAccounts() const {
    std::vector<std::string> accounts;
    for (const auto& [type, session] : sessions_) {
        accounts.push_back(type + ": " + session.userName);
    }
    return accounts;
}

bool AuthManager::removeAccount(const std::string& accountType) {
    sessions_.erase(accountType);
    ConfigManager::instance().getConfig().authTokens.erase(accountType);
    return ConfigManager::instance().save();
}

std::string AuthManager::getSessionPath(const std::string& accountType) const {
    return join(utils::PathUtils::getLauncherDir(), "sessions", accountType + ".json");
}

} // namespace mc
