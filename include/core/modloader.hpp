#pragma once
#include <string>
#include <vector>
#include "logger.hpp"

namespace mc {

// Mod 加载器类型
enum class ModLoaderType { NONE, FORGE, FABRIC, QUILT, NEOFORGE, LITELOADER, OPTIFINE };

// Mod 信息
struct ModInfo {
    std::string name;
    std::string version;
    std::string filePath;
    bool enabled = true;
    std::vector<std::string> dependencies;
};

// Mod 加载器基类
class ModLoader {
public:
    virtual ~ModLoader() = default;
    virtual ModLoaderType getType() const = 0;
    virtual bool isInstalled(const std::string& loaderPath) const = 0;
    virtual bool install(const std::string& version, const std::string& loaderVersion) = 0;
    virtual std::string getGlobalLibsDir() const = 0;
    virtual std::string getLoaderJar(const std::string& loaderPath) const = 0;
};

// Forge 加载器
class ForgeLoader : public ModLoader {
public:
    ModLoaderType getType() const override { return ModLoaderType::FORGE; }
    bool isInstalled(const std::string& loaderPath) const override;
    bool install(const std::string& version, const std::string& loaderVersion) override;
    std::string getGlobalLibsDir() const override;
    std::string getLoaderJar(const std::string& loaderPath) const override;
};

// Fabric 加载器
class FabricLoader : public ModLoader {
public:
    ModLoaderType getType() const override { return ModLoaderType::FABRIC; }
    bool isInstalled(const std::string& loaderPath) const override;
    bool install(const std::string& version, const std::string& loaderVersion) override;
    std::string getGlobalLibsDir() const override;
    std::string getLoaderJar(const std::string& loaderPath) const override;
};

// Quilt 加载器
class QuiltLoader : public ModLoader {
public:
    ModLoaderType getType() const override { return ModLoaderType::QUILT; }
    bool isInstalled(const std::string& loaderPath) const override;
    bool install(const std::string& version, const std::string& loaderVersion) override;
    std::string getGlobalLibsDir() const override;
    std::string getLoaderJar(const std::string& loaderPath) const override;
};

// NeoForge 加载器
class NeoForgeLoader : public ModLoader {
public:
    ModLoaderType getType() const override { return ModLoaderType::NEOFORGE; }
    bool isInstalled(const std::string& loaderPath) const override;
    bool install(const std::string& version, const std::string& loaderVersion) override;
    std::string getGlobalLibsDir() const override;
    std::string getLoaderJar(const std::string& loaderPath) const override;
};

// Mod 管理器
class ModManager {
public:
    ModManager();
    ~ModManager() = default;

    // 添加模组
    bool addMod(const std::string& version, const std::string& modPath);

    // 删除模组
    bool removeMod(const std::string& version, const std::string& modName);

    // 列出模组
    std::vector<ModInfo> listMods(const std::string& version) const;

    // 启用/禁用模组
    bool setModEnabled(const std::string& version, const std::string& modName, bool enabled);

    // 获取加载器类型
    ModLoaderType getLoaderType(const std::string& version) const;

    // 安装加载器
    bool installLoader(const std::string& version, ModLoaderType type, const std::string& versionId);

private:
    std::string getModsDir(const std::string& version) const;
    std::string getModConfigPath(const std::string& version) const;
    bool loadModConfig(const std::string& version);
    bool saveModConfig(const std::string& version);
    
    std::map<std::string, std::vector<ModInfo>> modCache_;
    std::map<std::string, ModLoaderType> loaderTypes_;
};

} // namespace mc
