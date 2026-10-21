#pragma once
#include <string>
#include <vector>
#include <map>
#include "logger.hpp"

namespace mc {

// 服务器信息
struct ServerInfo {
    std::string address;
    std::string name;
    std::string icon;
    bool recent = false;
    int lastPlayed = 0;
};

// 存档信息
struct SaveInfo {
    std::string name;
    std::string path;
    int lastPlayed = 0;
    int playTime = 0;
    int worldTier = 0;
    std::string dimension;
    bool hardMode = false;
};

// 设置管理器
class SettingsManager {
public:
    SettingsManager();
    ~SettingsManager() = default;

    // 获取设置
    std::string get(const std::string& key, const std::string& defaultVal = "") const;
    int getInt(const std::string& key, int defaultVal = 0) const;
    bool getBool(const std::string& key, bool defaultVal = false) const;

    // 设置
    void set(const std::string& key, const std::string& value);
    void setInt(const std::string& key, int value);
    void setBool(const std::string& key, bool value);

    // 保存
    bool save() const;

private:
    std::map<std::string, std::string> settings_;
    std::string settingsPath_;
    bool load();
};

// 存档管理器
class SaveManager {
public:
    SaveManager();
    ~SaveManager() = default;

    // 列出存档
    std::vector<SaveInfo> listSaves(const std::string& version) const;

    // 重命名存档
    bool renameSave(const std::string& version, const std::string& oldName, 
                    const std::string& newName);

    // 删除存档
    bool deleteSave(const std::string& version, const std::string& saveName);

    // 备份存档
    bool backupSave(const std::string& version, const std::string& saveName);

private:
    std::string getSaveDir(const std::string& version) const;
};

// 服务器列表
class ServerList {
public:
    ServerList();
    ~ServerList() = default;

    // 添加服务器
    bool addServer(const std::string& address, const std::string& name = "");

    // 删除服务器
    bool removeServer(const std::string& address);

    // 获取服务器
    std::optional<ServerInfo> getServer(const std::string& address) const;

    // 列出所有服务器
    std::vector<ServerInfo> listServers() const;

    // 保存
    bool save() const;

private:
    std::vector<ServerInfo> servers_;
    std::string serversPath_;
    bool load();
};

} // namespace mc
