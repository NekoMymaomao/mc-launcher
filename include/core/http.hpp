#pragma once
#include <string>
#include <functional>
#include <curl/curl.h>
#include "utils.hpp"

namespace mc {

class HttpClient {
public:
    using ProgressCallback = std::function<void(int64_t downloaded, int64_t total)>;

    HttpClient();
    ~HttpClient();

    std::string get(const std::string& url, int timeoutSec = 30);
    std::string post(const std::string& url, const std::string& body, int timeoutSec = 30);
    std::string download(const std::string& url, const std::string& dest,
                         ProgressCallback progress = nullptr, bool resume = true);
    void cancel();

private:
    CURL* easy = nullptr;
    bool cancelled = false;
    ProgressCallback progressCb;

    static size_t writeCallback(void* contents, size_t size, size_t nmemb, void* userp);
    static int progressCallback(void* clientp, curl_off_t dltotal, curl_off_t dlnow,
                                 curl_off_t ultotal, curl_off_t ulnow);
};

} // namespace mc
