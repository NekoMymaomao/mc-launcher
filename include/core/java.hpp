#pragma once
#include <string>
#include <vector>
#include <optional>
#include <memory>
#include "logger.hpp"

namespace mc {

// Java 信息
struct JavaInfo {
    std::string path;
    std::string version;
    int majorVersion = 0;
    std::string arch;
    bool isValid = false;
};

// Java 管理器
class JavaManager {
public:
    JavaManager();
    ~JavaManager() = default;

    // 查找 Java
    std::optional<JavaInfo> findJava(const std::string& minVersion = "17") const;
    
    // 列出所有 Java
    std::vector<JavaInfo> listJava() const;

    // 安装 Java (Temurin)
    bool installJava(const std::string& version = "17");

    // 设置 Java 路径
    bool setJavaPath(const std::string& path);

private:
    std::string getJavaHome() const;
    std::string getJavaFromRegistry() const;
    std::string getJavaFromPath() const;
    bool verifyJava(const std::string& path, int minVersion) const;
    std::string downloadUrl(const std::string& version) const;
    bool extractArchive(const std::string& zipPath, const std::string& destDir) const;
};

} // namespace mc
