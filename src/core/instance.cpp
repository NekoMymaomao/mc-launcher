#include "core/instance.hpp"
#include <fstream>
#include <iostream>

namespace mc {

InstanceManager::InstanceManager() {}

std::string InstanceManager::getDir(const std::string& versionId) const {
    return utils::PathUtils::join(utils::PathUtils::getLauncherDir(), "versions", versionId);
}

std::string InstanceManager::getLibsDir() const {
    return utils::PathUtils::join(utils::PathUtils::getLauncherDir(), "libraries");
}

bool InstanceManager::install(const std::string& versionId,
                               const std::function<void(int, int, const std::string&)>& progress) {
    if (progress) progress(0, 5, "获取版本信息...");

    HttpClient http;
    std::string manifestJson = http.get("https://launchermeta.mojang.com/mc/game/" + versionId + ".json");
    if (manifestJson.empty()) {
        std::cerr << "无法获取版本信息: " << versionId << std::endl;
        return false;
    }

    if (progress) progress(1, 5, "已下载版本信息");

    std::string dir = getDir(versionId);
    utils::PathUtils::createDirectory(dir);

    // 保存 manifest
    {
        std::ofstream f(dir + "/" + versionId + ".json");
        f << manifestJson;
    }

    if (progress) progress(2, 5, "下载核心库...");

    // 下载客户端 jar
    auto j = utils::Json::parse(manifestJson);
    if (j.contains("downloads") && j["downloads"].contains("client")) {
        std::string url = j["downloads"]["client"]["url"].getString();
        std::string dest = dir + "/" + versionId + ".jar";
        if (http.download(url, dest, progress ? nullptr : nullptr) != dest) {
            std::cerr << "下载客户端失败" << std::endl;
            return false;
        }
    }

    if (progress) progress(4, 5, "下载完成");
    std::cout << "✓ 版本 " << versionId << " 安装完成" << std::endl;
    return true;
}

bool InstanceManager::uninstall(const std::string& versionId) {
    std::string dir = getDir(versionId);
#ifdef MC_WINDOWS
    return (std::remove(dir.c_str()) == 0) ||
           (utils::ProcessUtils::run("rmdir /s /q \"" + dir + "\"") == 0);
#else
    return utils::ProcessUtils::run("rm -rf \"" + dir + "\"") == 0;
#endif
}

std::vector<std::string> InstanceManager::listInstalled() const {
    std::vector<std::string> versions;
    std::string dir = utils::PathUtils::join(utils::PathUtils::getLauncherDir(), "versions");
    auto files = utils::PathUtils::listFiles(dir);
    for (auto& f : files) {
        if (utils::PathUtils::fileExists(f + "/" + utils::PathUtils::getFileName(f) + ".json"))
            versions.push_back(utils::PathUtils::getFileName(f));
    }
    return versions;
}

} // namespace mc
