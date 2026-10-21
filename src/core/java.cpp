#include "core/java.hpp"
#include "core/http.hpp"
#include "core/utils.hpp"
#include <sstream>

namespace mc {

// ============================================================
// JavaManager 实现
// ============================================================

JavaManager::JavaManager() {}

std::optional<JavaInfo> JavaManager::findJava(const std::string& minVersion) const {
    int minVer = 17;
    std::sscanf(minVersion.c_str(), "%d", &minVer);
    
    // 1. 检查 JAVA_HOME
    std::string javaHome = getJavaHome();
    if (!javaHome.empty()) {
        std::string javaPath = javaHome + (utils::PathUtils::SEPARATOR == '\\' ? "\\bin\\java.exe" : "/bin/java");
        if (verifyJava(javaPath, minVer)) {
            JavaInfo info;
            info.path = javaPath;
            info.version = ProcessUtils::getVersion(javaPath);
            info.majorVersion = std::stoi(info.version.empty() ? "0" : info.version);
            info.arch = "64";
            info.isValid = true;
            return info;
        }
    }
    
    // 2. 检查 PATH
    std::string pathJava = getJavaFromPath();
    if (!pathJava.empty() && verifyJava(pathJava, minVer)) {
        JavaInfo info;
        info.path = pathJava;
        info.version = ProcessUtils::getVersion(pathJava);
        info.majorVersion = std::stoi(info.version.empty() ? "0" : info.version);
        info.arch = "64";
        info.isValid = true;
        return info;
    }
    
#ifdef MC_WINDOWS
    // 3. Windows 注册表
    std::string regJava = getJavaFromRegistry();
    if (!regJava.empty() && verifyJava(regJava, minVer)) {
        JavaInfo info;
        info.path = regJava;
        info.version = ProcessUtils::getVersion(regJava);
        info.majorVersion = std::stoi(info.version.empty() ? "0" : info.version);
        info.arch = "64";
        info.isValid = true;
        return info;
    }
#endif
    
    return std::nullopt;
}

std::vector<JavaInfo> JavaManager::listJava() const {
    std::vector<JavaInfo> results;
    
    // 检查 JAVA_HOME
    std::string javaHome = getJavaHome();
    if (!javaHome.empty()) {
        std::string javaPath = javaHome + (utils::PathUtils::SEPARATOR == '\\' ? "\\bin\\java.exe" : "/bin/java");
        JavaInfo info;
        info.path = javaPath;
        info.version = ProcessUtils::getVersion(javaPath);
        info.majorVersion = info.version.empty() ? 0 : std::stoi(info.version);
        info.arch = "64";
        info.isValid = !info.version.empty();
        results.push_back(info);
    }
    
    // 检查 PATH
    std::string pathJava = getJavaFromPath();
    if (!pathJava.empty()) {
        JavaInfo info;
        info.path = pathJava;
        info.version = ProcessUtils::getVersion(pathJava);
        info.majorVersion = info.version.empty() ? 0 : std::stoi(info.version);
        info.arch = "64";
        info.isValid = true;
        results.push_back(info);
    }
    
    // Windows: 检查注册表
#ifdef MC_WINDOWS
    std::string regJava = getJavaFromRegistry();
    if (!regJava.empty()) {
        JavaInfo info;
        info.path = regJava;
        info.version = ProcessUtils::getVersion(regJava);
        info.majorVersion = info.version.empty() ? 0 : std::stoi(info.version);
        info.arch = "64";
        info.isValid = true;
        results.push_back(info);
    }
#endif
    
    return results;
}

bool JavaManager::installJava(const std::string& version) {
    HttpClient http;
    std::string url = downloadUrl(version);
    if (url.empty()) return false;
    
    std::string dest = join(utils::PathUtils::getLauncherDir(), "java", 
                            "temurin-" + version + ".zip");
    std::string installDir = join(utils::PathUtils::getLauncherDir(), "java", 
                                   "temurin-" + version);
    
    MC_LOG_INFO << "正在下载 Java " << version << "...";
    std::string downloaded = http.download(url, dest);
    if (downloaded.empty()) return false;
    
    MC_LOG_INFO << "正在解压...";
    bool ok = extractArchive(downloaded, installDir);
    
    // 清理
    remove(downloaded.c_str());
    
    if (ok) {
        setJavaPath(join(installDir, "bin", "java"));
        MC_LOG_INFO << "Java " << version << " 安装完成";
    }
    
    return ok;
}

bool JavaManager::setJavaPath(const std::string& path) {
    ConfigManager& cfg = ConfigManager::instance();
    cfg.getConfig().javaPath = path;
    return cfg.save();
}

std::string JavaManager::getJavaHome() const {
    return getenv("JAVA_HOME") ? getenv("JAVA_HOME") : "";
}

std::string JavaManager::getJavaFromPath() const {
#ifdef MC_WINDOWS
    // 使用 where 命令
    std::string output;
    int ret = utils::ProcessUtils::run("where java", [&](const std::string& line) {
        output = line;
    });
    if (ret == 0 && !output.empty()) {
        // 去掉 "java" 后缀，得到目录
        size_t pos = output.find("\\bin\\java.exe");
        if (pos != std::string::npos) {
            return output.substr(0, pos + 4); // 包含 \bin
        }
    }
#elif defined(MC_LINUX)
    // 使用 which
    std::string output;
    int ret = utils::ProcessUtils::run("which java", [&](const std::string& line) {
        output = line;
    });
    if (ret == 0 && !output.empty()) {
        return utils::StringUtils::trim(output);
    }
#endif
    return "";
}

#ifdef MC_WINDOWS
std::string JavaManager::getJavaFromRegistry() const {
    char path[MAX_PATH] = {0};
    DWORD type = 0;
    DWORD size = sizeof(path);
    
    HKEY hKey;
    if (RegOpenKeyExA(HKEY_LOCAL_MACHINE, 
                      "SOFTWARE\\JavaSoft\\Java Runtime Environment",
                      0, KEY_READ, &hKey) == ERROR_SUCCESS) {
        if (RegQueryValueExA(hKey, "JavaHome", nullptr, &type, (LPBYTE)path, &size) == ERROR_SUCCESS) {
            RegCloseKey(hKey);
            return std::string(path) + "\\bin\\java.exe";
        }
        RegCloseKey(hKey);
    }
    
    // 尝试 32 位
    if (RegOpenKeyExA(HKEY_LOCAL_MACHINE, 
                      "SOFTWARE\\WOW6432Node\\JavaSoft\\Java Runtime Environment",
                      0, KEY_READ, &hKey) == ERROR_SUCCESS) {
        if (RegQueryValueExA(hKey, "JavaHome", nullptr, &type, (LPBYTE)path, &size) == ERROR_SUCCESS) {
            RegCloseKey(hKey);
            return std::string(path) + "\\bin\\java.exe";
        }
        RegCloseKey(hKey);
    }
    
    return "";
}
#endif

bool JavaManager::verifyJava(const std::string& path, int minVersion) const {
    // 检查路径是否存在
    if (!utils::PathUtils::fileExists(path)) return false;
    
    // 检查版本
    std::string version = ProcessUtils::getVersion(path);
    if (version.empty()) return false;
    
    int major = std::stoi(version);
    return major >= minVersion;
}

std::string JavaManager::downloadUrl(const std::string& version) const {
    // Adoptium Temurin 下载链接
    std::string arch = "x64";
#ifdef MC_LINUX
    arch = "x64";
#endif
    
    std::string os = "linux";
#ifdef MC_WINDOWS
    os = "windows";
#endif
    
    return "https://api.adoptium.net/v3/binary/latest/" + version + "/ga/" + 
           os + "/" + arch + "/jre/hotspot/normal/eclipse?project=jdk";
}

bool JavaManager::extractArchive(const std::string& zipPath, const std::string& destDir) const {
#ifdef MINIZ_H
    return utils::ZipUtils::extract(zipPath, destDir);
#else
    // fallback to system unzip
    std::string cmd = "unzip -q \"" + zipPath + "\" -d \"" + destDir + "\"";
    return utils::ProcessUtils::run(cmd) == 0;
#endif
}

} // namespace mc
