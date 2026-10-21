#pragma once
#include <string>
#include <vector>
#include <optional>
#include "http.hpp"
#include "utils.hpp"

namespace mc {

struct AuthSession {
    std::string accessToken;
    std::string userId;
    std::string userName;
    std::string UUID;
    std::string UUIDClean;
    std::string tokenType = "Offline";
    bool selected = false;
};

class AuthManager {
public:
    using ProgressCallback = std::function<void(const std::string&)>;

    AuthManager();
    std::optional<AuthSession> loginOffline(const std::string& username);
    std::optional<AuthSession> loginMicrosoft(ProgressCallback progress = nullptr);
    std::optional<AuthSession> loginAuthlib(const std::string& server,
                                             const std::string& username,
                                             const std::string& password);
    std::vector<std::string> listAccounts() const;
    bool removeAccount(const std::string& type);

private:
    std::vector<AuthSession> sessions_;
    void loadSessions();
    void saveSessions() const;
    std::string getSessionFile(const std::string& type) const;
};

} // namespace mc
