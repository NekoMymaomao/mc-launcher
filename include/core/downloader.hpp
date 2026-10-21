#pragma once
#include <string>
#include <functional>
#include <future>
#include <atomic>
#include <curl/curl.h>
#include "logger.hpp"

namespace mc {

// 跨平台下载器
class Downloader {
public:
    using ProgressCallback = std::function<void(int64_t downloaded, int64_t total)>;
    using ResultCallback = std::function<void(bool success, const std::string& path)>;

    Downloader();
    ~Downloader();

    // 同步下载
    std::string download(const std::string& url, const std::string& dest,
                        ProgressCallback progress = nullptr);

    // 下载为字符串
    std::string downloadString(const std::string& url);

    // 取消下载
    void cancel();

private:
    struct Impl;
    std::unique_ptr<Impl> impl_;
    
    static size_t writeCallback(void* contents, size_t size, size_t nmemb, std::string* userp);
    static size_t progressCallback(void* clientp, double dltotal, double dlnow, 
                                    double ultotal, double ulnow);
};

} // namespace mc
