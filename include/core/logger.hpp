#pragma once
#include <string>
#include <functional>
#include <atomic>
#include <curl/curl.h>
#include "utils.hpp"

namespace mc {

// 日志级别
enum class LogLevel { DEBUG, INFO, WARN, ERROR };

// 日志类
class Logger {
public:
    static Logger& instance() {
        static Logger inst;
        return inst;
    }

    void setLevel(LogLevel level) { level_ = level; }
    void setOutput(std::ostream& os) { output_ = &os; }

    template<typename... Args>
    void log(LogLevel level, const char* file, int line, Args&&... args) {
        if (level < level_) return;
        
        auto now = std::chrono::system_clock::now();
        auto time_t = std::chrono::system_clock::to_time_t(now);
        auto ms = std::chrono::duration_cast<std::chrono::milliseconds>(
            now.time_since_epoch()) % 1000;

        std::tm tm_buf;
#ifdef MC_WINDOWS
        localtime_s(&tm_buf, &time_t);
#else
        localtime_r(&time_t, &tm_buf);
#endif

        *output_ << "[" 
                 << std::put_time(&tm_buf, "%Y-%m-%d %H:%M:%S")
                 << "." << std::setfill('0') << std::setw(3) << ms.count() 
                 << "] [" << levelStr(level) << "] "
                 << utils::PathUtils::getFileName(file) << ":" << line << " - ";
        
        (output_ << ... << std::forward<Args>(args));
        *output_ << "\n";
        output_->flush();
    }

#define MC_LOG(level) mc::Logger::instance().log(level, __FILE__, __LINE__)
#define MC_LOG_DEBUG MC_LOG(LogLevel::DEBUG)
#define MC_LOG_INFO MC_LOG(LogLevel::INFO)
#define MC_LOG_WARN MC_LOG(LogLevel::WARN)
#define MC_LOG_ERROR MC_LOG(LogLevel::ERROR)

private:
    Logger() : level_(LogLevel::INFO), output_(&std::cout) {}
    
    static std::string levelStr(LogLevel level) {
        switch(level) {
            case LogLevel::DEBUG: return "DBG";
            case LogLevel::INFO:  return "INF";
            case LogLevel::WARN:  return "WRN";
            case LogLevel::ERROR: return "ERR";
        }
        return "UNK";
    }

    LogLevel level_;
    std::ostream* output_;
};

} // namespace mc
