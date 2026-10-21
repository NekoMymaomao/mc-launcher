#pragma once
#include <string>
#include <vector>
#include <optional>
#include "logger.hpp"

namespace mc {

// 认证会话
struct AuthSession {
    std::string accessToken;
    std::string userId;
    std::string userName;
    std::string UUID; // 带横杠
    std::string UUIDClean; // 不带横杠
    std::string tokenType = "Microsoft";
    bool selected = false;
    int expiresAt = 0;
};

// 认证管理器
class AuthManager {
public:
    using ProgressCallback = std::function<void(const std::string&)>;

    AuthManager();
    ~AuthManager() = default;

    // 离线登录
    std::optional<AuthSession> loginOffline(const std::string& username);

    // Microsoft 设备码登录
    std::optional<AuthSession> loginMicrosoft(ProgressCallback progress = nullptr);

    // Authlib Injector 认证
    std::optional<AuthSession> loginAuthlib(const std::string& server, 
                                             const std::string& username,
                                             const std::string& password);

    // 刷新 token
    bool refreshToken(AuthSession& session);

    // 保存 session
    bool saveSession(const std::string& accountType, const AuthSession& session);

    // 加载 session
    std::optional<AuthSession> loadSession(const std::string& accountType);

    // 列出所有账号
    std::vector<std::string> listAccounts() const;

    // 删除账号
    bool removeAccount(const std::string& accountType);

private:
    std::map<std::string, AuthSession> sessions_;
    std::string getSessionPath(const std::string& accountType) const;
};

} // namespace mc
