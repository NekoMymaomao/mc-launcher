#pragma once
#include <string>
#include <vector>
#include <functional>
#include "utils.hpp"
#include "http.hpp"

namespace mc {

class InstanceManager {
public:
    InstanceManager();
    bool install(const std::string& versionId,
                 const std::function<void(int, int, const std::string&)>& progress = nullptr);
    bool uninstall(const std::string& versionId);
    std::vector<std::string> listInstalled() const;

private:
    std::string getDir(const std::string& versionId) const;
    std::string getLibsDir() const;
};

} // namespace mc
