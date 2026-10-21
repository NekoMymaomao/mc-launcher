#pragma once
#include <string>
#include <vector>
#include <optional>
#include <curl/curl.h>
#include "logger.hpp"

namespace mc {

// JVM管理器
class JVMManager {
public:
    struct JVMInfo {
        std::string path;
        std::string version;
        std::string arch;
    };

    JVMManager();
    ~JVMManager() = default;

    // 查找JVM
    std::optional<JVMInfo> findJVM();
    
    // 验证JVM
    bool verifyJVM(const JVMInfo& info);
    
    // 构建启动参数
    std::vector<std::string> buildArgs(const std::string& main_class,
                                       const std::string& classpath,
                                       const std::string& username,
                                       const std::string& uuid,
                                       int width, int height);
    
    // 启动
    int launch(const JVMInfo& jvm, const std::string& classpath,
               const std::string& username, const std::string& uuid,
               int width, int height);

private:
    // 平台特定实现
    std::string findJavaExecutable() const;
    std::string findJavaFromEnv() const;
    std::string findJavaFromRegistry() const;
    std::string extractJavaVersion(const std::string& path) const;
};

} // namespace mc
