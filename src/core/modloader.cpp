#include "core/modloader.hpp"
#include "core/http.hpp"
#include <fstream>
#include <iostream>

namespace mc {

bool ModManager::addMod(const std::string& version, const std::string& modPath) {
    std::string modsDir = getModsDir(version);
    utils::PathUtils::createDirectory(modsDir);

    std::string fileName = utils::PathUtils::getFileName(modPath);
    std::string dest = modsDir + "/" + fileName;

    std::ifstream src(modPath, std::ios::binary);
    std::ofstream dst(dest, std::ios::binary);
    if (!src || !dst) return false;
    dst << src.rdbuf();

    ModInfo info;
    info.name = fileName;
    info.filePath = dest;
    info.enabled = true;
    mods_.push_back(info);
    return true;
}

bool ModManager::removeMod(const std::string& version, const std::string& modName) {
    auto it = std::find_if(mods_.begin(), mods_.end(),
        [&modName](const ModInfo& m){ return m.name == modName; });
    if (it == mods_.end()) return false;

    std::remove(it->filePath.c_str());
    mods_.erase(it);
    return true;
}

std::vector<ModInfo> ModManager::listMods(const std::string& version) const {
    return mods_;
}

ModLoaderType ModManager::getLoaderType(const std::string& version) const {
    return loaderType_;
}

bool ModManager::installLoader(const std::string& version, ModLoaderType type,
                                const std::string& versionId) {
    // 简化实现
    loaderType_ = type;
    return true;
}

std::string ModManager::getModsDir(const std::string& version) const {
    return utils::PathUtils::join(utils::PathUtils::getLauncherDir(), "versions", version, "mods");
}

std::string ModManager::getModConfigPath(const std::string& version) const {
    return getModsDir(version) + "/mods.json";
}

} // namespace mc
