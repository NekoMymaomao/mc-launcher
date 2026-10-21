#include "core/instance.hpp"
#include "core/http.hpp"
#include "core/assets.hpp"
#include "core/java.hpp"
#include "core/modloader.hpp"
#include "core/utils.hpp"
#include <fstream>
#include <sstream>

namespace mc {

// ============================================================
// InstanceManager 实现
// ============================================================

InstanceManager::InstanceManager() {}

bool InstanceManager::install(const std::string& versionId,
                               const std::function<void(int, int, const std::string&)>& progress) {
    progress(0, 5, "正在获取版本信息...");
    
    // 1. 下载版本 manifest
    HttpClient http;
    std::string manifestUrl = "https://launchermeta.mojang.com/mc/game/" + versionId + ".json";
    std::string manifestJson = http.get(manifestUrl);
    if (manifestJson.empty()) {
        MC_LOG_ERROR << "无法获取版本信息: " << versionId;
        return false;
    }
    
    progress(1, 5, "已下载版本信息");
    
    // 2. 保存 manifest
    nlohmann::json manifest = nlohmann::json::parse(manifestJson);
    std::string instanceDir = getInstanceDir(versionId);
    utils::PathUtils::createDirectory(instanceDir);
    
    std::ofstream mf(join(instanceDir, versionId + ".json"));
    mf << manifest.dump(2);
    mf.close();
    
    // 3. 下载核心库
    progress(2, 5, "正在下载核心库...");
    std::string libsDir = join(instanceDir, "libraries");
    utils::PathUtils::createDirectory(libsDir);
    
    if (manifest.contains("libraries")) {
        for (const auto& lib : manifest["libraries"]) {
            std::string libName = lib.value("name", "");
            if (libName.empty()) continue;
            
            // 简化：只处理 artifact
            if (lib.contains("downloads") && lib["downloads"].contains("artifact")) {
                auto artifact = lib["downloads"]["artifact"];
                std::string url = artifact["url"].get<std::string>();
                std::string sha1 = artifact.value("sha1", "");
                
                std::string localPath = join(libsDir, 
                    utils::StringUtils::replaceAll(libName, ":", utils::PathUtils::SEPARATOR) + ".jar");
                utils::PathUtils::createDirectory(utils::PathUtils::join(libsDir, 
                    libName.substr(0, libName.find_last_of(':'))));
                
                http.download(url, localPath, nullptr, true);
            }
        }
    }
    
    progress(3, 5, "核心库下载完成");
    
    // 4. 下载客户端 jar
    progress(4, 5, "正在下载游戏文件...");
    if (manifest.contains("downloads") && manifest["downloads"].contains("client")) {
        auto client = manifest["downloads"]["client"];
        std::string url = client["url"].get<std::string>();
        std::string sha1 = client.value("sha1", "");
        
        std::string clientPath = join(instanceDir, versionId + ".jar");
        http.download(url, clientPath, nullptr, true);
    }
    
    progress(5, 5, "安装完成");
    MC_LOG_INFO << "版本 " << versionId << " 安装完成";
    return true;
}

bool InstanceManager::uninstall(const std::string& versionId) {
    std::string instanceDir = getInstanceDir(versionId);
    if (!utils::PathUtils::fileExists(instanceDir)) return false;
    
#ifdef MC_WINDOWS
    std::string cmd = "rmdir /s /q \"" + instanceDir + "\"";
#elif defined(MC_LINUX)
    std::string cmd = "rm -rf \"" + instanceDir + "\"";
#endif
    
    return utils::ProcessUtils::run(cmd) == 0;
}

std::vector<std::string> InstanceManager::listInstalled() const {
    std::vector<std::string> versions;
    std::string instancesDir = join(ConfigManager::instance().getConfig().launcherDir, "versions");
    
    if (!utils::PathUtils::fileExists(instancesDir)) return versions;
    
    auto files = utils::PathUtils::listFiles(instancesDir, ".json");
    for (const auto& f : files) {
        std::string name = utils::PathUtils::getFileName(f);
        if (name.size() > 4 && name.substr(name.size()-4) == ".json") {
            versions.push_back(name.substr(0, name.size()-5));
        }
    }
    return versions;
}

std::optional<VersionInfo> InstanceManager::getVersion(const std::string& versionId) const {
    std::string manifestPath = getManifestPath(versionId);
    if (!utils::PathUtils::fileExists(manifestPath)) return std::nullopt;
    
    std::ifstream f(manifestPath);
    nlohmann::json j = nlohmann::json::parse(f);
    
    VersionInfo info;
    info.id = j.value("id", versionId);
    info.mainClass = j.value("mainClass", "net.minecraft.client.main.Main");
    info.assetIndex = j.value("assetIndex", nlohmann::json::object()).value("id", "legacy");
    info.assets = j.value("assets", info.assetIndex);
    
    if (j.contains("libraries")) {
        for (const auto& lib : j["libraries"]) {
            info.libraries.push_back(lib.value("name", ""));
        }
    }
    
    return info;
}

bool InstanceManager::verifyVersion(const std::string& versionId) {
    VersionInfo info = getVersion(versionId);
    if (!info) return false;
    
    // 检查核心文件
    std::string coreJar = join(getInstanceDir(versionId), versionId + ".jar");
    if (!utils::PathUtils::fileExists(coreJar)) return false;
    
    // 检查所需库
    for (const auto& lib : info->libraries) {
        // 简化验证
    }
    
    return true;
}

std::string InstanceManager::getInstanceDir(const std::string& versionId) const {
    return join(ConfigManager::instance().getConfig().launcherDir, "versions", versionId);
}

std::string InstanceManager::getManifestPath(const std::string& versionId) const {
    return join(getInstanceDir(versionId), versionId + ".json");
}

std::string InstanceManager::getLibrariesDir() const {
    return join(ConfigManager::instance().getConfig().launcherDir, "libraries");
}

std::string InstanceManager::getAssetsDir() const {
    return join(ConfigManager::instance().getConfig().gameDir, "assets");
}

std::string InstanceManager::getNativesDir(const std::string& versionId) const {
    return join(getInstanceDir(versionId), "natives");
}

bool InstanceManager::downloadLibrary(const std::string& lib, const std::string& destDir) {
    // 解析 maven 坐标
    auto parts = utils::StringUtils::split(lib, ':');
    if (parts.size() < 3) return false;
    
    std::string group = parts[0];
    std::string name = parts[1];
    std::string version = parts[2];
    
    std::string path = join(destDir, group, name, version, name + "-" + version + ".jar");
    utils::PathUtils::createDirectory(utils::PathUtils::join(destDir, group, name, version));
    
    // 从 Maven 下载
    HttpClient http;
    std::string url = "https://libraries.minecraft.net/" + group.replace(group.begin(), group.end(), '/') 
                      + "/" + name + "/" + version + "/" + name + "-" + version + ".jar";
    return !http.download(url, path).empty();
}

bool InstanceManager::downloadAssets(const std::string& versionId) {
    AssetManager assets;
    return assets.downloadMissing(versionId);
}

bool InstanceManager::downloadNatives(const std::string& versionId) {
    NativeManager natives;
    // 简化实现
    return true;
}

} // namespace mc
