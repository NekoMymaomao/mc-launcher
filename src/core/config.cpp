#include "core/config.hpp"

namespace mc {

bool Config::loadVersion(const std::string& json_str) {
    try {
        nlohmann::json j = nlohmann::json::parse(json_str);
        
        version_.id = j.value("id", "");
        version_.mainClass = j.value("mainClass", "net.minecraft.client.main.Main");
        version_.assetIndex = j.value("assetIndex", nlohmann::json::object()).value("id", "legacy");
        version_.assets = j.value("assets", version_.assetIndex);
        version_.arguments = j.value("arguments", "");
        
        parseLibraries(j);
        parseRules(j);
        
        MC_LOG_INFO << "Loaded version: " << version_.id;
        return true;
    } catch (const std::exception& e) {
        MC_LOG_ERROR << "Failed to parse version JSON: " << e.what();
        return false;
    }
}

bool Config::loadVersionFile(const std::string& path) {
    std::ifstream file(path);
    if (!file.is_open()) {
        MC_LOG_ERROR << "Cannot open version file: " << path;
        return false;
    }
    
    std::stringstream buffer;
    buffer << file.rdbuf();
    
    return loadVersion(buffer.str());
}

void Config::parseLibraries(const nlohmann::json& j) {
    if (!j.contains("libraries")) return;
    
    for (const auto& lib : j["libraries"]) {
        Library lib_entry;
        lib_entry.name = lib.value("name", "");
        
        if (lib.contains("downloads")) {
            auto& downloads = lib["downloads"];
            if (downloads.contains("artifact")) {
                lib_entry.url = downloads["artifact"]["url"].get<std::string>();
                lib_entry.classifier = downloads["artifact"].value("classifier", "");
                if (!lib_entry.classifier.empty()) {
                    lib_entry.natives = true;
                }
            }
        }
        
        version_.libraries.push_back(lib_entry);
    }
}

void Config::parseRules(const nlohmann::json& j) {
    if (!j.contains("rules")) return;
    
    for (const auto& rule : j["rules"]) {
        Rule r;
        r.action = rule.value("action", 0);
        if (rule.contains("os")) {
            auto& os = rule["os"];
            r.os_name = os.value("name", "");
            r.os_version = os.value("version", "");
        }
        version_.rules.push_back(r);
    }
}

std::string Config::resolveLibraryPath(const std::string& lib_name) const {
    // Maven坐标转路径: group:name:version -> group/name/version/name-version.jar
    auto parts = utils::StringUtils::split(lib_name, ':');
    if (parts.size() < 3) return lib_name;
    
    std::string group = parts[0];
    std::string name = parts[1];
    std::string version = parts[2];
    
    std::string path = library_cache_;
    for (size_t i = 0; i < group.size(); ++i) {
        path += utils::PathUtils::SEPARATOR;
        path += group[i];
    }
    path += utils::PathUtils::SEPARATOR + name;
    path += utils::PathUtils::SEPARATOR + version;
    path += utils::PathUtils::SEPARATOR + name + "-" + version + ".jar";
    
    return path;
}

std::string Config::buildClasspath() const {
    std::stringstream cp;
    for (const auto& lib : version_.libraries) {
        if (lib.natives) continue; // native库需要特殊处理
        std::string path = resolveLibraryPath(lib.name);
        if (utils::PathUtils::fileExists(path)) {
            if (!cp.str().empty()) cp << utils::PathUtils::SEPARATOR;
            cp << path;
        }
    }
    return cp.str();
}

} // namespace mc
