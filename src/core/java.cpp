#include "core/java.hpp"
#include "core/http.hpp"
#include <sstream>
#include <iostream>

namespace mc {

JavaManager::JavaManager() {}

std::string JavaManager::getJavaHome() const {
    const char* h = getenv("JAVA_HOME");
    return h ? h : "";
}

std::string JavaManager::getJavaFromPath() const {
#ifdef MC_WINDOWS
    std::string out;
    utils::ProcessUtils::run("where java", [&](const std::string& line){ out = line; });
    size_t p = out.find("\\bin\\java.exe");
    return (p != std::string::npos) ? out.substr(0, p + 4) : "";
#else
    std::string out;
    utils::ProcessUtils::run("which java", [&](const std::string& line){ out = line; });
    return utils::StringUtils::trim(out);
#endif
}

bool JavaManager::verifyJava(const std::string& path, int minVersion) const {
    if (!utils::PathUtils::fileExists(path)) return false;
    std::string ver = utils::ProcessUtils::getVersion(path);
    if (ver.empty()) return false;
    try { return std::stoi(ver) >= minVersion; } catch (...) { return false; }
}

std::optional<JavaInfo> JavaManager::findJava(const std::string& minVersion) const {
    int minVer = 17;
    try { minVer = std::stoi(minVersion); } catch (...) {}

    auto check = [&](const std::string& path) -> std::optional<JavaInfo> {
        if (!verifyJava(path, minVer)) return std::nullopt;
        JavaInfo info;
        info.path = path;
        info.version = utils::ProcessUtils::getVersion(path);
        info.majorVersion = info.version.empty() ? 0 : std::stoi(info.version);
        info.isValid = true;
        return info;
    };

    if (!javaPath_.empty()) {
        auto r = check(javaPath_);
        if (r) return r;
    }
    std::string home = getJavaHome();
    if (!home.empty()) {
#ifdef MC_WINDOWS
        std::string p = home + "\\bin\\java.exe";
#else
        std::string p = home + "/bin/java";
#endif
        auto r = check(p);
        if (r) return r;
    }
    std::string path = getJavaFromPath();
    if (!path.empty()) {
        auto r = check(path);
        if (r) return r;
    }
    return std::nullopt;
}

std::vector<JavaInfo> JavaManager::listJava() const {
    std::vector<JavaInfo> result;
    auto check = [&](const std::string& path) {
        if (!utils::PathUtils::fileExists(path)) return;
        JavaInfo info;
        info.path = path;
        info.version = utils::ProcessUtils::getVersion(path);
        info.majorVersion = info.version.empty() ? 0 : std::stoi(info.version);
        info.isValid = !info.version.empty();
        result.push_back(info);
    };
    if (!javaPath_.empty()) check(javaPath_);
    std::string home = getJavaHome();
    if (!home.empty()) {
#ifdef MC_WINDOWS
        check(home + "\\bin\\java.exe");
#else
        check(home + "/bin/java");
#endif
    }
    std::string path = getJavaFromPath();
    if (!path.empty()) check(path);
    return result;
}

bool JavaManager::setJavaPath(const std::string& path) {
    javaPath_ = path;
    return true;
}

std::string JavaManager::downloadUrl(const std::string& version) const {
    std::string os = "linux";
#ifdef MC_WINDOWS
    os = "windows";
#endif
    return "https://api.adoptium.net/v3/binary/latest/" + version + "/ga/" +
           os + "/x64/jre/hotspot/normal/eclipse?project=jdk";
}

bool JavaManager::installJava(const std::string& version) {
    HttpClient http;
    std::string url = downloadUrl(version);
    std::string dest = utils::PathUtils::getLauncherDir() + "/java/temurin-" + version + ".zip";
    std::string dir = utils::PathUtils::getLauncherDir() + "/java/temurin-" + version;
    utils::PathUtils::createDirectory(utils::PathUtils::join(utils::PathUtils::getLauncherDir(), "java"));

    std::cout << "正在下载 Java " << version << "..." << std::endl;
    auto result = http.download(url, dest);
    if (result.empty()) return false;

    std::cout << "正在解压..." << std::endl;
    // 简化：使用系统 unzip
#ifdef MC_WINDOWS
    std::string cmd = "tar -xzf \"" + dest + "\" -C \"" + dir + "\"";
#else
    std::string cmd = "unzip -q \"" + dest + "\" -d \"" + dir + "\"";
#endif
    int ret = utils::ProcessUtils::run(cmd);
    remove(dest.c_str());
    if (ret != 0) return false;

    setJavaPath(dir + (utils::PathUtils::SEP == '\\' ? "\\bin\\java.exe" : "/bin/java"));
    std::cout << "✓ Java " << version << " 安装完成" << std::endl;
    return true;
}

} // namespace mc
