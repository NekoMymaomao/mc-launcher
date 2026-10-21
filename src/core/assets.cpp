#include "core/assets.hpp"
#include "core/http.hpp"
#include "core/utils.hpp"
#include <fstream>
#include <sstream>

namespace mc {

// ============================================================
// AssetManager 实现
// ============================================================

AssetManager::AssetManager() {}

bool AssetManager::downloadIndex(const std::string& version) {
    std::string indexUrl = "https://launchermeta.mojang.com/mc/assets/" + version + "/index.json";
    std::string destPath = join(getAssetDir(), "indexes", version + ".json");
    
    HttpClient http;
    std::string content = http.get(indexUrl);
    if (content.empty()) return false;
    
    utils::PathUtils::createDirectory(utils::PathUtils::join(getAssetDir(), "indexes"));
    
    std::ofstream f(destPath);
    f << content;
    return f.good();
}

bool AssetManager::downloadMissing(const std::string& version) {
    std::string indexPath = join(getAssetDir(), "indexes", version + ".json");
    if (!utils::PathUtils::fileExists(indexPath)) {
        if (!downloadIndex(version)) return false;
    }
    
    std::ifstream f(indexPath);
    nlohmann::json index = nlohmann::json::parse(f);
    
    int total = index["objects"].size();
    int downloaded = 0;
    
    HttpClient http;
    for (const auto& [hash, info] : index["objects"].items()) {
        std::string expectedSize = info.value("size", "0").get<std::string>();
        std::string expectedHash = hash;
        
        std::string objPath = getObjectPath(hash);
        if (utils::PathUtils::fileExists(objPath)) {
            downloaded++;
            continue;
        }
        
        std::string url = "https://resources.download.minecraft.net/" + 
                         hash.substr(0, 2) + "/" + hash;
        
        utils::PathUtils::createDirectory(utils::PathUtils::join(getAssetDir(), "objects", 
                                                                  hash.substr(0, 2)));
        
        http.download(url, objPath, [&](int64_t d, int64_t t, double s) {
            downloaded++;
            // 可以显示进度
        }, false);
    }
    
    return true;
}

bool AssetManager::createVirtualResources(const std::string& version, const std::string& destDir) {
    std::string indexPath = join(getAssetDir(), "indexes", version + ".json");
    if (!utils::PathUtils::fileExists(indexPath)) return false;
    
    std::ifstream f(indexPath);
    nlohmann::json index = nlohmann::json::parse(f);
    
    // 创建虚拟资源目录结构
    for (const auto& [assetId, info] : index["objects"].items()) {
        std::string virtualPath = join(destDir, "assets", "virtual", assetId);
        std::string realPath = getObjectPath(assetId);
        
        if (utils::PathUtils::fileExists(realPath)) {
            utils::PathUtils::createDirectory(utils::PathUtils::join(destDir, "assets", "virtual"));
            // 创建符号链接或复制
            // 简化：只记录映射
        }
    }
    
    return true;
}

bool AssetManager::verifyAssets(const std::string& version) {
    std::string indexPath = join(getAssetDir(), "indexes", version + ".json");
    if (!utils::PathUtils::fileExists(indexPath)) return false;
    
    std::ifstream f(indexPath);
    nlohmann::json index = nlohmann::json::parse(f);
    
    int missing = 0;
    for (const auto& [hash, info] : index["objects"].items()) {
        std::string objPath = getObjectPath(hash);
        if (!utils::PathUtils::fileExists(objPath)) {
            missing++;
        }
    }
    
    MC_LOG_INFO << "资产验证完成: " << missing << " 个缺失";
    return missing == 0;
}

std::string AssetManager::getAssetDir() const {
    return join(ConfigManager::instance().getConfig().gameDir, "assets");
}

std::string AssetManager::getObjectPath(const std::string& hash) const {
    return join(getAssetDir(), "objects", hash.substr(0, 2), hash);
}

bool AssetManager::downloadObject(const std::string& url, const std::string& dest,
                                   int64_t expectedSize, const std::string& expectedHash) {
    HttpClient http;
    std::string downloaded = http.download(url, dest);
    
    if (downloaded.empty()) return false;
    
    // 验证完整性
    std::string actualHash = utils::Sha1::compute(
        []{ std::ifstream f(dest, std::ios::binary); 
           return std::string((std::istreambuf_iterator<char>(f)), 
                             std::istreambuf_iterator<char>()); }()
    );
    
    return actualHash == expectedHash;
}

// ============================================================
// NativeManager 实现
// ============================================================

NativeManager::NativeManager() {}

bool NativeManager::extractNatives(const std::string& jarPath, const std::string& destDir) {
    if (!utils::PathUtils::fileExists(jarPath)) return false;
    
    utils::PathUtils::createDirectory(destDir);
    
#ifdef MINIZ_H
    return utils::ZipUtils::extract(jarPath, destDir);
#else
    // fallback to system unzip
    std::string cmd = "unzip -o -q \"" + jarPath + "\" -d \"" + destDir + "\"";
    return utils::ProcessUtils::run(cmd) == 0;
#endif
}

bool NativeManager::cleanupNatives(const std::string& nativesDir) {
    if (!utils::PathUtils::fileExists(nativesDir)) return true;
    
    auto files = utils::PathUtils::listFiles(nativesDir);
    // 可以根据版本过滤
    return true;
}

bool NativeManager::isNativeJar(const std::string& libName) const {
    return libName.find(":natives-") != std::string::npos ||
           libName.find(":classifier=\"natives") != std::string::npos;
}

std::string NativeManager::getClassifier(const std::string& libName) const {
    // 提取 classifier
    size_t pos = libName.find("natives-");
    if (pos != std::string::npos) {
        size_t end = libName.find('"', pos);
        return libName.substr(pos + 8, end - pos - 8);
    }
    return "unknown";
}

} // namespace mc
