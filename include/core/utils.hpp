#pragma once
#include <string>
#include <vector>
#include <sstream>
#include <iomanip>
#include <algorithm>
#include <cstdint>
#include <cstring>
#include <functional>
#include <thread>
#include <mutex>
#include <atomic>
#include <optional>
#include <unordered_map>
#include <fstream>
#include <chrono>

#ifdef MC_WINDOWS
    #include <windows.h>
    #include <direct.h>
    #include <io.h>
    #include <shellapi.h>
    #define PATH_SEP '\\'
#else
    #include <unistd.h>
    #include <dirent.h>
    #include <sys/stat.h>
    #include <sys/wait.h>
    #define PATH_SEP '/'
#endif

namespace mc {
namespace utils {

// ============================================================
// 字符串工具
// ============================================================
class StringUtils {
public:
    static std::vector<std::string> split(const std::string& str, char delim) {
        std::vector<std::string> tokens;
        std::stringstream ss(str);
        std::string item;
        while (std::getline(ss, item, delim)) {
            if (!item.empty()) tokens.push_back(item);
        }
        return tokens;
    }

    static std::string trim(const std::string& str) {
        size_t start = str.find_first_not_of(" \t\n\r");
        size_t end = str.find_last_not_of(" \t\n\r");
        if (start == std::string::npos) return "";
        return str.substr(start, end - start + 1);
    }

    static std::string toLower(const std::string& str) {
        std::string result = str;
        std::transform(result.begin(), result.end(), result.begin(), ::tolower);
        return result;
    }

    static std::string toUpper(const std::string& str) {
        std::string result = str;
        std::transform(result.begin(), result.end(), result.begin(), ::toupper);
        return result;
    }

    static bool startsWith(const std::string& str, const std::string& prefix) {
        return str.size() >= prefix.size() && str.compare(0, prefix.size(), prefix) == 0;
    }

    static bool endsWith(const std::string& str, const std::string& suffix) {
        return str.size() >= suffix.size() && 
               str.compare(str.size() - suffix.size(), suffix.size(), suffix) == 0;
    }

    static std::string replaceAll(const std::string& str, const std::string& from, 
                                   const std::string& to) {
        std::string result = str;
        size_t pos = 0;
        while ((pos = result.find(from, pos)) != std::string::npos) {
            result.replace(pos, from.size(), to);
            pos += to.size();
        }
        return result;
    }

    // Base64 编码/解码
    static std::string base64Encode(const unsigned char* data, size_t len) {
        static const char table[] = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
        std::string result;
        for (size_t i = 0; i < len; i += 3) {
            uint32_t n = 0;
            for (int j = 0; j < 3 && i + j < len; j++) {
                n = (n << 8) | data[i + j];
            }
            n <<= (3 - (len - i)) * 8;
            for (int j = 0; j < 4 && i + j < len + (len - i) % 3; j++) {
                result += table[(n >> (18 - j * 6)) & 0x3F];
            }
            for (int j = 0; j < (len - i) % 3; j++) result.back() = '=';
        }
        return result;
    }

    static std::string base64Decode(const std::string& str) {
        static const int table[] = {
            -1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,
            -1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,
            -1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,62,-1,-1,-1,63,
            52,53,54,55,56,57,58,59,60,61,-1,-1,-1,-1,-1,-1,
            -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9,10,11,12,13,14,
            15,16,17,18,19,20,21,22,23,24,25,-1,-1,-1,-1,-1,
            -1,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,
            41,42,43,44,45,46,47,48,49,50,51,-1,-1,-1,-1,-1
        };
        std::string result;
        uint8_t buf[4];
        int idx = 0;
        for (char c : str) {
            if (c == '=') break;
            int val = c < 128 ? table[c] : -1;
            if (val == -1) continue;
            buf[idx++] = val;
            if (idx == 4) {
                result += (buf[0] << 2) | (buf[1] >> 4);
                if (buf[2] != 63) result += (buf[1] << 4) | (buf[2] >> 2);
                if (buf[3] != 63) result += (buf[2] << 6) | buf[3];
                idx = 0;
            }
        }
        return result;
    }
};

// ============================================================
// SHA-1 实现
// ============================================================
class Sha1 {
public:
    static std::string compute(const std::string& data) {
        uint32_t h0 = 0x67452301, h1 = 0xEFCDAB89, h2 = 0x98BADCFE;
        uint32_t h3 = 0x10325476, h4 = 0xC3D2E1F0;
        
        size_t len = data.size();
        uint64_t bits = len * 8;
        
        std::vector<uint8_t> msg(len + 1);
        memcpy(msg.data(), data.data(), len);
        msg[len] = 0x80;
        msg.resize(((len + 9) % 64) <= 56 ? len + 65 - (len % 64) : len + 129 - (len % 64));
        
        for (int i = 56; i < 64; i++) msg[msg.size()-64+i-56] = 0;
        for (int i = 0; i < 8; i++) msg[msg.size()-8+i] = (bits >> (56 - i*8)) & 0xFF;
        
        for (size_t offset = 0; offset < msg.size(); offset += 64) {
            uint32_t w[64];
            for (int i = 0; i < 16; i++) {
                w[i] = (msg[offset+i*4] << 24) | (msg[offset+i*4+1] << 16) |
                        (msg[offset+i*4+2] << 8) | msg[offset+i*4+3];
            }
            for (int i = 16; i < 64; i++) {
                w[i] = rol(w[i-3] ^ w[i-8] ^ w[i-14] ^ w[i-15], 1);
            }
            
            uint32_t a = h0, b = h1, c = h2, d = h3, e = h4;
            for (int i = 0; i < 64; i++) {
                uint32_t f, k;
                if (i < 20) { f = (b & c) | (~b & d); k = 0x5A827999; }
                else if (i < 40) { f = b ^ c ^ d; k = 0x6ED9EBA1; }
                else if (i < 60) { f = (b & c) | (b & d) | (c & d); k = 0x8F1BBCDC; }
                else { f = b ^ c ^ d; k = 0xCA62C1D6; }
                
                uint32_t temp = rol(a, 5) + f + e + k + w[i];
                e = d; d = c; c = rol(b, 30); b = a; a = temp;
            }
            h0 += a; h1 += b; h2 += c; h3 += d; h4 += e;
        }
        
        std::ostringstream oss;
        oss << std::hex << std::setfill('0');
        oss << std::setw(8) << h0 << std::setw(8) << h1 << std::setw(8) << h2 
            << std::setw(8) << h3 << std::setw(8) << h4;
        return oss.str();
    }

private:
    static uint32_t rol(uint32_t x, int n) { return (x << n) | (x >> (32 - n)); }
};

// ============================================================
// UUID 生成
// ============================================================
class Uuid {
public:
    static std::string generateV4() {
        thread_local std::mt19937 rng(std::chrono::steady_clock::now().time_since_epoch().count());
        std::uniform_int_distribution<int> dist(0, 15);
        std::uniform_int_distribution<int> dist2(8, 11); // 8, 9, A, B
        
        std::string hex = "0123456789abcdef";
        std::string uuid;
        
        for (int i = 0; i < 36; i++) {
            if (i == 8 || i == 13 || i == 18 || i == 23) {
                uuid += '-';
            } else if (i == 12) {
                uuid += hex[dist2(rng)];
            } else {
                uuid += hex[dist(rng)];
            }
        }
        return uuid;
    }

    static std::string stripDashes(const std::string& uuid) {
        std::string result;
        for (char c : uuid) {
            if (c != '-') result += c;
        }
        return result;
    }
};

// ============================================================
// 路径工具
// ============================================================
class PathUtils {
public:
    static std::string join(const std::string& a, const std::string& b) {
        if (a.empty()) return b;
        if (b.empty()) return a;
        char last = a.back();
        char first = b.front();
        if (last == PATH_SEP || last == '/') {
            if (first == PATH_SEP || first == '/') return a + b.substr(1);
            return a + b;
        }
        return a + PATH_SEP + b;
    }

    static std::string getHomeDir() {
#ifdef MC_WINDOWS
        const char* home = getenv("USERPROFILE");
        if (!home) home = getenv("HOME");
        return home ? home : "C:\\Users\\Default";
#else
        const char* home = getenv("HOME");
        return home ? home : "/tmp";
#endif
    }

    static std::string getLauncherDir() {
        // 优先使用环境变量
        const char* env = getenv("MC_LAUNCHER_DIR");
        if (env && env[0]) return env;
        
        // 尝试 exe 目录
        std::string exePath = getExePath();
        if (!exePath.empty()) {
            // 去掉可执行文件名，得到目录
            size_t pos = exePath.find_last_of("/\\");
            if (pos != std::string::npos) {
                std::string exeDir = exePath.substr(0, pos);
                std::string launcherDir = join(exeDir, "launcher");
                if (canWrite(launcherDir)) return launcherDir;
            }
        }
        
        // 默认: ~/.mc-launcher
        return join(getHomeDir(), ".mc-launcher");
    }

    static std::string getExePath() {
#ifdef MC_WINDOWS
        char path[MAX_PATH] = {0};
        GetModuleFileNameA(nullptr, path, MAX_PATH);
        return std::string(path);
#else
        char path[4096] = {0};
        ssize_t len = readlink("/proc/self/exe", path, sizeof(path)-1);
        if (len > 0) {
            path[len] = '\0';
            return std::string(path);
        }
        return "";
#endif
    }

    static bool createDirectory(const std::string& path) {
#ifdef MC_WINDOWS
        return _mkdir(path.c_str()) == 0 || errno == EEXIST;
#else
        return mkdir(path.c_str(), 0755) == 0 || errno == EEXIST;
#endif
    }

    static bool fileExists(const std::string& path) {
#ifdef MC_WINDOWS
        return _access(path.c_str(), 0) == 0;
#else
        return access(path.c_str(), F_OK) == 0;
#endif
    }

    static bool canWrite(const std::string& path) {
#ifdef MC_WINDOWS
        return _access(path.c_str(), 2) == 0; // 2 = WRONLY
#else
        return access(path.c_str(), W_OK) == 0;
#endif
    }

    static std::vector<std::string> listFiles(const std::string& dir, 
                                               const std::string& ext = "") {
        std::vector<std::string> files;
#ifdef MC_WINDOWS
        WIN32_FIND_DATAA findData;
        std::string searchPath = join(dir, "*");
        HANDLE hFind = FindFirstFileA(searchPath.c_str(), &findData);
        if (hFind == INVALID_HANDLE_VALUE) return files;
        do {
            std::string name = findData.cFileName;
            if (name == "." || name == "..") continue;
            if (!ext.empty() && !StringUtils::endsWith(name, ext)) continue;
            files.push_back(join(dir, name));
        } while (FindNextFileA(hFind, &findData));
        FindClose(hFind);
#else
        DIR* dirp = opendir(dir.c_str());
        if (!dirp) return files;
        struct dirent* ent;
        while ((ent = readdir(dirp)) != nullptr) {
            std::string name = ent->d_name;
            if (name == "." || name == "..") continue;
            if (!ext.empty() && !StringUtils::endsWith(name, ext)) continue;
            files.push_back(join(dir, name));
        }
        closedir(dirp);
#endif
        return files;
    }

    static std::string getFileExt(const std::string& path) {
        size_t pos = path.find_last_of('.');
        return pos != std::string::npos ? path.substr(pos) : "";
    }

    static std::string getFileName(const std::string& path) {
        size_t pos = path.find_last_of("/\\");
        return pos != std::string::npos ? path.substr(pos + 1) : path;
    }
};

// ============================================================
// 进程工具
// ============================================================
class ProcessUtils {
public:
    using OutputCallback = std::function<void(const std::string&)>;
    
    static int run(const std::string& cmd, OutputCallback out = nullptr, 
                   OutputCallback err = nullptr, int timeoutSec = 60) {
#ifdef MC_WINDOWS
        // Windows: 使用 CreateProcess
        STARTUPINFOA si = {};
        PROCESS_INFORMATION pi = {};
        si.cb = sizeof(si);
        si.dwFlags = STARTF_USESTDHANDLES;
        
        // 简化实现：使用 system
        std::string cmdCopy = cmd;
        int ret = system(cmdCopy.c_str());
        
        // 设置退出码
        if (ret == -1) {
            if (err) err("Process failed to start");
            return -1;
        }
        return ret;
#else
        // Linux: 使用 pipe 捕获输出
        FILE* pipe = popen(cmd.c_str(), "r");
        if (!pipe) return -1;
        
        char buffer[4096];
        std::string output;
        std::string error;
        
        while (fgets(buffer, sizeof(buffer), pipe)) {
            std::string line(buffer);
            output += line;
            if (out) out(line);
        }
        
        int status = pclose(pipe);
        return WIFEXITED(status) ? WEXITSTATUS(status) : -1;
#endif
    }

    static std::string getVersion(const std::string& javaPath) {
        std::string cmd = "\"" + javaPath + "\" -version 2>&1";
        std::string output;
        
#ifdef MC_WINDOWS
        // Windows
        std::wstring wideCmd = stringToWide(cmd);
        wchar_t* cmdLine = const_cast<wchar_t*>(wideCmd.c_str());
        
        STARTUPINFOW si = {};
        PROCESS_INFORMATION pi = {};
        si.cb = sizeof(si);
        
        if (!CreateProcessW(nullptr, cmdLine, nullptr, nullptr, FALSE, 
                           CREATE_NO_WINDOW | CREATE_UNICODE_ENVIRONMENT,
                           nullptr, nullptr, &si, &pi)) {
            return "unknown";
        }
        
        WaitForSingleObject(pi.hProcess, 5000);
        DWORD exitCode;
        GetExitCodeProcess(pi.hProcess, &exitCode);
        CloseHandle(pi.hProcess);
        CloseHandle(pi.hThread);
        
        return exitCode == 0 ? "17+" : "unknown";
#else
        FILE* pipe = popen(cmd.c_str(), "r");
        if (!pipe) return "unknown";
        
        char buffer[4096];
        std::string versionOutput;
        while (fgets(buffer, sizeof(buffer), pipe)) {
            versionOutput += buffer;
        }
        pclose(pipe);
        
        // 解析版本字符串 "openjdk version \"17.0.1\""
        size_t pos = versionOutput.find("version \"");
        if (pos != std::string::npos) {
            pos += 9;
            size_t end = versionOutput.find('"', pos);
            if (end != std::string::npos) {
                std::string ver = versionOutput.substr(pos, end - pos);
                // 提取主版本号
                size_t dot = ver.find('.');
                if (dot != std::string::npos) {
                    return ver.substr(0, dot);
                }
            }
        }
        return "unknown";
#endif
    }

private:
#ifdef MC_WINDOWS
    static std::wstring stringToWide(const std::string& str) {
        int size = MultiByteToWideChar(CP_UTF8, 0, str.c_str(), -1, nullptr, 0);
        std::wstring result(size, L'\0');
        MultiByteToWideChar(CP_UTF8, 0, str.c_str(), -1, &result[0], size);
        return result;
    }
#endif
};

// ============================================================
// 时间工具
// ============================================================
class TimeUtils {
public:
    static std::string getTimestamp() {
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
        
        std::ostringstream oss;
        oss << std::put_time(&tm_buf, "%Y-%m-%d_%H-%M-%S");
        oss << "_" << std::setfill('0') << std::setw(3) << ms.count();
        return oss.str();
    }

    static std::string formatDate(time_t t) {
        std::tm tm_buf;
#ifdef MC_WINDOWS
        localtime_s(&tm_buf, &t);
#else
        localtime_r(&t, &tm_buf);
#endif
        std::ostringstream oss;
        oss << std::put_time(&tm_buf, "%Y-%m-%d %H:%M:%S");
        return oss.str();
    }
};

// ============================================================
// Zip 工具 (使用 miniz)
// ============================================================
#ifdef MINIZ_H
class ZipUtils {
public:
    static bool extract(const std::string& zipPath, const std::string& destDir) {
        mz_zip_archive zip_archive;
        memset(&zip_archive, 0, sizeof(zip_archive));
        
        if (!mz_zip_reader_init_file(&zip_archive, zipPath.c_str(), 0)) {
            return false;
        }
        
        mz_uint numEntries = mz_zip_reader_get_num_files(&zip_archive);
        for (mz_uint i = 0; i < numEntries; i++) {
            mz_zip_archive_file_stat file_stat;
            if (!mz_zip_reader_file_stat(&zip_archive, i, &file_stat)) continue;
            
            std::string fileName = file_stat.m_filename;
            if (fileName.empty() || fileName.back() == '/') continue;
            
            // 确保目标目录存在
            std::string filePath = join(destDir, fileName);
            createParentDirs(filePath);
            
            // 提取文件
            void* data = nullptr;
            mz_uint32 data_size = mz_zip_reader_extract_to_mem(&zip_archive, i, &data, 0, 0);
            if (data && data_size > 0) {
                std::ofstream ofs(filePath, std::ios::binary);
                if (ofs) {
                    ofs.write(static_cast<char*>(data), data_size);
                }
                free(data);
            }
        }
        
        mz_zip_reader_end(&zip_archive);
        return true;
    }

    static std::vector<std::string> listEntries(const std::string& zipPath) {
        std::vector<std::string> entries;
        mz_zip_archive zip_archive;
        memset(&zip_archive, 0, sizeof(zip_archive));
        
        if (!mz_zip_reader_init_file(&zip_archive, zipPath.c_str(), 0)) {
            return entries;
        }
        
        mz_uint numEntries = mz_zip_reader_get_num_files(&zip_archive);
        for (mz_uint i = 0; i < numEntries; i++) {
            mz_zip_archive_file_stat file_stat;
            if (mz_zip_reader_file_stat(&zip_archive, i, &file_stat)) {
                entries.push_back(file_stat.m_filename);
            }
        }
        
        mz_zip_reader_end(&zip_archive);
        return entries;
    }

private:
    static void createParentDirs(const std::string& path) {
        size_t pos = path.find_last_of("/\\");
        if (pos != std::string::npos) {
            std::string dir = path.substr(0, pos);
            PathUtils::createDirectory(dir);
        }
    }
};
#endif

} // namespace utils
} // namespace mc
