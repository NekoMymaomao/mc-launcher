#include "core/crash.hpp"
#include "core/utils.hpp"
#include <fstream>
#include <regex>

namespace mc {

// ============================================================
// CrashAnalyzer 实现
// ============================================================

CrashAnalyzer::CrashAnalyzer() {}

CrashReport CrashAnalyzer::analyze(const std::string& crashFilePath) {
    CrashReport report;
    report.filePath = crashFilePath;
    
    std::ifstream f(crashFilePath);
    if (!f.is_open()) return report;
    
    std::stringstream ss;
    ss << f.rdbuf();
    std::string content = ss.str();
    f.close();
    
    // 解析基本信息
    std::regex timeRegex(R"(Time: ([^\n]+))");
    std::smatch match;
    if (std::regex_search(content, match, timeRegex)) {
        report.crashTime = match[1].str();
    }
    
    // 解析错误信息
    report.errorMessage = parseErrorMessage(content);
    
    // 解析堆栈
    report.stackTrace = parseStackTrace(content);
    
    // 解析模组
    report.modules = parseModules(content);
    
    // 计算严重性
    report.severity = calculateSeverity(report);
    
    return report;
}

std::vector<std::string> CrashAnalyzer::listCrashes(const std::string& version) const {
    std::vector<std::string> crashes;
    std::string crashDir = join(ConfigManager::instance().getConfig().gameDir, "crash-reports");
    
    auto files = utils::PathUtils::listFiles(crashDir, ".txt");
    for (const auto& f : files) {
        crashes.push_back(f);
    }
    return crashes;
}

bool CrashAnalyzer::deleteCrash(const std::string& crashPath) {
    return remove(crashPath.c_str()) == 0;
}

std::string CrashAnalyzer::analyzeLog(const std::string& logPath) {
    std::ifstream f(logPath);
    if (!f.is_open()) return "无法打开日志文件";
    
    std::stringstream ss;
    ss << f.rdbuf();
    std::string content = ss.str();
    
    // 查找错误
    std::vector<std::string> errors;
    std::string line;
    std::stringstream lineStream(content);
    while (std::getline(lineStream, line)) {
        if (line.find("ERROR") != std::string::npos || 
            line.find("Exception") != std::string::npos ||
            line.find("Fatal") != std::string::npos) {
            errors.push_back(line);
        }
    }
    
    if (errors.empty()) return "未发现明显错误";
    
    std::string result = "发现 " + std::to_string(errors.size()) + " 个错误:\n\n";
    for (size_t i = 0; i < std::min(errors.size(), size_t(10)); i++) {
        result += errors[i] + "\n";
    }
    if (errors.size() > 10) result += "... (还有 " + std::to_string(errors.size() - 10) + " 个)\n";
    
    return result;
}

std::string CrashAnalyzer::parseErrorMessage(const std::string& content) {
    // 尝试多种模式
    std::regex patterns[] = {
        std::regex(R"(Exception: ([^\n]+))"),
        std::regex(R"(Caused by: ([^\n]+))"),
        std::regex(R"(FATAL ERROR: ([^\n]+))")
    };
    
    for (const auto& pattern : patterns) {
        std::smatch match;
        if (std::regex_search(content, match, pattern)) {
            return match[1].str();
        }
    }
    return "未知错误";
}

std::vector<std::string> CrashAnalyzer::parseStackTrace(const std::string& content) {
    std::vector<std::string> stack;
    std::string line;
    std::stringstream lineStream(content);
    bool inStack = false;
    
    while (std::getline(lineStream, line)) {
        if (line.find("at ") == 0 || line.find("Caused by:") == 0) {
            inStack = true;
            stack.push_back(line);
        } else if (inStack && line.find("...") != std::string::npos) {
            stack.push_back(line);
        } else if (inStack && !line.empty() && line[0] != ' ') {
            break;
        } else if (inStack) {
            stack.push_back(line);
        }
    }
    return stack;
}

std::vector<std::string> CrashAnalyzer::parseModules(const std::string& content) {
    std::vector<std::string> modules;
    std::regex modulePattern(R"(\[([^\]]+)\])");
    std::sregex_iterator it(content.begin(), content.end(), modulePattern);
    std::sregex_iterator end;
    
    for (; it != end; ++it) {
        std::string mod = (*it)[1].str();
        if (mod.find("Java") == std::string::npos && 
            mod.find("GL") == std::string::npos &&
            mod.find("Thread") == std::string::npos) {
            modules.push_back(mod);
        }
    }
    return modules;
}

int CrashAnalyzer::calculateSeverity(const CrashReport& report) {
    if (report.errorMessage.find("OutOfMemory") != std::string::npos) return 1;
    if (report.errorMessage.find("Crash") != std::string::npos) return 2;
    if (report.errorMessage.find("Fatal") != std::string::npos) return 3;
    return 0;
}

} // namespace mc
