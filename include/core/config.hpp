#pragma once
#include <string>
#include <vector>
#include <map>
#include <memory>
#include <fstream>
#include <sstream>
#include <algorithm>
#include <nlohmann/json.hpp>
#include "logger.hpp"

#ifdef MC_WINDOWS
    #include <windows.h>
    #include <direct.h>
    #include <io.h>
#elif defined(MC_LINUX)
    #include <unistd.h>
    #include <dirent.h>
    #include <sys/stat.h>
#endif

namespace mc {

// 版本数据结构
struct Library {
    std::string name;
    std::string url;
    bool natives = false;
    std::string classifier;
};

struct Rule {
    int action = 0;
    std::string os_name;
    std::string os_version;
};

struct Version {
    std::string id;
    std::string mainClass = "net.minecraft.client.main.Main";
    std::string assetIndex = "legacy";
    std::string assets = "legacy";
    std::vector<Library> libraries;
    std::vector<Rule> rules;
    std::string arguments;
};

// 配置类
class Config {
public:
    Config() = default;
    ~Config() = default;

    // 加载版本JSON
    bool loadVersion(const std::string& json_str);
    bool loadVersionFile(const std::string& path);
    
    // 获取配置
    const Version& getVersion() const { return version_; }
    const std::string& getGameDir() const { return game_dir_; }
    void setGameDir(const std::string& dir) { game_dir_ = dir; }
    
    // 解析库路径
    std::string resolveLibraryPath(const std::string& lib_name) const;
    
    // 构建类路径
    std::string buildClasspath() const;

private:
    Version version_;
    std::string game_dir_;
    std::string library_cache_;
    
    void parseLibraries(const nlohmann::json& j);
    void parseRules(const nlohmann::json& j);
};

} // namespace mc
