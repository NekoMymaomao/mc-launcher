#pragma once
#include <string>
#include <vector>
#include <functional>
#include "logger.hpp"

namespace mc {

// 资产管理器
class AssetManager {
public:
    AssetManager();
    ~AssetManager() = default;

    // 下载资产索引
    bool downloadIndex(const std::string& version);

    // 下载所有缺失的资产
    bool downloadMissing(const std::string& version);

    // 创建虚拟资源目录
    bool createVirtualResources(const std::string& version, const std::string& destDir);

    // 验证资产完整性
    bool verifyAssets(const std::string& version);

private:
    std::string getAssetDir() const;
    std::string getObjectPath(const std::string& hash) const;
    bool downloadObject(const std::string& url, const std::string& dest,
                        int64_t expectedSize, const std::string& expectedHash);
};

// 原生库管理器
class NativeManager {
public:
    NativeManager();
    ~NativeManager() = default;

    // 从 jar 中提取原生库
    bool extractNatives(const std::string& jarPath, const std::string& destDir);

    // 清理过期的原生文件
    bool cleanupNatives(const std::string& nativesDir);

private:
    bool isNativeJar(const std::string& libName) const;
    std::string getClassifier(const std::string& libName) const;
};

} // namespace mc
