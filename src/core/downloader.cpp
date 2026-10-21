#include "core/downloader.hpp"
#include <cstring>
#include <thread>
#include <atomic>

namespace mc {

struct Downloader::Impl {
    CURL* easy_handle = nullptr;
    std::atomic<bool> cancelled{false};
};

Downloader::Downloader() : impl_(std::make_unique<Impl>()) {
    curl_global_init(CURL_GLOBAL_DEFAULT);
    impl_->easy_handle = curl_easy_init();
}

Downloader::~Downloader() {
    if (impl_->easy_handle) {
        curl_easy_cleanup(impl_->easy_handle);
    }
    curl_global_cleanup();
}

std::string Downloader::download(const std::string& url, const std::string& dest,
                                 ProgressCallback progress) {
    if (!impl_->easy_handle) return "";
    
    if (cancelled_) {
        MC_LOG_WARN << "Download cancelled";
        return "";
    }
    
    FILE* fp = fopen(dest.c_str(), "wb");
    if (!fp) {
        MC_LOG_ERROR << "Cannot open file: " << dest;
        return "";
    }
    
    curl_easy_reset(impl_->easy_handle);
    curl_easy_setopt(impl_->easy_handle, CURLOPT_URL, url.c_str());
    curl_easy_setopt(impl_->easy_handle, CURLOPT_WRITEFUNCTION, writeCallback);
    curl_easy_setopt(impl_->easy_handle, CURLOPT_WRITEDATA, fp);
    curl_easy_setopt(impl_->easy_handle, CURLOPT_FAILONERROR, true);
    curl_easy_setopt(impl_->easy_handle, CURLOPT_FOLLOWLOCATION, true);
    curl_easy_setopt(impl_->easy_handle, CURLOPT_TIMEOUT, 60L);
    
    if (progress) {
        curl_easy_setopt(impl_->easy_handle, CURLOPT_NOPROGRESS, false);
        curl_easy_setopt(impl_->easy_handle, CURLOPT_PROGRESSFUNCTION, progressCallback);
        curl_easy_setopt(impl_->easy_handle, CURLOPT_PROGRESSDATA, (void*)&progress);
    }
    
    CURLcode res = curl_easy_perform(impl_->easy_handle);
    fclose(fp);
    
    if (res != CURLE_OK) {
        MC_LOG_ERROR << "Download failed: " << curl_easy_strerror(res);
        return "";
    }
    
    return dest;
}

std::string Downloader::downloadString(const std::string& url) {
    if (!impl_->easy_handle) return "";
    
    std::string response;
    
    curl_easy_reset(impl_->easy_handle);
    curl_easy_setopt(impl_->easy_handle, CURLOPT_URL, url.c_str());
    curl_easy_setopt(impl_->easy_handle, CURLOPT_WRITEFUNCTION, writeCallback);
    curl_easy_setopt(impl_->easy_handle, CURLOPT_WRITEDATA, &response);
    curl_easy_setopt(impl_->easy_handle, CURLOPT_FAILONERROR, true);
    curl_easy_setopt(impl_->easy_handle, CURLOPT_FOLLOWLOCATION, true);
    curl_easy_setopt(impl_->easy_handle, CURLOPT_TIMEOUT, 30L);
    
    CURLcode res = curl_easy_perform(impl_->easy_handle);
    
    if (res != CURLE_OK) {
        MC_LOG_ERROR << "Download failed: " << curl_easy_strerror(res);
        return "";
    }
    
    return response;
}

void Downloader::cancel() {
    cancelled_ = true;
}

size_t Downloader::writeCallback(void* contents, size_t size, size_t nmemb, std::string* userp) {
    userp->append((char*)contents, size * nmemb);
    return size * nmemb;
}

size_t Downloader::progressCallback(void* clientp, double dltotal, double dlnow,
                                     double ultotal, double ulnow) {
    auto* progress = reinterpret_cast<ProgressCallback*>(clientp);
    if (progress && dltotal > 0) {
        (*progress)((int64_t)dlnow, (int64_t)dltotal);
    }
    return 0;
}

} // namespace mc
