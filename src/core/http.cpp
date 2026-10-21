#include "core/http.hpp"
#include <sstream>
#include <thread>
#include <mutex>
#include <algorithm>

namespace mc {

struct HttpClient::Impl {
    CURL* easy = nullptr;
    std::atomic<bool> cancelled{false};
    ProgressCallback progressCb;
    std::mutex progressMutex;
};

HttpClient::HttpClient() : impl_(std::make_unique<Impl>()) {
    curl_global_init(CURL_GLOBAL_DEFAULT);
    impl_->easy = curl_easy_init();
}

HttpClient::~HttpClient() {
    if (impl_->easy) curl_easy_cleanup(impl_->easy);
    curl_global_cleanup();
}

std::string HttpClient::get(const std::string& url, 
                            const std::unordered_map<std::string, std::string>& headers,
                            int timeoutSec) {
    if (!impl_->easy || impl_->cancelled) return "";
    
    struct curl_slist* headerList = nullptr;
    for (const auto& [key, val] : headers) {
        headerList = curl_slist_append(headerList, (key + ": " + val).c_str());
    }
    
    std::string response;
    curl_easy_reset(impl_->easy);
    curl_easy_setopt(impl_->easy, CURLOPT_URL, url.c_str());
    curl_easy_setopt(impl_->easy, CURLOPT_WRITEFUNCTION, writeCallback);
    curl_easy_setopt(impl_->easy, CURLOPT_WRITEDATA, &response);
    curl_easy_setopt(impl_->easy, CURLOPT_HTTPHEADER, headerList);
    curl_easy_setopt(impl_->easy, CURLOPT_TIMEOUT, (long)timeoutSec);
    curl_easy_setopt(impl_->easy, CURLOPT_FOLLOWLOCATION, 1L);
    curl_easy_setopt(impl_->easy, CURLOPT_FAILONERROR, true);
    
    CURLcode res = curl_easy_perform(impl_->easy);
    curl_slist_free_all(headerList);
    
    if (res != CURLE_OK) {
        MC_LOG_ERROR << "GET failed: " << curl_easy_strerror(res);
        return "";
    }
    return response;
}

std::string HttpClient::post(const std::string& url, const std::string& body,
                              const std::unordered_map<std::string, std::string>& headers,
                              int timeoutSec) {
    if (!impl_->easy || impl_->cancelled) return "";
    
    struct curl_slist* headerList = nullptr;
    headerList = curl_slist_append(headerList, "Content-Type: application/json");
    for (const auto& [key, val] : headers) {
        headerList = curl_slist_append(headerList, (key + ": " + val).c_str());
    }
    
    std::string response;
    curl_easy_reset(impl_->easy);
    curl_easy_setopt(impl_->easy, CURLOPT_URL, url.c_str());
    curl_easy_setopt(impl_->easy, CURLOPT_POST, 1L);
    curl_easy_setopt(impl_->easy, CURLOPT_POSTFIELDS, body.c_str());
    curl_easy_setopt(impl_->easy, CURLOPT_POSTFIELDSIZE, body.size());
    curl_easy_setopt(impl_->easy, CURLOPT_WRITEFUNCTION, writeCallback);
    curl_easy_setopt(impl_->easy, CURLOPT_WRITEDATA, &response);
    curl_easy_setopt(impl_->easy, CURLOPT_HTTPHEADER, headerList);
    curl_easy_setopt(impl_->easy, CURLOPT_TIMEOUT, (long)timeoutSec);
    curl_easy_setopt(impl_->easy, CURLOPT_FOLLOWLOCATION, 1L);
    curl_easy_setopt(impl_->easy, CURLOPT_FAILONERROR, true);
    
    CURLcode res = curl_easy_perform(impl_->easy);
    curl_slist_free_all(headerList);
    
    if (res != CURLE_OK) {
        MC_LOG_ERROR << "POST failed: " << curl_easy_strerror(res);
        return "";
    }
    return response;
}

std::string HttpClient::download(const std::string& url, const std::string& dest,
                                  ProgressCallback progress, bool resume) {
    if (!impl_->easy || impl_->cancelled) return "";
    
    // 检查断点续传
    long fileSize = 0;
    if (resume && utils::PathUtils::fileExists(dest)) {
        std::ifstream f(dest, std::ios::binary | std::ios::ate);
        fileSize = f.tellg();
        f.close();
    }
    
    FILE* fp = fopen(dest.c_str(), resume && fileSize > 0 ? "ab" : "wb");
    if (!fp) {
        MC_LOG_ERROR << "Cannot open file: " << dest;
        return "";
    }
    
    curl_easy_reset(impl_->easy);
    curl_easy_setopt(impl_->easy, CURLOPT_URL, url.c_str());
    curl_easy_setopt(impl_->easy, CURLOPT_WRITEFUNCTION, writeCallback);
    curl_easy_setopt(impl_->easy, CURLOPT_WRITEDATA, fp);
    curl_easy_setopt(impl_->easy, CURLOPT_FAILONERROR, true);
    curl_easy_setopt(impl_->easy, CURLOPT_FOLLOWLOCATION, true);
    curl_easy_setopt(impl_->easy, CURLOPT_TIMEOUT, 300L);
    
    if (resume && fileSize > 0) {
        curl_easy_setopt(impl_->easy, CURLOPT_RESUME_FROM_LARGE, (curl_off_t)fileSize);
        curl_easy_setopt(impl_->easy, CURLOPT_RANGE, 
            std::to_string(fileSize).c_str());
    }
    
    if (progress) {
        impl_->progressCb = progress;
        curl_easy_setopt(impl_->easy, CURLOPT_NOPROGRESS, false);
        curl_easy_setopt(impl_->easy, CURLOPT_PROGRESSFUNCTION, progressCallback);
        curl_easy_setopt(impl_->easy, CURLOPT_PROGRESSDATA, (void*)this);
    }
    
    CURLcode res = curl_easy_perform(impl_->easy);
    fclose(fp);
    
    if (res != CURLE_OK) {
        MC_LOG_ERROR << "Download failed: " << curl_easy_strerror(res);
        remove(dest.c_str());
        return "";
    }
    
    return dest;
}

std::vector<std::string> HttpClient::downloadParallel(
    const std::vector<std::pair<std::string, std::string>>& tasks,
    ProgressCallback progress, int concurrent) {
    
    std::vector<std::string> results(tasks.size(), "");
    std::mutex resultMutex;
    std::atomic<int> completed{0};
    
    auto worker = [&](int id) {
        for (size_t i = id; i < tasks.size(); i += concurrent) {
            if (impl_->cancelled) break;
            
            std::string result = download(tasks[i].first, tasks[i].second, progress);
            
            std::lock_guard<std::mutex> lock(resultMutex);
            results[i] = result;
            completed++;
            
            if (progress) {
                progress(completed.load(), (int)tasks.size(), 0);
            }
        }
    };
    
    std::vector<std::thread> threads;
    for (int i = 0; i < concurrent; i++) {
        threads.emplace_back(worker, i);
    }
    for (auto& t : threads) {
        t.join();
    }
    
    return results;
}

void HttpClient::cancel() {
    impl_->cancelled = true;
}

size_t HttpClient::writeCallback(void* contents, size_t size, size_t nmemb, void* userp) {
    std::string* buf = reinterpret_cast<std::string*>(userp);
    buf->append((char*)contents, size * nmemb);
    return size * nmemb;
}

int HttpClient::progressCallback(void* clientp, curl_off_t dltotal, curl_off_t dlnow,
                                  curl_off_t /*ultotal*/, curl_off_t /*ulnow*/) {
    auto* self = static_cast<HttpClient*>(clientp);
    if (self->impl_->progressCb && dltotal > 0) {
        std::lock_guard<std::mutex> lock(self->impl_->progressMutex);
        self->impl_->progressCb(dlnow, dltotal, 0);
    }
    return self->impl_->cancelled ? 1 : 0;
}

} // namespace mc
