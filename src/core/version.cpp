#include "core/version.hpp"
#include "core/http.hpp"
#include <nlohmann/json.hpp>
#include <fstream>
#include <sstream>
#include <algorithm>

namespace mc {

// ============================================================
// ConfigManager 实现
// ============================================================

bool ConfigManager::load() {
    configPath_ = join(utils::PathUtils::getLauncherDir(), "settings.json");
    
    std::ifstream file(configPath_);
    if (!file.is_open()) return false;
    
    try {
        nlohmann::json j = nlohmann::json::parse(file);
        
        config_.gameDir = j.value("gameDir", "");
        config_.launcherDir = j.value("launcherDir", "");
        config_.javaPath = j.value("javaPath", "");
        config_.maxMemory = j.value("maxMemory", 4096);
        config_.minMemory = j.value("minMemory", 1024);
        config_.nativeAssets = j.value("nativeAssets", false);
        config_.autoJavaInstall = j.value("autoJavaInstall", true);
        
        if (j.contains("jvmArgs")) {
            for (const auto& arg : j["jvmArgs"]) {
                config_.jvmArgs.push_back(arg.get<std::string>());
            }
        }
        if (j.contains("gameArgs")) {
            for (const auto& arg : j["gameArgs"]) {
                config_.gameArgs.push_back(arg.get<std::string>());
            }
        }
        if (j.contains("authTokens")) {
            for (const auto& [key, val] : j["authTokens"].items()) {
                config_.authTokens[key] = val.get<std::string>();
            }
        }
    } catch (...) {
        return false;
    }
    
    return true;
}

bool ConfigManager::save() {
    if (config_.launcherDir.empty()) {
        config_.launcherDir = utils::PathUtils::getLauncherDir();
    }
    if (config_.gameDir.empty()) {
        config_.gameDir = join(config_.launcherDir, ".minecraft");
    }
    
    utils::PathUtils::createDirectory(config_.launcherDir);
    utils::PathUtils::createDirectory(config_.gameDir);
    
    nlohmann::json j;
    j["gameDir"] = config_.gameDir;
    j["launcherDir"] = config_.launcherDir;
    j["javaPath"] = config_.javaPath;
    j["maxMemory"] = config_.maxMemory;
    j["minMemory"] = config_.minMemory;
    j["nativeAssets"] = config_.nativeAssets;
    j["autoJavaInstall"] = config_.autoJavaInstall;
    j["jvmArgs"] = config_.jvmArgs;
    j["gameArgs"] = config_.gameArgs;
    j["authTokens"] = config_.authTokens;
    
    std::ofstream file(configPath_);
    file << j.dump(2);
    
    return file.good();
}

} // namespace mc
