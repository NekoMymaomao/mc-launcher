#pragma once
#include <string>
#include <vector>
#include <map>
#include <optional>
#include "logger.hpp"

namespace mc {

// 版本信息
struct VersionInfo {
    std::string id;
    std::string mainClass;
    std::string assetIndex;
    std::string assets;
    std::vector<std::string> libraries;
    std::vector<std::string> arguments;
    int releaseTime = 0;
    int installTime = 0;
    std::string javaVersion = ">=17";
    bool isSnapshot = false;
};

// 版本 manifest（从网络获取）
struct VersionManifest {
    std::vector<std::pair<std::string, VersionInfo>> versions;
    std::string latestRelease;
    std::string latestSnapshot;
};

// 资产索引
struct AssetIndex {
    std::string id;
    std::map<std::string, std::pair<std::string, int>> objects; // hash -> (size)
};

// 原生库
struct NativeLib {
    std::string name;
    std::string classifier;
    std::string url;
    std::string sha1;
    std::string extractDir;
};

// 配置文件
struct Config {
    std::string gameDir;
    std::string launcherDir;
    std::string javaPath;
    int maxMemory = 4096;
    int minMemory = 1024;
    std::vector<std::string> jvmArgs;
    std::vector<std::string> gameArgs;
    bool nativeAssets = false;
    bool autoJavaInstall = true;
    std::map<std::string, std::string> authTokens;
};

class ConfigManager {
public:
    static ConfigManager& instance() {
        static ConfigManager inst;
        return inst;
    }

    Config& getConfig() { return config_; }
    const Config& getConfig() const { return config_; }
    
    bool save();
    bool load();

private:
    ConfigManager() = default;
    Config config_;
    std::string configPath_;
};

} // namespace mc
