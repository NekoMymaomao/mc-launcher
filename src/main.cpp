#include "core/config.hpp"
#include "core/jvm.hpp"
#include "core/downloader.hpp"
#include "core/utils.hpp"
#include <iostream>
#include <string>
#include <fstream>
#include <algorithm>

using namespace mc;

void printUsage(const char* prog) {
    std::cout << "MC Launcher - Cross-platform Minecraft Launcher\n\n"
              << "Usage: " << prog << " [OPTIONS]\n\n"
              << "Options:\n"
              << "  --version <id>      Select version (default: latest)\n"
              << "  --username <name>   Minecraft username (required)\n"
              << "  --uuid <uuid>       Player UUID\n"
              << "  --width <px>        Window width (default: 854)\n"
              << "  --height <px>       Window height (default: 480)\n"
              << "  --debug             Enable debug logging\n"
              << "  --help              Show this help\n";
}

int main(int argc, char* argv[]) {
    Logger::instance().setLevel(LogLevel::INFO);
    
    // 解析命令行参数
    std::string version_id;
    std::string username;
    std::string uuid;
    int width = 854;
    int height = 480;
    
    for (int i = 1; i < argc; ++i) {
        std::string arg = argv[i];
        if (arg == "--help" || arg == "-h") {
            printUsage(argv[0]);
            return 0;
        } else if (arg == "--debug") {
            Logger::instance().setLevel(LogLevel::DEBUG);
        } else if (arg == "--version" && i + 1 < argc) {
            version_id = argv[++i];
        } else if (arg == "--username" && i + 1 < argc) {
            username = argv[++i];
        } else if (arg == "--uuid" && i + 1 < argc) {
            uuid = argv[++i];
        } else if (arg == "--width" && i + 1 < argc) {
            width = std::stoi(argv[++i]);
        } else if (arg == "--height" && i + 1 < argc) {
            height = std::stoi(argv[++i]);
        }
    }
    
    if (username.empty()) {
        std::cerr << "Error: --username is required\n";
        printUsage(argv[0]);
        return 1;
    }
    
    // 获取游戏目录
    std::string game_dir = utils::PathUtils::getGameDir();
    utils::PathUtils::createDirectory(game_dir);
    
    MC_LOG_INFO << "Game directory: " << game_dir;
    MC_LOG_INFO << "Username: " << username;
    
    // 加载版本配置
    Config config;
    std::string version_file = game_dir + "/version.json";
    
    if (!config.loadVersionFile(version_file)) {
        // 尝试从网络加载
        MC_LOG_INFO << "Loading version from network...";
        Downloader downloader;
        std::string version_url = "https://piston-meta.mojang.com/v1/packages/latest.json";
        std::string json = downloader.downloadString(version_url);
        
        if (json.empty()) {
            std::cerr << "Error: Failed to load version data\n";
            return 1;
        }
        
        config.loadVersion(json);
        
        // 保存本地缓存
        std::ofstream out(version_file);
        out << json;
    }
    
    // 查找JVM
    JVMManager jvm_mgr;
    auto jvm = jvm_mgr.findJVM();
    if (!jvm.has_value()) {
        std::cerr << "\nError: Java not found!\n\n";
#ifdef MC_WINDOWS
        std::cerr << "Please install Java 17 or later:\n"
                  << "  1. Download from https://adoptium.net/\n"
                  << "  2. Set JAVA_HOME environment variable\n"
                  << "  3. Add %JAVA_HOME%\\bin to PATH\n\n";
#elif defined(MC_LINUX)
        std::cerr << "Please install Java 17 or later:\n"
                  << "  Ubuntu/Debian: sudo apt install openjdk-17-jdk\n"
                  << "  Fedora: sudo dnf install java-17-openjdk\n"
                  << "  Arch: sudo pacman -S jdk17-openjdk\n\n";
#endif
        return 1;
    }
    
    MC_LOG_INFO << "Using JVM: " << jvm->path;
    
    // 生成UUID（如果没有提供）
    if (uuid.empty()) {
        uuid = "00000000-0000-0000-0000-000000000000";
    }
    
    // 构建类路径
    std::string classpath = config.buildClasspath();
    if (classpath.empty()) {
        std::cerr << "Warning: Empty classpath\n";
    }
    
    // 启动游戏
    MC_LOG_INFO << "Launching Minecraft...";
    int result = jvm_mgr.launch(*jvm, classpath, username, uuid, width, height);
    
    MC_LOG_INFO << "Game exited with code: " << result;
    return result;
}
