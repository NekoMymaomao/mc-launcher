#pragma once
#include <string>
#include <vector>
#include <optional>
#include "logger.hpp"
#include "version.hpp"

namespace mc {

// 实例管理器
class InstanceManager {
public:
    InstanceManager();
    ~InstanceManager() = default;

    // 安装版本（完整流程：下载版本+依赖+资源+原生）
    bool install(const std::string& versionId, 
                 const std::function<void(int, int, const std::string&)>& progress = nullptr);

    // 卸载版本
    bool uninstall(const std::string& versionId);

    // 列出已安装版本
    std::vector<std::string> listInstalled() const;

    // 获取版本信息
    std::optional<VersionInfo> getVersion(const std::string& versionId) const;

    // 检查版本完整性
    bool verifyVersion(const std::string& versionId);

private:
    std::string getInstanceDir(const std::string& versionId) const;
    std::string getManifestPath(const std::string& versionId) const;
    std::string getLibrariesDir() const;
    std::string getAssetsDir() const;
    std::string getNativesDir(const std::string& versionId) const;
    
    bool downloadLibrary(const std::string& lib, const std::string& destDir);
    bool downloadAssets(const std::string& versionId);
    bool downloadNatives(const std::string& versionId);
};

} // namespace mc
