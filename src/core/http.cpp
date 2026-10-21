#include "core/http.hpp"
#include <fstream>
#include <sstream>

namespace mc {

HttpClient::HttpClient() {
    curl_global_init(CURL_GLOBAL_DEFAULT);
    easy = curl_easy_init();
}

HttpClient::~HttpClient() {
    if (easy) curl_easy_cleanup(easy);
    curl_global_cleanup();
}

std::string HttpClient::get(const std::string& url, int timeoutSec) {
    if (!easy || cancelled) return "";
    std::string response;
    curl_easy_reset(easy);
    curl_easy_setopt(easy, CURLOPT_URL, url.c_str());
    curl_easy_setopt(easy, CURLOPT_WRITEFUNCTION, writeCallback);
    curl_easy_setopt(easy, CURLOPT_WRITEDATA, &response);
    curl_easy_setopt(easy, CURLOPT_TIMEOUT, (long)timeoutSec);
    curl_easy_setopt(easy, CURLOPT_FOLLOWLOCATION, 1L);
    CURLcode res = curl_easy_perform(easy);
    return (res == CURLE_OK) ? response : "";
}

std::string HttpClient::post(const std::string& url, const std::string& body, int timeoutSec) {
    if (!easy || cancelled) return "";
    std::string response;
    curl_easy_reset(easy);
    curl_easy_setopt(easy, CURLOPT_URL, url.c_str());
    curl_easy_setopt(easy, CURLOPT_POST, 1L);
    curl_easy_setopt(easy, CURLOPT_POSTFIELDS, body.c_str());
    curl_easy_setopt(easy, CURLOPT_POSTFIELDSIZE, body.size());
    curl_easy_setopt(easy, CURLOPT_WRITEFUNCTION, writeCallback);
    curl_easy_setopt(easy, CURLOPT_WRITEDATA, &response);
    curl_easy_setopt(easy, CURLOPT_TIMEOUT, (long)timeoutSec);
    curl_easy_setopt(easy, CURLOPT_FOLLOWLOCATION, 1L);
    struct curl_slist* headers = curl_slist_append(nullptr, "Content-Type: application/json");
    curl_easy_setopt(easy, CURLOPT_HTTPHEADER, headers);
    CURLcode res = curl_easy_perform(easy);
    curl_slist_free_all(headers);
    return (res == CURLE_OK) ? response : "";
}

std::string HttpClient::download(const std::string& url, const std::string& dest,
                                   ProgressCallback progress, bool resume) {
    if (!easy || cancelled) return "";

    long fileSize = 0;
    if (resume) {
        std::ifstream f(dest, std::ios::binary | std::ios::ate);
        if (f.is_open()) {
            f.seekg(0, std::ios::end);
            fileSize = f.tellg();
            f.close();
        }
    }

    FILE* fp = fopen(dest.c_str(), resume && fileSize > 0 ? "ab" : "wb");
    if (!fp) return "";

    curl_easy_reset(easy);
    curl_easy_setopt(easy, CURLOPT_URL, url.c_str());
    curl_easy_setopt(easy, CURLOPT_WRITEFUNCTION, writeCallback);
    curl_easy_setopt(easy, CURLOPT_WRITEDATA, fp);
    curl_easy_setopt(easy, CURLOPT_TIMEOUT, 300L);
    curl_easy_setopt(easy, CURLOPT_FOLLOWLOCATION, 1L);
    if (resume && fileSize > 0)
        curl_easy_setopt(easy, CURLOPT_RESUME_FROM_LARGE, (curl_off_t)fileSize);

    if (progress) {
        progressCb = progress;
        curl_easy_setopt(easy, CURLOPT_NOPROGRESS, false);
        curl_easy_setopt(easy, CURLOPT_PROGRESSFUNCTION, progressCallback);
        curl_easy_setopt(easy, CURLOPT_PROGRESSDATA, this);
    }

    CURLcode res = curl_easy_perform(easy);
    fclose(fp);

    if (res != CURLE_OK) {
        remove(dest.c_str());
        return "";
    }
    return dest;
}

void HttpClient::cancel() { cancelled = true; }

size_t HttpClient::writeCallback(void* contents, size_t size, size_t nmemb, void* userp) {
    std::string* buf = reinterpret_cast<std::string*>(userp);
    buf->append((char*)contents, size * nmemb);
    return size * nmemb;
}

int HttpClient::progressCallback(void* clientp, curl_off_t dltotal, curl_off_t dlnow,
                                  curl_off_t, curl_off_t) {
    auto* self = static_cast<HttpClient*>(clientp);
    if (self->progressCb && dltotal > 0)
        self->progressCb(dlnow, dltotal);
    return self->cancelled ? 1 : 0;
}

} // namespace mc
