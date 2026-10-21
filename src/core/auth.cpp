#include "core/auth.hpp"
#include <fstream>
#include <sstream>
#include <iostream>
#include <thread>
#include <chrono>

namespace mc {

AuthManager::AuthManager() { loadSessions(); }

void AuthManager::loadSessions() {
    std::string dir = utils::PathUtils::getLauncherDir() + "/sessions";
    utils::PathUtils::createDirectory(dir);
    auto files = utils::PathUtils::listFiles(dir, ".json");
    for (auto& f : files) {
        std::ifstream in(f);
        if (!in) continue;
        std::stringstream ss;
        ss << in.rdbuf();
        std::string content = ss.str();
        std::string type = utils::PathUtils::getFileName(f).replace(
            utils::PathUtils::getFileName(f).size() - 5, 5, "");
        // 简化：只记录存在
    }
}

void AuthManager::saveSessions() const {
    // 简化实现
}

std::string AuthManager::getSessionFile(const std::string& type) const {
    return utils::PathUtils::getLauncherDir() + "/sessions/" + type + ".json";
}

std::optional<AuthSession> AuthManager::loginOffline(const std::string& username) {
    AuthSession s;
    s.userName = username.empty() ? "Player" : username;
    s.UUID = utils::Uuid::generateV4();
    s.UUIDClean = utils::Uuid::stripDashes(s.UUID);
    s.accessToken = s.UUIDClean;
    s.tokenType = "Offline";
    s.selected = true;
    sessions_.push_back(s);
    saveSessions();
    return s;
}

std::optional<AuthSession> AuthManager::loginMicrosoft(ProgressCallback progress) {
    // 设备码登录流程
    HttpClient http;
    progress("正在获取 Microsoft 登录链接...");

    auto resp = http.post("https://login.microsoftonline.com/consumers/oauth2/v2.0/devicecode",
        R"({"client_id":"5b8cf6a0-6f0a-4d51-9f8a-7b5c8e8c8f8a","scope":"XboxLive.signin offline_access"})");

    if (resp.empty()) return std::nullopt;

    auto j = utils::Json::parse(resp);
    std::string userCode = j["user_code"].getString();
    std::string deviceCode = j["device_code"].getString();
    std::string uri = j["verification_uri"].getString();

    std::cout << "请访问: " << uri << std::endl;
    std::cout << "输入验证码: " << userCode << std::endl;
    progress("请在浏览器中登录");

    for (int i = 0; i < 24; i++) {
        std::this_thread::sleep_for(std::chrono::seconds(5));
        auto tokenResp = http.post("https://login.microsoftonline.com/consumers/oauth2/v2.0/token",
            R"({"grant_type":"device_code","client_id":"5b8cf6a0-6f0a-4d51-9f8a-7b5c8e8c8f8a","device_code":"")"+deviceCode+R"("})");
        if (tokenResp.empty()) continue;
        auto tj = utils::Json::parse(tokenResp);
        if (tj["access_token"].isString()) {
            progress("登录成功!");
            AuthSession s;
            s.accessToken = tj["access_token"].getString();
            s.tokenType = "Microsoft";
            s.selected = true;
            sessions_.push_back(s);
            saveSessions();
            return s;
        }
    }
    return std::nullopt;
}

std::optional<AuthSession> AuthManager::loginAuthlib(const std::string& server,
                                                      const std::string& username,
                                                      const std::string& password) {
    HttpClient http;
    auto resp = http.post(server + "/authserver/authenticate",
        R"({"agent":{"name":"Minecraft","version":1},"username":"")"+username+
        R"(","password":"")"+password+"\"})");
    if (resp.empty()) return std::nullopt;
    auto j = utils::Json::parse(resp);
    if (!j["accessToken"].isString()) return std::nullopt;

    AuthSession s;
    s.accessToken = j["accessToken"].getString();
    s.userName = j["username"].isString() ? j["username"].getString() : username;
    s.UUID = j["id"].isString() ? j["id"].getString() : utils::Uuid::generateV4();
    s.UUIDClean = utils::Uuid::stripDashes(s.UUID);
    s.tokenType = "Authlib:" + server;
    s.selected = true;
    sessions_.push_back(s);
    saveSessions();
    return s;
}

std::vector<std::string> AuthManager::listAccounts() const {
    std::vector<std::string> result;
    for (auto& s : sessions_)
        result.push_back(s.tokenType + ": " + s.userName);
    return result;
}

bool AuthManager::removeAccount(const std::string& type) {
    sessions_.erase(
        std::remove_if(sessions_.begin(), sessions_.end(),
            [&type](const AuthSession& s){ return s.tokenType == type; }),
        sessions_.end());
    saveSessions();
    return true;
}

} // namespace mc
