#include "core/modloader.hpp"
#include "core/http.hpp"
#include "core/utils.hpp"
#include <nlohmann/json.hpp>
#include <fstream>

namespace mc {

// ============================================================
// ModLoader 实现
// ============================================================

// Forge
bool ForgeLoader::isInstalled(const std::string& loaderPath) const {
    return utils::PathUtils::fileExists(join(loaderPath, "forge-installer.jar"));
}

bool ForgeLoader::install(const std::string& version, const std::string& loaderVersion) {
    // Forge 安装比较复杂，需要运行 installer jar
    // 这里返回 true 表示支持
    MC_LOG_INFO << "Forge " << loaderVersion << " 安装需要手动运行 installer";
    return true;
}

std::string ForgeLoader::getGlobalLibsDir() const {
    return join(ConfigManager::instance().getConfig().launcherDir, "libraries", "forge");
}

std::string ForgeLoader::getLoaderJar(const std::string& loaderPath) const {
    return join(loaderPath, "forge-installer.jar");
}

// Fabric
bool FabricLoader::isInstalled(const std::string& loaderPath) const {
    return utils::PathUtils::fileExists(join(loaderPath, "fabric-loader.jar"));
}

bool FabricLoader::install(const std::string& version, const std::string& loaderVersion) {
    HttpClient http;
    std::string metaUrl = "https://meta.fabricmc.net/v2/versions/loader/" + version + "/" + 
                          loaderVersion + "/profile/json";
    
    std::string metaJson = http.get(metaUrl);
    if (metaJson.empty()) return false;
    
    auto meta = nlohmann::json::parse(metaJson);
    
    // 保存加载器配置
    std::string installDir = join(ConfigManager::instance().getConfig().launcherDir, 
                                   "loaders", "fabric", loaderVersion);
    utils::PathUtils::createDirectory(installDir);
    
    std::ofstream f(join(installDir, "profile.json"));
    f << meta.dump(2);
    
    // 下载主 jar
    if (meta.contains("installer") && meta["installer"].contains("url")) {
        std::string installerUrl = meta["installer"]["url"].get<std::string>();
        http.download(installerUrl, join(installDir, "fabric-loader.jar"));
    }
    
    return true;
}

std::string FabricLoader::getGlobalLibsDir() const {
    return join(ConfigManager::instance().getConfig().launcherDir, "libraries", "fabric");
}

std::string FabricLoader::getLoaderJar(const std::string& loaderPath) const {
    return join(loaderPath, "fabric-loader.jar");
}

// Quilt
bool QuiltLoader::isInstalled(const std::string& loaderPath) const {
    return utils::PathUtils::fileExists(join(loaderPath, "quilt-loader.jar"));
}

bool QuiltLoader::install(const std::string& version, const std::string& loaderVersion) {
    // Quilt 与 Fabric 类似，使用不同的 API
    HttpClient http;
    std::string metaUrl = "https://meta.quiltmc.org/v3/versions/loader/" + version + "/" + 
                          loaderVersion + "/profile/json";
    
    std::string metaJson = http.get(metaUrl);
    if (metaJson.empty()) return false;
    
    auto meta = nlohmann::json::parse(metaJson);
    
    std::string installDir = join(ConfigManager::instance().getConfig().launcherDir, 
                                   "loaders", "quilt", loaderVersion);
    utils::PathUtils::createDirectory(installDir);
    
    std::ofstream f(join(installDir, "profile.json"));
    f << meta.dump(2);
    
    return true;
}

std::string QuiltLoader::getGlobalLibsDir() const {
    return join(ConfigManager::instance().getConfig().launcherDir, "libraries", "quilt");
}

std::string QuiltLoader::getLoaderJar(const std::string& loaderPath) const {
    return join(loaderPath, "quilt-loader.jar");
}

// NeoForge
bool NeoForgeLoader::isInstalled(const std::string& loaderPath) const {
    return utils::PathUtils::fileExists(join(loaderPath, "neoforge-installer.jar"));
}

bool NeoForgeLoader::install(const std::string& version, const std::string& loaderVersion) {
    // NeoForge 类似于 Forge
    MC_LOG_INFO << "NeoForge " << loaderVersion << " 安装需要手动运行 installer";
    return true;
}

std::string NeoForgeLoader::getGlobalLibsDir() const {
    return join(ConfigManager::instance().getConfig().launcherDir, "libraries", "neoforge");
}

std::string NeoForgeLoader::getLoaderJar(const std::string& loaderPath) const {
    return join(loaderPath, "neoforge-installer.jar");
}

// ============================================================
// ModManager 实现
// ============================================================

ModManager::ModManager() {}

bool ModManager::addMod(const std::string& version, const std::string& modPath) {
    std::string modsDir = getModsDir(version);
    utils::PathUtils::createDirectory(modsDir);
    
    std::string fileName = utils::PathUtils::getFileName(modPath);
    std::string destPath = join(modsDir, fileName);
    
    // 复制文件
    std::ifstream src(modPath, std::ios::binary);
    std::ofstream dst(destPath, std::ios::binary);
    if (src && dst) {
        dst << src.rdbuf();
        src.close();
        dst.close();
        
        // 更新缓存
        ModInfo info;
        info.name = fileName;
        info.version = "";
        info.filePath = destPath;
        info.enabled = true;
        modCache_[version].push_back(info);
        
        saveModConfig(version);
        return true;
    }
    return false;
}

bool ModManager::removeMod(const std::string& version, const std::string& modName) {
    std::string modsDir = getModsDir(version);
    std::string modPath = join(modsDir, modName);
    
    if (utils::PathUtils::fileExists(modPath)) {
        remove(modPath.c_str());
        
        // 更新缓存
        auto& mods = modCache_[version];
        mods.erase(std::remove_if(mods.begin(), mods.end(),
            [&modName](const ModInfo& m) { return m.name == modName; }), mods.end());
        
        saveModConfig(version);
        return true;
    }
    return false;
}

std::vector<ModInfo> ModManager::listMods(const std::string& version) const {
    loadModConfig(version);
    return modCache_.count(version) ? modCache_.at(version) : std::vector<ModInfo>();
}

bool ModManager::setModEnabled(const std::string& version, const std::string& modName, bool enabled) {
    auto& mods = modCache_[version];
    for (auto& mod : mods) {
        if (mod.name == modName) {
            mod.enabled = enabled;
            break;
        }
    }
    return saveModConfig(version);
}

ModLoaderType ModManager::getLoaderType(const std::string& version) const {
    return loaderTypes_.count(version) ? loaderTypes_.at(version) : ModLoaderType::NONE;
}

bool ModManager::installLoader(const std::string& version, ModLoaderType type, 
                                const std::string& versionId) {
    std::unique_ptr<ModLoader> loader;
    switch (type) {
        case ModLoaderType::FORGE:
            loader = std::make_unique<ForgeLoader>();
            break;
        case ModLoaderType::FABRIC:
            loader = std::make_unique<FabricLoader>();
            break;
        case ModLoaderType::QUILT:
            loader = std::make_unique<QuiltLoader>();
            break;
        case ModLoaderType::NEOFORGE:
            loader = std::make_unique<NeoForgeLoader>();
            break;
        default:
            return false;
    }
    
    // 选择加载器版本
    std::string loaderVersion = "latest";
    bool ok = loader->install(versionId, loaderVersion);
    if (ok) {
        loaderTypes_[version] = type;
    }
    return ok;
}

std::string ModManager::getModsDir(const std::string& version) const {
    return join(ConfigManager::instance().getConfig().launcherDir, 
                "versions", version, "mods");
}

std::string ModManager::getModConfigPath(const std::string& version) const {
    return join(getModsDir(version), "mods.json");
}

bool ModManager::loadModConfig(const std::string& version) {
    std::string configPath = getModConfigPath(version);
    if (!utils::PathUtils::fileExists(configPath)) return false;
    
    std::ifstream f(configPath);
    nlohmann::json j = nlohmann::json::parse(f);
    
    modCache_[version].clear();
    for (const auto& mod : j) {
        ModInfo info;
        info.name = mod.value("name", "");
        info.version = mod.value("version", "");
        info.filePath = mod.value("filePath", "");
        info.enabled = mod.value("enabled", true);
        modCache_[version].push_back(info);
    }
    return true;
}

bool ModManager::saveModConfig(const std::string& version) {
    nlohmann::json j;
    for (const auto& mod : modCache_[version]) {
        nlohmann::json m;
        m["name"] = mod.name;
        m["version"] = mod.version;
        m["filePath"] = mod.filePath;
        m["enabled"] = mod.enabled;
        j.push_back(m);
    }
    
    std::string configPath = getModConfigPath(version);
    std::ofstream f(configPath);
    f << j.dump(2);
    return f.good();
}

} // namespace mc
