#pragma once
#include <string>
#include <vector>
#include <stdexcept>
#include <curl/curl.h>
#include "logger.hpp"

namespace mc {
namespace utils {

// 跨平台路径处理
class PathUtils {
public:
#ifdef MC_WINDOWS
    static constexpr char SEPARATOR = '\\';
    static constexpr char ALT_SEPARATOR = '/';
#else
    static constexpr char SEPARATOR = '/';
    static constexpr char ALT_SEPARATOR = '\\';
#endif

    static std::string join(const std::string& a, const std::string& b) {
        if (a.empty()) return b;
        if (b.empty()) return a;
        
        char last = a.back();
        char first = b.front();
        
        if (last == SEPARATOR || last == ALT_SEPARATOR) {
            if (first == SEPARATOR || first == ALT_SEPARATOR) {
                return a + b.substr(1);
            }
            return a + b;
        }
        return a + std::string(1, SEPARATOR) + b;
    }

    static std::string getHomeDir() {
#ifdef MC_WINDOWS
        const char* home = getenv("USERPROFILE");
        if (!home) home = getenv("HOME");
        return home ? home : "C:\\Users\\Default";
#else
        const char* home = getenv("HOME");
        return home ? home : "/tmp";
#endif
    }

    static std::string getGameDir() {
#ifdef MC_WINDOWS
        const char* minecraft = getenv("MINECRAFT_DIR");
        if (minecraft) return minecraft;
        return getHomeDir() + "\\AppData\\Roaming\\.minecraft";
#else
        const char* minecraft = getenv("MINECRAFT_DIR");
        if (minecraft) return minecraft;
        return getHomeDir() + "/.minecraft";
#endif
    }

    static bool createDirectory(const std::string& path) {
#ifdef MC_WINDOWS
        return _mkdir(path.c_str()) == 0 || errno == EEXIST;
#else
        return mkdir(path.c_str(), 0755) == 0 || errno == EEXIST;
#endif
    }

    static bool fileExists(const std::string& path) {
#ifdef MC_WINDOWS
        return _access(path.c_str(), 0) == 0;
#else
        return access(path.c_str(), F_OK) == 0;
#endif
    }
};

// 字符串处理
class StringUtils {
public:
    static std::vector<std::string> split(const std::string& str, char delim) {
        std::vector<std::string> tokens;
        std::stringstream ss(str);
        std::string item;
        while (std::getline(ss, item, delim)) {
            if (!item.empty()) {
                tokens.push_back(item);
            }
        }
        return tokens;
    }

    static std::string trim(const std::string& str) {
        size_t start = str.find_first_not_of(" \t\n\r");
        size_t end = str.find_last_not_of(" \t\n\r");
        if (start == std::string::npos) return "";
        return str.substr(start, end - start + 1);
    }

    static std::string toLower(const std::string& str) {
        std::string result = str;
        std::transform(result.begin(), result.end(), result.begin(), ::tolower);
        return result;
    }
};

} // namespace utils
} // namespace mc
