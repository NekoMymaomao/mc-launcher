#pragma once
#include <string>
#include <vector>
#include <optional>
#include "utils.hpp"

namespace mc {

struct JavaInfo {
    std::string path;
    std::string version;
    int majorVersion = 0;
    bool isValid = false;
};

class JavaManager {
public:
    JavaManager();
    std::optional<JavaInfo> findJava(const std::string& minVersion = "17") const;
    std::vector<JavaInfo> listJava() const;
    bool installJava(const std::string& version = "17");
    bool setJavaPath(const std::string& path);

private:
    std::string javaPath_;
    std::string getJavaHome() const;
    std::string getJavaFromPath() const;
    bool verifyJava(const std::string& path, int minVersion) const;
    std::string downloadUrl(const std::string& version) const;
};

} // namespace mc
