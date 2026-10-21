#pragma once
#include <string>
#include <vector>
#include "utils.hpp"

namespace mc {

struct ModInfo {
    std::string name;
    std::string filePath;
    bool enabled = true;
};

enum class ModLoaderType { NONE, FORGE, FABRIC, QUILT, NEOFORGE };

class ModManager {
public:
    bool addMod(const std::string& version, const std::string& modPath);
    bool removeMod(const std::string& version, const std::string& modName);
    std::vector<ModInfo> listMods(const std::string& version) const;
    ModLoaderType getLoaderType(const std::string& version) const;
    bool installLoader(const std::string& version, ModLoaderType type, const std::string& versionId);

private:
    std::string getModsDir(const std::string& version) const;
    std::string getModConfigPath(const std::string& version) const;
    mutable std::vector<ModInfo> mods_;
    mutable ModLoaderType loaderType_ = ModLoaderType::NONE;
};

} // namespace mc
