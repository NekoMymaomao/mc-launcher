#pragma once
#include <string>
#include <sstream>
#include <vector>
#include <map>
#include <cstdint>
#include <cstring>
#include <algorithm>
#include <iomanip>
#include <functional>

#ifdef MC_WINDOWS
    #include <windows.h>
    #include <direct.h>
#else
    #include <unistd.h>
    #include <sys/stat.h>
    #include <dirent.h>
#endif

namespace mc {
namespace utils {

// ============================================================
// 字符串工具
// ============================================================
class StringUtils {
public:
    static std::vector<std::string> split(const std::string& s, char delim) {
        std::vector<std::string> tokens;
        std::stringstream ss(s);
        std::string item;
        while (std::getline(ss, item, delim))
            if (!item.empty()) tokens.push_back(item);
        return tokens;
    }

    static std::string trim(const std::string& s) {
        size_t a = s.find_first_not_of(" \t\n\r");
        size_t b = s.find_last_not_of(" \t\n\r");
        return (a == std::string::npos) ? "" : s.substr(a, b - a + 1);
    }

    static std::string toLower(const std::string& s) {
        std::string r = s;
        std::transform(r.begin(), r.end(), r.begin(), ::tolower);
        return r;
    }

    static bool startsWith(const std::string& s, const std::string& prefix) {
        return s.size() >= prefix.size() && s.compare(0, prefix.size(), prefix) == 0;
    }

    static std::string replaceAll(const std::string& s, const std::string& from, const std::string& to) {
        std::string r = s;
        size_t pos = 0;
        while ((pos = r.find(from, pos)) != std::string::npos) {
            r.replace(pos, from.size(), to);
            pos += to.size();
        }
        return r;
    }

    // Base64 编码
    static std::string base64Encode(const unsigned char* data, size_t len) {
        static const char table[] = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
        std::string r;
        for (size_t i = 0; i < len; i += 3) {
            uint32_t n = 0;
            for (int j = 0; j < 3 && i + j < len; j++)
                n = (n << 8) | data[i + j];
            n <<= (3 - (len - i)) * 8;
            for (int j = 0; j < 4 && i + j < len + (len - i) % 3; j++)
                r += table[(n >> (18 - j * 6)) & 0x3F];
            for (int j = 0; j < (len - i) % 3; j++) r.back() = '=';
        }
        return r;
    }

    // Base64 解码
    static std::string base64Decode(const std::string& s) {
        static const int table[] = {
            -1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,
            -1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,
            -1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,62,-1,-1,-1,63,
            52,53,54,55,56,57,58,59,60,61,-1,-1,-1,-1,-1,-1,
            -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9,10,11,12,13,14,
            15,16,17,18,19,20,21,22,23,-1,-1,-1,-1,-1,-1,
            -1,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,
            41,42,43,44,45,46,47,48,49,50,51,-1,-1,-1,-1,-1
        };
        std::string r;
        uint8_t buf[4];
        int idx = 0;
        for (char c : s) {
            if (c == '=') break;
            int val = c < 128 ? table[c] : -1;
            if (val == -1) continue;
            buf[idx++] = val;
            if (idx == 4) {
                r += (buf[0] << 2) | (buf[1] >> 4);
                if (buf[2] != 63) r += (buf[1] << 4) | (buf[2] >> 2);
                if (buf[3] != 63) r += (buf[2] << 6) | buf[3];
                idx = 0;
            }
        }
        return r;
    }
};

// ============================================================
// SHA-1 实现
// ============================================================
class Sha1 {
public:
    static std::string compute(const std::string& data) {
        uint32_t h0 = 0x67443201, h1 = 0xEF3D669A, h2 = 0x9BDC06E4;
        uint32_t h3 = 0xC1059ED8, h4 = 0x367CD507;

        size_t len = data.size();
        uint64_t bits = len * 8;

        std::vector<uint8_t> msg(len + 1);
        memcpy(msg.data(), data.data(), len);
        msg[len] = 0x80;
        msg.resize(((len + 9) % 64) <= 56 ? len + 65 - (len % 64) : len + 129 - (len % 64));

        for (int i = 56; i < 64; i++) msg[msg.size() - 64 + i - 56] = 0;
        for (int i = 0; i < 8; i++) msg[msg.size() - 8 + i] = (bits >> (56 - i * 8)) & 0xFF;

        for (size_t off = 0; off < msg.size(); off += 64) {
            uint32_t w[64];
            for (int i = 0; i < 16; i++)
                w[i] = (msg[off+i*4] << 24) | (msg[off+i*4+1] << 16) | (msg[off+i*4+2] << 8) | msg[off+i*4+3];
            for (int i = 16; i < 64; i++)
                w[i] = rol(w[i-3] ^ w[i-8] ^ w[i-14] ^ w[i-15], 1);

            uint32_t a=h0,b=h1,c=h2,d=h3,e=h4;
            for (int i = 0; i < 64; i++) {
                uint32_t f,k;
                if (i < 20) { f=(b&c)|(~b&d); k=0x5A827999; }
                else if (i < 40) { f=b^c^d; k=0x6ED9EBA1; }
                else if (i < 60) { f=(b&c)|(b&d)|(c&d); k=0x8F1BBCDC; }
                else { f=b^c^d; k=0xCA62C1D6; }
                uint32_t t=rol(a,5)+f+e+k+w[i];
                e=d; d=c; c=rol(b,30); b=a; a=t;
            }
            h0+=a; h1+=b; h2+=c; h3+=d; h4+=e;
        }

        std::ostringstream os;
        os << std::hex << std::setfill('0')
           << std::setw(8) << h0 << std::setw(8) << h1 << std::setw(8) << h2
           << std::setw(8) << h3 << std::setw(8) << h4;
        return os.str();
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
        static thread_local std::mt19937 rng(std::chrono::steady_clock::now().time_since_epoch().count());
        std::uniform_int_distribution<int> dist(0, 15);
        std::uniform_int_distribution<int> dist2(8, 11);
        std::string h = "0123456789abcdef";
        std::string uuid;
        for (int i = 0; i < 36; i++) {
            if (i == 8||i==13||i==18||i==23) { uuid += '-'; }
            else if (i == 12) { uuid += h[dist2(rng)]; }
            else { uuid += h[dist(rng)]; }
        }
        return uuid;
    }

    static std::string stripDashes(const std::string& u) {
        std::string r;
        for (char c : u) if (c != '-') r += c;
        return r;
    }
};

// ============================================================
// 路径工具
// ============================================================
class PathUtils {
public:
#ifdef MC_WINDOWS
    static constexpr char SEP = '\\';
#else
    static constexpr char SEP = '/';
#endif

    static std::string join(const std::string& a, const std::string& b) {
        if (a.empty()) return b;
        if (b.empty()) return a;
        char last = a.back(), first = b.front();
        if (last == SEP || last == '/') {
            if (first == SEP || first == '/') return a + b.substr(1);
            return a + b;
        }
        return a + SEP + b;
    }

    static std::string getHomeDir() {
#ifdef MC_WINDOWS
        const char* h = getenv("USERPROFILE");
        return h ? h : "C:\\Users\\Default";
#else
        const char* h = getenv("HOME");
        return h ? h : "/tmp";
#endif
    }

    static std::string getLauncherDir() {
        const char* env = getenv("MC_LAUNCHER_DIR");
        if (env && env[0]) return env;
        std::string exe = getExePath();
        if (!exe.empty()) {
            size_t p = exe.find_last_of("/\\");
            if (p != std::string::npos) {
                std::string dir = exe.substr(0, p);
                std::string ld = join(dir, "launcher");
                if (canWrite(ld)) return ld;
            }
        }
        return join(getHomeDir(), ".mc-launcher");
    }

    static std::string getExePath() {
#ifdef MC_WINDOWS
        char p[MAX_PATH] = {0};
        GetModuleFileNameA(nullptr, p, MAX_PATH);
        return std::string(p);
#else
        char p[4096] = {0};
        ssize_t len = readlink("/proc/self/exe", p, sizeof(p)-1);
        return (len > 0) ? std::string(p, len) : "";
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
        return _access(path.c_str(), 2) == 0;
#else
        return access(path.c_str(), W_OK) == 0;
#endif
    }

    static std::vector<std::string> listFiles(const std::string& dir, const std::string& ext = "") {
        std::vector<std::string> files;
#ifdef MC_WINDOWS
        WIN32_FIND_DATAA fd;
        HANDLE h = FindFirstFileA(join(dir, "*").c_str(), &fd);
        if (h == INVALID_HANDLE_VALUE) return files;
        do {
            std::string name = fd.cFileName;
            if (name == "." || name == "..") continue;
            if (!ext.empty() && !endsWith(name, ext)) continue;
            files.push_back(join(dir, name));
        } while (FindNextFileA(h, &fd));
        FindClose(h);
#else
        DIR* d = opendir(dir.c_str());
        if (!d) return files;
        struct dirent* ent;
        while ((ent = readdir(d))) {
            std::string name = ent->d_name;
            if (name == "." || name == "..") continue;
            if (!ext.empty() && !endsWith(name, ext)) continue;
            files.push_back(join(dir, name));
        }
        closedir(d);
#endif
        return files;
    }

    static std::string getFileName(const std::string& path) {
        size_t p = path.find_last_of("/\\");
        return p != std::string::npos ? path.substr(p + 1) : path;
    }

private:
    static bool endsWith(const std::string& s, const std::string& suffix) {
        return s.size() >= suffix.size() && s.compare(s.size()-suffix.size(), suffix.size(), suffix) == 0;
    }
};

// ============================================================
// 进程工具
// ============================================================
class ProcessUtils {
public:
    static int run(const std::string& cmd) {
#ifdef MC_WINDOWS
        return system(cmd.c_str());
#else
        return system(cmd.c_str());
#endif
    }

    static std::string getVersion(const std::string& javaPath) {
        std::string cmd = "\"" + javaPath + "\" -version 2>&1";
        FILE* f = popen(cmd.c_str(), "r");
        if (!f) return "unknown";
        char buf[4096];
        std::string out;
        while (fgets(buf, sizeof(buf), f)) out += buf;
        pclose(f);

        size_t pos = out.find("version \"");
        if (pos == std::string::npos) return "unknown";
        pos += 9;
        size_t end = out.find('"', pos);
        if (end == std::string::npos) return "unknown";
        std::string ver = out.substr(pos, end - pos);
        size_t dot = ver.find('.');
        return (dot != std::string::npos) ? ver.substr(0, dot) : "unknown";
    }
};

// ============================================================
// 简单 JSON 解析器
// ============================================================
class Json {
public:
    enum Type { NULL_TYPE, BOOL_TYPE, NUMBER_TYPE, STRING_TYPE, ARRAY_TYPE, OBJECT_TYPE };

    class Value {
    public:
        Type type = NULL_TYPE;
        bool bool_val = false;
        double num_val = 0.0;
        std::string str_val;
        std::vector<Value> arr_val;
        std::map<std::string, Value> obj_val;

        Value() = default;
        Value(bool b) : type(BOOL_TYPE), bool_val(b) {}
        Value(double n) : type(NUMBER_TYPE), num_val(n) {}
        Value(const std::string& s) : type(STRING_TYPE), str_val(s) {}
        Value(const char* s) : type(STRING_TYPE), str_val(s) {}

        bool isNull() const { return type == NULL_TYPE; }
        bool isBool() const { return type == BOOL_TYPE; }
        bool isNumber() const { return type == NUMBER_TYPE; }
        bool isString() const { return type == STRING_TYPE; }
        bool isArray() const { return type == ARRAY_TYPE; }
        bool isObject() const { return type == OBJECT_TYPE; }

        bool getBool(bool def = false) const { return isBool() ? bool_val : def; }
        double getNumber(double def = 0.0) const { return isNumber() ? num_val : def; }
        std::string getString(const std::string& def = "") const { return isString() ? str_val : def; }

        Value& operator[](size_t idx) { return arr_val[idx]; }
        const Value& operator[](size_t idx) const { return arr_val[idx]; }
        Value& operator[](const std::string& key) { return obj_val[key]; }
        const Value& operator[](const std::string& key) const {
            static Value null_val;
            auto it = obj_val.find(key);
            return it != obj_val.end() ? it->second : null_val;
        }

        bool contains(const std::string& key) const {
            return type == OBJECT_TYPE && obj_val.count(key) > 0;
        }

        size_t size() const {
            if (type == ARRAY_TYPE) return arr_val.size();
            if (type == OBJECT_TYPE) return obj_val.size();
            return 0;
        }

        std::string dump(int indent = 0) const {
            std::ostringstream os;
            dumpInternal(os, indent, 0);
            return os.str();
        }

    private:
        void dumpInternal(std::ostringstream& os, int indent, int depth) const {
            std::string pad(indent * depth, ' ');
            std::string pad1(indent * (depth + 1), ' ');
            switch (type) {
                case NULL_TYPE: os << "null"; break;
                case BOOL_TYPE: os << (bool_val ? "true" : "false"); break;
                case NUMBER_TYPE: os << num_val; break;
                case STRING_TYPE: os << "\"" << escapeString(str_val) << "\""; break;
                case ARRAY_TYPE:
                    os << "[\n";
                    for (size_t i = 0; i < arr_val.size(); i++) {
                        os << pad1 << arr_val[i].dumpInternal(indent, depth + 1);
                        if (i < arr_val.size() - 1) os << ",";
                        os << "\n";
                    }
                    os << pad << "]";
                    break;
                case OBJECT_TYPE:
                    os << "{\n";
                    size_t i = 0;
                    for (auto& [k, v] : obj_val) {
                        os << pad1 << "\"" << escapeString(k) << "\": " << v.dumpInternal(indent, depth + 1);
                        if (++i < obj_val.size()) os << ",";
                        os << "\n";
                    }
                    os << pad << "}";
                    break;
            }
        }

        static std::string escapeString(const std::string& s) {
            std::string r;
            for (char c : s) {
                switch (c) {
                    case '"': r += "\\\""; break;
                    case '\\': r += "\\\\"; break;
                    case '\n': r += "\\n"; break;
                    case '\r': r += "\\r"; break;
                    case '\t': r += "\\t"; break;
                    default: r += c;
                }
            }
            return r;
        }
    };

    static Value parse(const std::string& s) {
        SizeT pos = 0;
        return parseValue(s, pos);
    }

private:
    using SizeT = size_t;

    static Value parseValue(const std::string& s, SizeT& pos) {
        skipWhitespace(s, pos);
        if (pos >= s.size()) return Value();

        char c = s[pos];
        if (c == '"') return parseString(s, pos);
        if (c == '{') return parseObject(s, pos);
        if (c == '[') return parseArray(s, pos);
        if (c == 't' || c == 'f') return parseBool(s, pos);
        if (c == 'n') return parseNull(s, pos);
        if (c == '-' || (c >= '0' && c <= '9')) return parseNumber(s, pos);

        return Value();
    }

    static void skipWhitespace(const std::string& s, SizeT& pos) {
        while (pos < s.size() && (s[pos] == ' ' || s[pos] == '\t' || s[pos] == '\n' || s[pos] == '\r')) pos++;
    }

    static Value parseString(const std::string& s, SizeT& pos) {
        if (s[pos] != '"') return Value();
        pos++;
        std::string r;
        while (pos < s.size() && s[pos] != '"') {
            if (s[pos] == '\\' && pos + 1 < s.size()) {
                pos++;
                switch (s[pos]) {
                    case '"': r += '"'; break;
                    case '\\': r += '\\'; break;
                    case '/': r += '/'; break;
                    case 'n': r += '\n'; break;
                    case 'r': r += '\r'; break;
                    case 't': r += '\t'; break;
                    case 'u':
                        if (pos + 4 < s.size()) {
                            std::string hex = s.substr(pos + 1, 4);
                            uint32_t cp = std::stoul(hex, nullptr, 16);
                            if (cp < 0x80) r += (char)cp;
                            else if (cp < 0x800) {
                                r += (char)(0xC0 | (cp >> 6));
                                r += (char)(0x80 | (cp & 0x3F));
                            } else {
                                r += (char)(0xE0 | (cp >> 12));
                                r += (char)(0x80 | ((cp >> 6) & 0x3F));
                                r += (char)(0x80 | (cp & 0x3F));
                            }
                            pos += 4;
                        }
                        break;
                    default: r += s[pos];
                }
            } else {
                r += s[pos];
            }
            pos++;
        }
        if (pos < s.size()) pos++;
        return Value(r);
    }

    static Value parseNumber(const std::string& s, SizeT& pos) {
        SizeT start = pos;
        if (s[pos] == '-') pos++;
        while (pos < s.size() && s[pos] >= '0' && s[pos] <= '9') pos++;
        bool frac = false;
        if (pos < s.size() && s[pos] == '.') { frac = true; pos++; while (pos < s.size() && s[pos] >= '0' && s[pos] <= '9') pos++; }
        bool exp = false;
        if (pos < s.size() && (s[pos] == 'e' || s[pos] == 'E')) { exp = true; pos++; if (pos < s.size() && (s[pos] == '+' || s[pos] == '-')) pos++; while (pos < s.size() && s[pos] >= '0' && s[pos] <= '9') pos++; }
        return Value(std::stod(s.substr(start, pos - start)));
    }

    static Value parseBool(const std::string& s, SizeT& pos) {
        if (s.compare(pos, 4, "true") == 0) { pos += 4; return Value(true); }
        if (s.compare(pos, 5, "false") == 0) { pos += 5; return Value(false); }
        return Value();
    }

    static Value parseNull(const std::string& s, SizeT& pos) {
        if (s.compare(pos, 4, "null") == 0) { pos += 4; return Value(); }
        return Value();
    }

    static Value parseArray(const std::string& s, SizeT& pos) {
        if (s[pos] != '[') return Value();
        pos++;
        Value arr;
        arr.type = ARRAY_TYPE;
        skipWhitespace(s, pos);
        if (pos < s.size() && s[pos] == ']') { pos++; return arr; }
        while (pos < s.size()) {
            arr.arr_val.push_back(parseValue(s, pos));
            skipWhitespace(s, pos);
            if (pos < s.size() && s[pos] == ',') { pos++; skipWhitespace(s, pos); }
            else break;
        }
        if (pos < s.size() && s[pos] == ']') pos++;
        return arr;
    }

    static Value parseObject(const std::string& s, SizeT& pos) {
        if (s[pos] != '{') return Value();
        pos++;
        Value obj;
        obj.type = OBJECT_TYPE;
        skipWhitespace(s, pos);
        if (pos < s.size() && s[pos] == '}') { pos++; return obj; }
        while (pos < s.size()) {
            Value key = parseString(s, pos);
            skipWhitespace(s, pos);
            if (pos < s.size() && s[pos] == ':') pos++;
            skipWhitespace(s, pos);
            Value val = parseValue(s, pos);
            obj.obj_val[key.str_val] = val;
            skipWhitespace(s, pos);
            if (pos < s.size() && s[pos] == ',') { pos++; skipWhitespace(s, pos); }
            else break;
        }
        if (pos < s.size() && s[pos] == '}') pos++;
        return obj;
    }
};

} // namespace utils
} // namespace mc
