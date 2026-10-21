#include "core/utils.hpp"
#include "core/http.hpp"
#include "core/auth.hpp"
#include "core/java.hpp"
#include "core/instance.hpp"
#include "core/modloader.hpp"
#include <iostream>
#include <string>
#include <vector>

using namespace mc;
using namespace mc::utils;

void printHelp(const char* prog) {
    std::cout << R"(
╔══════════════════════════════════════════════════════╗
║         MC Launcher - 跨平台 Minecraft 启动器         ║
╚══════════════════════════════════════════════════════╝

命令:
  -i, -install <ver>     安装版本
  -l, -launch <ver>      启动游戏
  -ls, -list             列出已安装版本
  -rm, -remove <ver>     删除版本
  -java                  Java 管理
    -java list           列出已安装 Java
    -java install [ver]  安装 Java
  -login                 登录认证
    -login ms            Microsoft 登录
    -login offline [name] 离线登录
  -account               账号管理
    -account list        列出账号
  -mods <ver>            模组管理
    -mods <ver> list     列出模组
    -mods <ver> add <path> 添加模组
  -info                  显示调试信息
  -h, -help              显示帮助
)";
}

int main(int argc, char* argv[]) {
    if (argc < 2) { printHelp(argv[0]); return 0; }

    std::string cmd = argv[1];

    if (cmd == "-h" || cmd == "-help") {
        printHelp(argv[0]);
        return 0;
    }

    if (cmd == "-i" || cmd == "-install") {
        if (argc < 3) { std::cerr << "需要指定版本 ID\n"; return 1; }
        InstanceManager inst;
        return inst.install(argv[2]) ? 0 : 1;
    }

    if (cmd == "-l" || cmd == "-launch") {
        if (argc < 3) { std::cerr << "需要指定版本 ID\n"; return 1; }
        std::string versionId = argv[2];

        JavaManager javaMgr;
        auto java = javaMgr.findJava("17");
        if (!java) {
            std::cerr << "\n错误: 未找到 Java 17+\n请运行: " << argv[0] << " -java install\n";
            return 1;
        }

        AuthManager auth;
        std::string username = "Player";
        std::string uuid = Uuid::generateV4();
        std::string accessToken = uuid;

        std::string gameDir = PathUtils::join(PathUtils::getLauncherDir(), ".minecraft");
        std::string instanceDir = PathUtils::join(PathUtils::getLauncherDir(), "versions", versionId);
        std::string gameJar = instanceDir + "/" + versionId + ".jar";

        std::cout << "启动: " << versionId << std::endl;
        std::cout << "Java: " << java->path << std::endl;
        std::cout << "用户名: " << username << std::endl;

        // 构建命令
        std::string cmd_str = "\"" + java->path + "\"";
        cmd_str += " -Xmx2G -Xms1G";
        cmd_str += " -Dfml.ignoreInvalidMinecraftCertificates=true";
        cmd_str += " -Dhttps.protocols=TLSv1.2";
        cmd_str += " -jar \"" + gameJar + "\"";
        cmd_str += " --username \"" + username + "\"";
        cmd_str += " --version \"" + versionId + "\"";
        cmd_str += " --gameDir \"" + gameDir + "\"";
        cmd_str += " --uuid \"" + uuid + "\"";
        cmd_str += " --accessToken \"" + accessToken + "\"";
        cmd_str += " --userType mock";

        std::cout << "运行: " << cmd_str << "\n";
        return system(cmd_str.c_str());
    }

    if (cmd == "-ls" || cmd == "-list") {
        InstanceManager inst;
        auto versions = inst.listInstalled();
        if (versions.empty()) std::cout << "暂无已安装版本\n";
        else { std::cout << "已安装版本:\n"; for (auto& v : versions) std::cout << "  - " << v << "\n"; }
        return 0;
    }

    if (cmd == "-rm" || cmd == "-remove") {
        if (argc < 3) { std::cerr << "需要指定版本 ID\n"; return 1; }
        InstanceManager inst;
        return inst.uninstall(argv[2]) ? 0 : 1;
    }

    if (cmd == "-java") {
        JavaManager mgr;
        if (argc >= 3) {
            std::string sub = argv[2];
            if (sub == "list") {
                auto javasearch = mgr.listJava();
                if (javasearch.empty()) std::cout << "未找到 Java\n";
                else for (auto& j : javasearch) std::cout << "  " << j.path << " (版本 " << j.version << ")\n";
            } else if (sub == "install") {
                std::string ver = (argc >= 4) ? argv[3] : "17";
                std::cout << "正在安装 Java " << ver << "...\n";
                return mgr.installJava(ver) ? 0 : 1;
            }
        } else {
            std::cout << "Java 管理:\n  -java list\n  -java install [版本]\n";
        }
        return 0;
    }

    if (cmd == "-login") {
        AuthManager auth;
        if (argc >= 3) {
            std::string type = argv[2];
            if (type == "ms") {
                std::cout << "正在获取 Microsoft 登录链接...\n";
                auto session = auth.loginMicrosoft([](const std::string& msg){ std::cout << msg << "\n"; });
                return session ? 0 : 1;
            } else if (type == "offline") {
                std::string name = (argc >= 4) ? argv[3] : "Player";
                auto session = auth.loginOffline(name);
                if (session) {
                    std::cout << "✓ 离线登录成功: " << session->userName << "\n";
                    return 0;
                }
                return 1;
            }
        } else {
            std::cout << "登录方式:\n  -login ms\n  -login offline [用户名]\n";
        }
        return 0;
    }

    if (cmd == "-account") {
        AuthManager auth;
        if (argc >= 3 && argv[2] == std::string("list")) {
            auto accounts = auth.listAccounts();
            if (accounts.empty()) std::cout << "暂无账号\n";
            else for (auto& a : accounts) std::cout << "  - " << a << "\n";
        }
        return 0;
    }

    if (cmd == "-mods") {
        if (argc < 3) { std::cerr << "需要指定版本 ID\n"; return 1; }
        ModManager mods;
        if (argc >= 4) {
            std::string sub = argv[3];
            if (sub == "list") {
                auto m = mods.listMods(argv[2]);
                if (m.empty()) std::cout << "暂无模组\n";
                else for (auto& mod : m) std::cout << "  " << (mod.enabled ? "[✓]" : "[✗]") << " " << mod.name << "\n";
            } else if (sub == "add" && argc >= 5) {
                if (mods.addMod(argv[2], argv[4])) {
                    std::cout << "✓ 已添加模组: " << argv[4] << "\n";
                    return 0;
                }
                return 1;
            }
        } else {
            std::cout << "模组管理:\n  -mods <版本> list\n  -mods <版本> add <路径>\n";
        }
        return 0;
    }

    if (cmd == "-info") {
        std::cout << "MC Launcher 调试信息\n";
        std::cout << "启动器目录: " << PathUtils::getLauncherDir() << "\n";
        std::cout << "可执行文件: " << PathUtils::getExePath() << "\n";
        return 0;
    }

    std::cerr << "未知命令: " << cmd << "\n";
    printHelp(argv[0]);
    return 1;
}
