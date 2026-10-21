#pragma once
#include <string>
#include <vector>
#include <map>
#include "logger.hpp"

namespace mc {

// 崩溃分析报告
struct CrashReport {
    std::string filePath;
    std::string version;
    std::string minecraftVersion;
    std::string crashTime;
    std::string errorMessage;
    std::vector<std::string> stackTrace;
    std::vector<std::string> modules;
    int severity = 0; // 0-3
};

// 崩溃分析器
class CrashAnalyzer {
public:
    CrashAnalyzer();
    ~CrashAnalyzer() = default;

    // 分析崩溃报告
    CrashReport analyze(const std::string& crashFilePath);

    // 列出崩溃报告
    std::vector<std::string> listCrashes(const std::string& version) const;

    // 删除崩溃报告
    bool deleteCrash(const std::string& crashPath);

    // 分析日志
    std::string analyzeLog(const std::string& logPath);

private:
    std::string parseErrorMessage(const std::string& content);
    std::vector<std::string> parseStackTrace(const std::string& content);
    std::vector<std::string> parseModules(const std::string& content);
    int calculateSeverity(const CrashReport& report);
};

} // namespace mc
