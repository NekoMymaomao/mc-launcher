#pragma once
#include <string>
#include <functional>
#include <vector>
#include <optional>
#include <curl/curl.h>
#include "logger.hpp"

namespace mc {

// HTTP 下载器
class HttpClient {
public:
    using ProgressCallback = std::function<void(int64_t downloaded, int64_t total, double speed)>;
    using ResponseCallback = std::function<void(const std::string& data, size_t size)>;

    HttpClient();
    ~HttpClient();

    // GET 请求
    std::string get(const std::string& url, 
                    const std::unordered_map<std::string, std::string>& headers = {},
                    int timeoutSec = 30);

    // POST 请求
    std::string post(const std::string& url, 
                     const std::string& body,
                     const std::unordered_map<std::string, std::string>& headers = {},
                     int timeoutSec = 30);

    // 下载文件（支持断点续传）
    std::string download(const std::string& url, const std::string& dest,
                         ProgressCallback progress = nullptr,
                         bool resume = true);

    // 多线程并发下载
    std::vector<std::string> downloadParallel(
        const std::vector<std::pair<std::string, std::string>>& tasks,
        ProgressCallback progress = nullptr,
        int concurrent = 4);

    // 取消下载
    void cancel();

private:
    struct Impl;
    std::unique_ptr<Impl> impl_;
    
    static size_t writeCallback(void* contents, size_t size, size_t nmemb, void* userp);
    static int progressCallback(void* clientp, curl_off_t dltotal, curl_off_t dlnow,
                                 curl_off_t ultotal, curl_off_t ulnow);
};

} // namespace mc
