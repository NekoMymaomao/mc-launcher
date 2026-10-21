#include "core/settings.hpp"
#include "core/utils.hpp"
#include <fstream>

namespace mc {

// ============================================================
// SettingsManager 实现
// ============================================================

SettingsManager::SettingsManager() {
    settingsPath_ = join(utils::PathUtils::getLauncherDir(), "launcher_settings.json");
    load();
}

std::string SettingsManager::get(const std::string& key, const std::string& defaultVal) const {
    auto it = settings_.find(key);
    return it != settings_.end() ? it->second : defaultVal;
}

int SettingsManager::getInt(const std::string& key, int defaultVal) const {
    auto it = settings_.find(key);
    if (it == settings_.end()) return defaultVal;
    try {
        return std::stoi(it->second);
    } catch (...) {
        return defaultVal;
    }
}

bool SettingsManager::getBool(const std::string& key, bool defaultVal) const {
    auto it = settings_.find(key);
    if (it == settings_.end()) return defaultVal;
    return it->second == "true" || it->second == "1";
}

void SettingsManager::set(const std::string& key, const std::string& value) {
    settings_[key] = value;
}

void SettingsManager::setInt(const std::string& key, int value) {
    settings_[key] = std::to_string(value);
}

void SettingsManager::setBool(const std::string& key, bool value) {
    settings_[key] = value ? "true" : "false";
}

bool SettingsManager::save() const {
    nlohmann::json j;
    for (const auto& [key, val] : settings_) {
        j[key] = val;
    }
    
    std::ofstream f(settingsPath_);
    f << j.dump(2);
    return f.good();
}

bool SettingsManager::load() {
    if (!utils::PathUtils::fileExists(settingsPath_)) return false;
    
    std::ifstream f(settingsPath_);
    try {
        nlohmann::json j = nlohmann::json::parse(f);
        for (const auto& [key, val] : j.items()) {
            settings_[key] = val.get<std::string>();
        }
    } catch (...) {
        return false;
    }
    return true;
}

// ============================================================
// SaveManager 实现
// ============================================================

SaveManager::SaveManager() {}

std::vector<SaveInfo> SaveManager::listSaves(const std::string& version) const {
    std::vector<SaveInfo> saves;
    std::string saveDir = getSaveDir(version);
    
    if (!utils::PathUtils::fileExists(saveDir)) return saves;
    
    auto dirs = utils::PathUtils::listFiles(saveDir);
    for (const auto& dir : dirs) {
        if (utils::PathUtils::getFileName(dir).starts_with(".")) continue;
        
        SaveInfo info;
        info.name = utils::PathUtils::getFileName(dir);
        info.path = dir;
        
        // 尝试读取 level.dat 获取更多信息
        std::string levelDat = join(dir, "level.dat");
        if (utils::PathUtils::fileExists(levelDat)) {
            // 简化：只记录存在
            info.lastPlayed = 0;
            info.playTime = 0;
        }
        
        saves.push_back(info);
    }
    return saves;
}

bool SaveManager::renameSave(const std::string& version, const std::string& oldName,
                              const std::string& newName) {
    std::string oldPath = join(getSaveDir(version), oldName);
    std::string newPath = join(getSaveDir(version), newName);
    
    if (!utils::PathUtils::fileExists(oldPath)) return false;
    
#ifdef MC_WINDOWS
    return MoveFileExA(oldPath.c_str(), newPath.c_str(), MOVEFILE_REPLACE_EXISTING) != 0;
#else
    return rename(oldPath.c_str(), newPath.c_str()) == 0;
#endif
}

bool SaveManager::deleteSave(const std::string& version, const std::string& saveName) {
    std::string savePath = join(getSaveDir(version), saveName);
    if (!utils::PathUtils::fileExists(savePath)) return false;
    
#ifdef MC_WINDOWS
    return RemoveDirectoryA(savePath.c_str()) || 
           (std::system(("rmdir /s /q \"" + savePath + "\"").c_str()) == 0);
#else
    return std::system(("rm -rf \"" + savePath + "\"").c_str()) == 0;
#endif
}

bool SaveManager::backupSave(const std::string& version, const std::string& saveName) {
    std::string src = join(getSaveDir(version), saveName);
    std::string backupDir = join(getSaveDir(version), "..", "backups");
    std::string dest = join(backupDir, saveName + "_" + 
                           utils::TimeUtils::getTimestamp());
    
    utils::PathUtils::createDirectory(backupDir);
    
    // 复制目录
    std::string cmd = "cp -r \"" + src + "\" \"" + dest + "\"";
    return utils::ProcessUtils::run(cmd) == 0;
}

std::string SaveManager::getSaveDir(const std::string& version) const {
    return join(ConfigManager::instance().getConfig().gameDir, "saves");
}

// ============================================================
// ServerList 实现
// ============================================================

ServerList::ServerList() {
    serversPath_ = join(utils::PathUtils::getLauncherDir(), "servers.json");
    load();
}

bool ServerList::addServer(const std::string& address, const std::string& name) {
    // 检查是否已存在
    for (auto& s : servers_) {
        if (s.address == address) {
            s.name = name.empty() ? address : name;
            s.recent = true;
            s.lastPlayed = 0; // TODO: 实际时间
            return save();
        }
    }
    
    ServerInfo info;
    info.address = address;
    info.name = name.empty() ? address : name;
    servers_.push_back(info);
    return save();
}

bool ServerList::removeServer(const std::string& address) {
    servers_.erase(std::remove_if(servers_.begin(), servers_.end(),
        [&address](const ServerInfo& s) { return s.address == address; }),
        servers_.end());
    return save();
}

std::optional<ServerInfo> ServerList::getServer(const std::string& address) const {
    auto it = std::find_if(servers_.begin(), servers_.end(),
        [&address](const ServerInfo& s) { return s.address == address; });
    return it != servers_.end() ? std::make_optional(*it) : std::nullopt;
}

std::vector<ServerInfo> ServerList::listServers() const {
    return servers_;
}

bool ServerList::save() const {
    nlohmann::json j;
    for (const auto& s : servers_) {
        nlohmann::json server;
        server["address"] = s.address;
        server["name"] = s.name;
        server["icon"] = s.icon;
        server["recent"] = s.recent;
        server["lastPlayed"] = s.lastPlayed;
        j.push_back(server);
    }
    
    std::ofstream f(serversPath_);
    f << j.dump(2);
    return f.good();
}

bool ServerList::load() {
    if (!utils::PathUtils::fileExists(serversPath_)) return false;
    
    std::ifstream f(serversPath_);
    try {
        nlohmann::json j = nlohmann::json::parse(f);
        for (const auto& server : j) {
            ServerInfo info;
            info.address = server.value("address", "");
            info.name = server.value("name", info.address);
            info.icon = server.value("icon", "");
            info.recent = server.value("recent", false);
            info.lastPlayed = server.value("lastPlayed", 0);
            servers_.push_back(info);
        }
    } catch (...) {
        return false;
    }
    return true;
}

} // namespace mc
