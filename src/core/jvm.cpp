#include "core/jvm.hpp"
#include "core/utils.hpp"
#include <cstdio>
#include <cstdlib>
#include <sstream>
#include <regex>
#include <iostream>

#ifdef MC_WINDOWS
    #include <windows.h>
    #include <shellapi.h>
#elif defined(MC_LINUX)
    #include <unistd.h>
    #include <sys/wait.h>
#endif

namespace mc {

JVMManager::JVMManager() {}

std::optional<JVMManager::JVMInfo> JVMManager::findJVM() {
    // 1. 环境变量 JAVA_HOME
    auto env_jvm = findJavaFromEnv();
    if (!env_jvm.empty()) {
        JVMInfo info;
        info.path = env_jvm;
        info.version = extractJavaVersion(env_jvm);
        info.arch = "64";
        if (verifyJVM(info)) return info;
    }
    
    // 2. 平台特定搜索
#ifdef MC_WINDOWS
    auto reg_jvm = findJavaFromRegistry();
    if (!reg_jvm.empty()) {
        JVMInfo info;
        info.path = reg_jvm;
        info.version = extractJavaVersion(reg_jvm);
        info.arch = "64";
        if (verifyJVM(info)) return info;
    }
#elif defined(MC_LINUX)
    // Linux: 搜索常见路径
    std::vector<std::string> search_paths = {
        "/usr/lib/jvm/java-17-openjdk/bin/java",
        "/usr/lib/jvm/java-11-openjdk/bin/java",
        "/usr/bin/java"
    };
    
    for (const auto& path : search_paths) {
        if (utils::PathUtils::fileExists(path)) {
            JVMInfo info;
            info.path = path;
            info.version = extractJavaVersion(path);
            info.arch = "64";
            if (verifyJVM(info)) return info;
        }
    }
#endif
    
    return std::nullopt;
}

bool JVMManager::verifyJVM(const JVMInfo& info) {
#ifdef MC_WINDOWS
    // Windows: 尝试运行 java -version
    STARTUPINFOA si = {};
    PROCESS_INFORMATION pi = {};
    si.cb = sizeof(si);
    
    std::string cmd = "\"" + info.path + "\" -version";
    
    if (CreateProcessA(nullptr, const_cast<char*>(cmd.c_str()),
                       nullptr, nullptr, FALSE, CREATE_NO_WINDOW,
                       nullptr, nullptr, &si, &pi)) {
        WaitForSingleObject(pi.hProcess, 5000);
        DWORD exitCode;
        GetExitCodeProcess(pi.hProcess, &exitCode);
        CloseHandle(pi.hProcess);
        CloseHandle(pi.hThread);
        return exitCode == 0;
    }
#elif defined(MC_LINUX)
    // Linux: 使用system()
    std::string cmd = "\"" + info.path + "\" -version 2>&1";
    int ret = system(cmd.c_str());
    return WIFEXITED(ret) && WEXITSTATUS(ret) == 0;
#endif
    return false;
}

std::vector<std::string> JVMManager::buildArgs(const std::string& main_class,
                                                const std::string& classpath,
                                                const std::string& username,
                                                const std::string& uuid,
                                                int width, int height) {
    std::vector<std::string> args;
    
    // JVM参数
    args.push_back("-Xmx1G");
    args.push_back("-Xms512M");
    args.push_back("-XX:+UseAdaptiveSizePolicy");
    args.push_back("-XX:+UnlockExperimentalVMOptions");
    args.push_back("-XX:+G1NewTypeAlignment");
    args.push_back("-XX:G1HeapWastePercent=5");
    args.push_back("-XX:G1MixedGCCountTarget=4");
    args.push_back("-XX:InitiatingHeapOccupancyPercent=20");
    args.push_back("-XX:G1MixedGCLiveThresholdPercent=90");
    args.push_back("-XX:G1RSetUpdatingPauseTimePercent=25");
    args.push_back("-XX:SurvivorRatio=32");
    args.push_back("-XX:+PerfDisableSharedMem");
    args.push_back("-XX:MaxTenuringThreshold=1");
    
    // Minecraft参数
    args.push_back("-Dfml.ignoreInvalidMinecraftCertificates=true");
    args.push_back("-Dfml.ignorePatchDiscrepancies=true");
    
    // 主类和类路径
    args.push_back("-cp");
    args.push_back(classpath);
    args.push_back(main_class);
    
    // 游戏参数
    args.push_back("--username");
    args.push_back(username);
    args.push_back("--uuid");
    args.push_back(uuid);
    args.push_back("--width");
    args.push_back(std::to_string(width));
    args.push_back("--height");
    args.push_back(std::to_string(height));
    
    return args;
}

int JVMManager::launch(const JVMInfo& jvm, const std::string& classpath,
                       const std::string& username, const std::string& uuid,
                       int width, int height) {
    auto args = buildArgs(jvm.path, classpath, username, uuid, width, height);
    
    MC_LOG_INFO << "Launching JVM: " << jvm.path;
    
#ifdef MC_WINDOWS
    // Windows: 使用CreateProcess
    std::string cmd = "\"" + jvm.path + "\"";
    for (const auto& arg : args) {
        cmd += " \"" + arg + "\"";
    }
    
    STARTUPINFOA si = {};
    PROCESS_INFORMATION pi = {};
    si.cb = sizeof(si);
    
    if (!CreateProcessA(nullptr, const_cast<char*>(cmd.c_str()),
                        nullptr, nullptr, FALSE,
                        0, nullptr, nullptr, &si, &pi)) {
        MC_LOG_ERROR << "Failed to create process";
        return -1;
    }
    
    WaitForSingleObject(pi.hProcess, INFINITE);
    
    DWORD exitCode;
    GetExitCodeProcess(pi.hProcess, &exitCode);
    
    CloseHandle(pi.hProcess);
    CloseHandle(pi.hThread);
    
    return (int)exitCode;
    
#elif defined(MC_LINUX)
    // Linux: 使用fork/exec
    pid_t pid = fork();
    if (pid < 0) {
        MC_LOG_ERROR << "Failed to fork";
        return -1;
    }
    
    if (pid == 0) {
        // 子进程
        std::vector<char*> argv;
        argv.push_back(const_cast<char*>(jvm.path.c_str()));
        for (const auto& arg : args) {
            argv.push_back(const_cast<char*>(arg.c_str()));
        }
        argv.push_back(nullptr);
        
        execv(jvm.path.c_str(), argv.data());
        _exit(127);
    }
    
    // 父进程
    int status;
    waitpid(pid, &status, 0);
    
    if (WIFEXITED(status)) {
        return WEXITSTATUS(status);
    }
    return -1;
#endif
    
    return 0;
}

std::string JVMManager::findJavaExecutable() const {
#ifdef MC_WINDOWS
    return findJavaFromRegistry();
#else
    return findJavaFromEnv();
#endif
}

std::string JVMManager::findJavaFromEnv() const {
    const char* java_home = getenv("JAVA_HOME");
    if (java_home) {
#ifdef MC_WINDOWS
        return std::string(java_home) + "\\bin\\java.exe";
#else
        return std::string(java_home) + "/bin/java";
#endif
    }
    
    const char* path = getenv("PATH");
    if (path) {
#ifdef MC_WINDOWS
        // Windows PATH搜索
        std::string path_str = path;
        size_t start = 0;
        size_t pos;
        while ((pos = path_str.find(';', start)) != std::string::npos) {
            std::string dir = path_str.substr(start, pos - start);
            std::string java_path = dir + "\\java.exe";
            if (utils::PathUtils::fileExists(java_path)) {
                return java_path;
            }
            start = pos + 1;
        }
#else
        // Linux PATH搜索
        std::stringstream ss(path);
        std::string dir;
        while (std::getline(ss, dir, ':')) {
            std::string java_path = dir + "/java";
            if (utils::PathUtils::fileExists(java_path)) {
                return java_path;
            }
        }
#endif
    }
    
    return "";
}

#ifdef MC_WINDOWS
std::string JVMManager::findJavaFromRegistry() const {
    HKEY hKey;
    char path[MAX_PATH];
    DWORD type, size = MAX_PATH;
    
    // 尝试64位Java
    if (RegOpenKeyExA(HKEY_LOCAL_MACHINE, 
                      "SOFTWARE\\JavaSoft\\Java Runtime Environment",
                      0, KEY_READ, &hKey) == ERROR_SUCCESS) {
        if (RegQueryValueExA(hKey, "JavaHome", nullptr, &type, (LPBYTE)path, &size) == ERROR_SUCCESS) {
            RegCloseKey(hKey);
            return std::string(path) + "\\bin\\java.exe";
        }
        RegCloseKey(hKey);
    }
    
    // 尝试32位Java
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

std::string JVMManager::extractJavaVersion(const std::string& path) const {
    // 简化实现：从路径提取
    // 实际应该解析 java -version 输出
    return "unknown";
}

} // namespace mc
