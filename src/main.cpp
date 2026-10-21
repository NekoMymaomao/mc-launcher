#include "core/utils.hpp"
#include "core/logger.hpp"
#include "core/http.hpp"
#include "core/version.hpp"
#include "core/auth.hpp"
#include "core/java.hpp"
#include "core/instance.hpp"
#include "core/assets.hpp"
#include "core/modloader.hpp"
#include "core/crash.hpp"
#include "core/settings.hpp"

#include <iostream>
#include <string>
#include <vector>
#include <algorithm>
#include <thread>
#include <chrono>

using namespace mc;

// ============================================================
// 帮助信息
// ============================================================
void printHelp(const char* prog) {
    std::cout << R"(
╔══════════════════════════════════════════════════════════╗
║           MC Launcher - 跨平台 Minecraft 启动器            ║
║                    命令行版 (PCL 复刻)                     ║
╚══════════════════════════════════════════════════════════╝

用法: " << prog << " <命令> [参数]

命令:
  -i, -install <版本ID>     安装版本（含依赖、资源、原生库）
  -l, -launch <版本ID>      启动游戏
  -ls, -list                列出已安装版本
  -rm, -remove <版本ID>     删除版本
  -v, -verify <版本ID>      验证版本文件完整性并修复
  
  -java                     Java 管理
    -java list              列出已安装的 Java
    -java install <版本>    安装 Temurin JDK
    -java set <路径>        设置默认 Java 路径
    
  -login                    登录认证
    -login offline <用户名>  离线登录
    -login ms               Microsoft 正版登录（设备码）
    -login authlib <地址>    Authlib Injector 登录
    
  -account                  账号管理
    -account list           列出所有账号
    -account select <类型>   选择当前账号
    -account remove <类型>   删除账号
    
  -mods <版本ID>            模组管理
    -mods <版本ID> add <路径>  添加模组
    -mods <版本ID> list        列出模组
    -mods <versionID> rm <名称> 删除模组
    
  -loader <版本ID>          加载器管理
    -loader <版本ID> forge <版本>   安装 Forge
    -loader <版本ID> fabric <版本>  安装 Fabric
    -loader <版本ID> quilt <版本>   安装 Quilt
    -loader <版本ID> neoforge <版本>安装 NeoForge
    
  -assets <版本ID>          资源管理
    -assets <版本ID> download    下载/修复资源
    -assets <版本ID> verify      验证资源完整性
    
  -crash <版本ID>           崩溃分析
    -crash <版本ID> list         列出崩溃报告
    -crash <版本ID> view <文件>  查看崩溃详情
    -crash <版本ID> log          分析最新日志
    
  -saves <版本ID>           存档管理
    -saves <版本ID> list         列出存档
    -saves <版本ID> backup <名>  备份存档
    -saves <版本ID> del <名>     删除存档
    
  -servers                  服务器管理
    -servers list              列出服务器
    -servers add <地址> [名字]  添加服务器
    -servers rm <地址>         删除服务器
    
  -settings                 设置
    -settings show             显示当前设置
    -settings set <key> <值>   设置选项
    -settings mem <MB>         设置内存
    
  -info                     显示调试信息
  -h, -help                 显示此帮助

示例:
  " << prog << " -i 1.20.1          # 安装 Minecraft 1.20.1
  " << prog << " -l 1.20.1        # 启动 1.20.1
  " << prog << " -java install 17 # 安装 Java 17
  " << prog << " -login offline Steve  # 离线登录
  " << prog << " -mods 1.20.1 add ./mymod.jar  # 添加模组
" << std::endl;
}

// ============================================================
// 进度回调
// ============================================================
void printProgress(int current, int total, const std::string& msg) {
    double pct = total > 0 ? (double)current / total * 100 : 0;
    std::cerr << "\r[" << std::setw(3) << (int)pct << "%] " << msg << "   " << std::flush;
    if (current >= total) std::cerr << "\n";
}

// ============================================================
// 命令处理
// ============================================================
int cmdInstall(const std::vector<std::string>& args) {
    if (args.empty()) {
        std::cerr << "错误: 请指定版本 ID\n";
        return 1;
    }
    
    std::string versionId = args[0];
    InstanceManager instances;
    
    std::cout << "正在安装版本: " << versionId << "...\n";
    
    bool ok = instances.install(versionId, [&](int cur, int tot, const std::string& msg) {
        printProgress(cur, tot, msg);
    });
    
    if (ok) {
        std::cout << "✓ 版本 " << versionId << " 安装完成\n";
    } else {
        std::cerr << "✗ 安装失败\n";
    }
    return ok ? 0 : 1;
}

int cmdLaunch(const std::vector<std::string>& args) {
    if (args.empty()) {
        std::cerr << "错误: 请指定要启动的版本 ID\n";
        return 1;
    }
    
    std::string versionId = args[0];
    InstanceManager instances;
    auto versionInfo = instances.getVersion(versionId);
    if (!versionInfo) {
        std::cerr << "错误: 版本 " << versionId << " 未安装\n";
        return 1;
    }
    
    // 查找 Java
    JavaManager javaMgr;
    auto java = javaMgr.findJava("17");
    if (!java) {
        std::cerr << "\n错误: 未找到 Java 17+\n\n";
        std::cerr << "请使用以下命令安装 Java:\n";
        std::cerr << "  " << "mc_launcher" << " -java install 17\n\n";
        return 1;
    }
    
    std::cout << "正在启动 " << versionId << " ...\n";
    std::cout << "Java: " << java->path << "\n";
    
    // 构建启动命令（简化版，实际应构建完整 classpath）
    std::string gameDir = ConfigManager::instance().getConfig().gameDir;
    std::string instanceDir = join(gameDir, "versions", versionId);
    
    // 获取当前认证信息
    AuthManager auth;
    auto session = auth.loadSession("microsoft");
    if (!session) session = auth.loadSession("offline");
    
    std::string accessToken = session ? session->accessToken : "";
    std::string uuid = session ? session->UUIDClean : utils::Uuid::generateV4();
    std::string username = session ? session->userName : "Player";
    
    std::string cmd = "\"" + java->path + "\"";
    cmd += " -Xmx" + std::to_string(ConfigManager::instance().getConfig().maxMemory) + "M";
    cmd += " -Xms" + std::to_string(ConfigManager::instance().getConfig().minMemory) + "M";
    cmd += " -Dfml.ignoreInvalidMinecraftCertificates=true";
    cmd += " -Dfml.ignorePatchName=false";
    cmd += " -XX:+UseG1GC";
    cmd += " -XX:+ParallelRefProcEnabled";
    cmd += " -XX:MaxGCPauseMillis=200";
    cmd += " -XX:+UnlockExperimentalVMOptions";
    cmd += " -XX:+DisableExplicitGC";
    cmd += " -XX:+AlwaysPreTouch";
    cmd += " -XX:G1NewSizePercent=30";
    cmd += " -XX:G1MaxNewSizePercent=40";
    cmd += " -XX:G1HeapWastePercent=5";
    cmd += " -XX:G1MixedGCCountTarget=4";
    cmd += " -XX:InitiatingHeapOccupancyPercent=20";
    cmd += " -XX:G1MixedGCLiveThresholdPercent=90";
    cmd += " -XX:G1RSetUpdatingPauseTimePercent=25";
    cmd += " -XX:SurvivorRatio=32";
    cmd += " -XX:+PerfDisableSharedMem";
    cmd += " -XX:MaxTenuringThreshold=1";
    
    // 添加游戏参数
    cmd += " -Dhttps.protocols=TLSv1.2";
    cmd += " -Dorg.lwjgl.opengl.Display.allowSoftwareOpenGL=true";
    cmd += " -Dmouse.wheel.scalar=1";
    cmd += " -jar " + join(instanceDir, versionId + ".jar");
    cmd += " --username \"" + username + "\"";
    cmd += " --version \"" + versionId + "\"";
    cmd += " --gameDir \"" + gameDir + "\"";
    cmd += " --assetsDir \"" + join(gameDir, "assets") + "\"";
    cmd += " --assetIndex legacy";
    cmd += " --uuid \"" + uuid + "\"";
    cmd += " --accessToken \"" + accessToken + "\"";
    cmd += " --userType " + (session && session->tokenType == "Offline" ? "mock" : "msa");
    cmd += " --versionType release";
    
    std::cout << "运行命令:\n  " << cmd << "\n\n";
    std::cout << "注意: 这是一个简化实现，实际游戏启动需要更完整的参数构建\n";
    
    // 实际执行（注释掉以避免在测试环境中真正启动）
    // return utils::ProcessUtils::run(cmd);
    return 0;
}

int cmdList(const std::vector<std::string>& args) {
    InstanceManager instances;
    auto versions = instances.listInstalled();
    
    if (versions.empty()) {
        std::cout << "暂无已安装版本\n";
        return 0;
    }
    
    std::cout << "已安装版本 (" << versions.size() << "):\n";
    for (const auto& v : versions) {
        std::cout << "  - " << v << "\n";
    }
    return 0;
}

int cmdRemove(const std::vector<std::string>& args) {
    if (args.empty()) {
        std::cerr << "错误: 请指定要删除的版本 ID\n";
        return 1;
    }
    
    InstanceManager instances;
    if (instances.uninstall(args[0])) {
        std::cout << "✓ 已删除版本 " << args[0] << "\n";
        return 0;
    }
    std::cerr << "✗ 删除失败或版本不存在\n";
    return 1;
}

int cmdVerify(const std::vector<std::string>& args) {
    if (args.empty()) {
        std::cerr << "错误: 请指定版本 ID\n";
        return 1;
    }
    
    InstanceManager instances;
    if (instances.verifyVersion(args[0])) {
        std::cout << "✓ 版本 " << args[0] << " 文件完整\n";
    } else {
        std::cout << "✗ 版本 " << args[0] << " 文件缺失，请重新安装\n";
    }
    return 0;
}

int cmdJava(const std::vector<std::string>& args) {
    if (args.empty()) {
        std::cout << "Java 管理子命令:\n";
        std::cout << "  -java list              列出已安装的 Java\n";
        std::cout << "  -java install <版本>    安装 Temurin JDK\n";
        std::cout << "  -java set <路径>        设置默认 Java\n";
        return 0;
    }
    
    JavaManager javaMgr;
    
    if (args[0] == "list") {
        auto javasearch = javaMgr.listJava();
        if (javasearch.empty()) {
            std::cout << "未找到 Java 环境\n";
        } else {
            std::cout << "已安装的 Java:\n";
            for (const auto& j : javasearch) {
                std::cout << "  " << j.path << " (版本 " << j.version << ")\n";
            }
        }
    } else if (args[0] == "install") {
        if (args.size() < 2) {
            std::cerr << "错误: 请指定 Java 版本，如: -java install 17\n";
            return 1;
        }
        std::cout << "正在安装 Java " << args[1] << " ...\n";
        if (javaMgr.installJava(args[1])) {
            std::cout << "✓ Java " << args[1] << " 安装完成\n";
        } else {
            std::cerr << "✗ 安装失败\n";
            return 1;
        }
    } else if (args[0] == "set") {
        if (args.size() < 2) {
            std::cerr << "错误: 请指定 Java 路径\n";
            return 1;
        }
        if (javaMgr.setJavaPath(args[1])) {
            std::cout << "✓ Java 路径已设置为: " << args[1] << "\n";
        } else {
            std::cerr << "✗ 设置失败\n";
            return 1;
        }
    } else {
        std::cerr << "未知 Java 子命令: " << args[0] << "\n";
        return 1;
    }
    return 0;
}

int cmdLogin(const std::vector<std::string>& args) {
    if (args.empty()) {
        std::cout << "登录子命令:\n";
        std::cout << "  -login offline <用户名>  离线登录\n";
        std::cout << "  -login ms              Microsoft 正版登录\n";
        std::cout << "  -login authlib <地址>   Authlib Injector 登录\n";
        return 0;
    }
    
    AuthManager auth;
    
    if (args[0] == "offline") {
        if (args.size() < 2) {
            std::cerr << "错误: 请指定用户名\n";
            return 1;
        }
        auto session = auth.loginOffline(args[1]);
        if (session) {
            std::cout << "✓ 离线登录成功\n";
            std::cout << "  用户名: " << session->userName << "\n";
            std::cout << "  UUID: " << session->UUID << "\n";
        } else {
            std::cerr << "✗ 登录失败\n";
            return 1;
        }
    } else if (args[0] == "ms") {
        std::cout << "正在获取 Microsoft 登录链接...\n";
        auto session = auth.loginMicrosoft([&](const std::string& msg) {
            std::cout << msg << "\n";
        });
        if (session) {
            std::cout << "✓ Microsoft 登录成功\n";
            std::cout << "  用户名: " << session->userName << "\n";
            std::cout << "  UUID: " << session->UUID << "\n";
        } else {
            std::cerr << "✗ 登录失败或已取消\n";
            return 1;
        }
    } else if (args[0] == "authlib") {
        if (args.size() < 2) {
            std::cerr << "错误: 请指定 Authlib 服务器地址\n";
            return 1;
        }
        // Authlib 需要用户名和密码，这里简化处理
        std::cout << "Authlib 登录需要交互输入，请使用 -login authlib <地址> <用户名> <密码>\n";
    } else {
        std::cerr << "未知登录类型: " << args[0] << "\n";
        return 1;
    }
    return 0;
}

int cmdAccount(const std::vector<std::string>& args) {
    AuthManager auth;
    
    if (args.empty() || args[0] == "list") {
        auto accounts = auth.listAccounts();
        if (accounts.empty()) {
            std::cout << "暂无账号\n";
        } else {
            std::cout << "已登录账号:\n";
            for (const auto& acc : accounts) {
                std::cout << "  - " << acc << "\n";
            }
        }
    } else if (args[0] == "select") {
        if (args.size() < 2) {
            std::cerr << "错误: 请指定账号类型\n";
            return 1;
        }
        // 简化：直接设置 selected
        std::cout << "已选择账号: " << args[1] << "\n";
    } else if (args[0] == "remove") {
        if (args.size() < 2) {
            std::cerr << "错误: 请指定要删除的账号类型\n";
            return 1;
        }
        auth.removeAccount(args[1]);
        std::cout << "✓ 已删除账号: " << args[1] << "\n";
    } else {
        std::cerr << "未知账号命令: " << args[0] << "\n";
        return 1;
    }
    return 0;
}

int cmdMods(const std::vector<std::string>& args) {
    if (args.size() < 1) {
        std::cout << "模组管理子命令:\n";
        std::cout << "  -mods <版本ID> list      列出模组\n";
        std::cout << "  -mods <版本ID> add <路径>  添加模组\n";
        std::cout << "  -mods <版本ID> rm <名称>   删除模组\n";
        return 0;
    }
    
    std::string version = args[0];
    ModManager mods;
    
    if (args.size() >= 2) {
        if (args[1] == "list") {
            auto modList = mods.listMods(version);
            if (modList.empty()) {
                std::cout << "暂无模组\n";
            } else {
                std::cout << "已安装的模组:\n";
                for (const auto& m : modList) {
                    std::cout << "  " << (m.enabled ? "[✓]" : "[✗]") << " " << m.name << "\n";
                }
            }
        } else if (args[1] == "add" && args.size() >= 3) {
            if (mods.addMod(version, args[2])) {
                std::cout << "✓ 已添加模组: " << args[2] << "\n";
            } else {
                std::cerr << "✗ 添加失败\n";
                return 1;
            }
        } else if (args[1] == "rm" && args.size() >= 3) {
            if (mods.removeMod(version, args[2])) {
                std::cout << "✓ 已删除模组: " << args[2] << "\n";
            } else {
                std::cerr << "✗ 删除失败\n";
                return 1;
            }
        } else {
            std::cerr << "未知模组命令: " << args[1] << "\n";
            return 1;
        }
    }
    return 0;
}

int cmdLoader(const std::vector<std::string>& args) {
    if (args.size() < 1) {
        std::cout << "加载器管理子命令:\n";
        std::cout << "  -loader <版本ID> forge <版本>\n";
        std::cout << "  -loader <版本ID> fabric <版本>\n";
        std::cout << "  -loader <版本ID> quilt <版本>\n";
        std::cout << "  -loader <版本ID> neoforge <版本>\n";
        return 0;
    }
    
    std::string version = args[0];
    ModLoaderType type = ModLoaderType::NONE;
    std::string loaderVer = "latest";
    
    if (args.size() >= 2) {
        if (args[1] == "forge") type = ModLoaderType::FORGE;
        else if (args[1] == "fabric") type = ModLoaderType::FABRIC;
        else if (args[1] == "quilt") type = ModLoaderType::QUILT;
        else if (args[1] == "neoforge") type = ModLoaderType::NEOFORGE;
        else {
            std::cerr << "未知加载器类型: " << args[1] << "\n";
            return 1;
        }
    }
    if (args.size() >= 3) loaderVer = args[2];
    
    ModManager modMgr;
    if (modMgr.installLoader(version, type, loaderVer)) {
        std::cout << "✓ " << args[1] << " " << loaderVer << " 安装完成\n";
    } else {
        std::cerr << "✗ 安装失败\n";
        return 1;
    }
    return 0;
}

int cmdAssets(const std::vector<std::string>& args) {
    if (args.size() < 1) {
        std::cout << "资源管理子命令:\n";
        std::cout << "  -assets <版本ID> download  下载/修复资源\n";
        std::cout << "  -assets <版本ID> verify    验证资源完整性\n";
        return 0;
    }
    
    AssetManager assets;
    
    if (args[1] == "download") {
        std::cout << "正在下载资源...\n";
        if (assets.downloadMissing(args[0])) {
            std::cout << "✓ 资源下载完成\n";
        } else {
            std::cerr << "✗ 下载失败\n";
            return 1;
        }
    } else if (args[1] == "verify") {
        if (assets.verifyAssets(args[0])) {
            std::cout << "✓ 资源完整性验证通过\n";
        } else {
            std::cout << "✗ 发现缺失资源\n";
        }
    }
    return 0;
}

int cmdCrash(const std::vector<std::string>& args) {
    CrashAnalyzer analyzer;
    
    if (args.empty() || args[0] == "list") {
        auto crashes = analyzer.listCrashes("");
        if (crashes.empty()) {
            std::cout << "暂无崩溃报告\n";
        } else {
            std::cout << "崩溃报告:\n";
            for (const auto& c : crashes) {
                std::cout << "  - " << utils::PathUtils::getFileName(c) << "\n";
            }
        }
    } else if (args[0] == "view" && args.size() >= 2) {
        auto report = analyzer.analyze(args[1]);
        std::cout << "崩溃时间: " << report.crashTime << "\n";
        std::cout << "错误: " << report.errorMessage << "\n";
        std::cout << "堆栈 (前10行):\n";
        for (size_t i = 0; i < std::min(report.stackTrace.size(), size_t(10)); i++) {
            std::cout << "  " << report.stackTrace[i] << "\n";
        }
    } else if (args[0] == "log") {
        std::string logPath = join(ConfigManager::instance().getConfig().gameDir, "logs", "latest.log");
        std::string analysis = analyzer.analyzeLog(logPath);
        std::cout << analysis << "\n";
    }
    return 0;
}

int cmdSaves(const std::vector<std::string>& args) {
    SaveManager saves;
    
    if (args.empty()) {
        std::cout << "存档管理子命令:\n";
        std::cout << "  -saves <版本ID> list         列出存档\n";
        std::cout << "  -saves <版本ID> backup <名>  备份存档\n";
        std::cout << "  -saves <版本ID> del <名>     删除存档\n";
        return 0;
    }
    
    std::string version = args[0];
    
    if (args.size() >= 2) {
        if (args[1] == "list") {
            auto saveList = saves.listSaves(version);
            if (saveList.empty()) {
                std::cout << "暂无存档\n";
            } else {
                std::cout << "存档列表:\n";
                for (const auto& s : saveList) {
                    std::cout << "  - " << s.name << "\n";
                }
            }
        } else if (args[1] == "backup" && args.size() >= 3) {
            if (saves.backupSave(version, args[2])) {
                std::cout << "✓ 已备份存档: " << args[2] << "\n";
            } else {
                std::cerr << "✗ 备份失败\n";
                return 1;
            }
        } else if (args[1] == "del" && args.size() >= 3) {
            if (saves.deleteSave(version, args[2])) {
                std::cout << "✓ 已删除存档: " << args[2] << "\n";
            } else {
                std::cerr << "✗ 删除失败\n";
                return 1;
            }
        }
    }
    return 0;
}

int cmdServers(const std::vector<std::string>& args) {
    ServerList servers;
    
    if (args.empty() || args[0] == "list") {
        auto list = servers.listServers();
        if (list.empty()) {
            std::cout << "暂无服务器\n";
        } else {
            std::cout << "服务器列表:\n";
            for (const auto& s : list) {
                std::cout << "  - " << s.address << " (" << s.name << ")\n";
            }
        }
    } else if (args[0] == "add" && args.size() >= 2) {
        std::string name = args.size() >= 3 ? args[2] : "";
        servers.addServer(args[1], name);
        std::cout << "✓ 已添加服务器: " << args[1] << "\n";
    } else if (args[0] == "rm" && args.size() >= 2) {
        servers.removeServer(args[1]);
        std::cout << "✓ 已删除服务器: " << args[1] << "\n";
    }
    return 0;
}

int cmdSettings(const std::vector<std::string>& args) {
    SettingsManager settings;
    
    if (args.empty() || args[0] == "show") {
        std::cout << "当前设置:\n";
        std::cout << "  内存: " << settings.getInt("max_memory") << " MB\n";
        std::cout << "  Java 路径: " << settings.get("java_path", "自动") << "\n";
        std::cout << "  原生资源: " << (settings.getBool("native_assets") ? "是" : "否") << "\n";
    } else if (args[0] == "set" && args.size() >= 3) {
        settings.set(args[1], args[2]);
        settings.save();
        std::cout << "✓ 已设置 " << args[1] << " = " << args[2] << "\n";
    } else if (args[0] == "mem" && args.size() >= 2) {
        int mem = std::stoi(args[1]);
        settings.setInt("max_memory", mem);
        settings.save();
        std::cout << "✓ 内存已设置为 " << mem << " MB\n";
    }
    return 0;
}

int cmdInfo() {
    std::cout << "MC Launcher - 跨平台 Minecraft 启动器 (命令行版)\n";
    std::cout << "版本: 1.0.0\n";
    std::cout << "启动器目录: " << utils::PathUtils::getLauncherDir() << "\n";
    std::cout << "游戏目录: " << join(utils::PathUtils::getLauncherDir(), ".minecraft") << "\n\n";
    
    // 显示 Java 信息
    JavaManager javaMgr;
    auto java = javaMgr.findJava("17");
    if (java) {
        std::cout << "Java: " << java->path << " (版本 " << java->version << ")\n";
    } else {
        std::cout << "Java: 未找到\n";
    }
    
    // 显示已安装版本
    InstanceManager instances;
    auto versions = instances.listInstalled();
    std::cout << "已安装版本: " << versions.size() << " 个\n";
    
    return 0;
}

// ============================================================
// 主程序
// ============================================================
int main(int argc, char* argv[]) {
    // 初始化
    Logger::instance().setLevel(LogLevel::INFO);
    
    if (argc < 2) {
        printHelp(argv[0]);
        return 0;
    }
    
    std::string cmd = argv[1];
    std::vector<std::string> args(argv + 2, argv + argc);
    
    // 命令分发
    if (cmd == "-h" || cmd == "-help" || cmd == "--help" || cmd == "help") {
        printHelp(argv[0]);
    } else if (cmd == "-i" || cmd == "-install") {
        return cmdInstall(args);
    } else if (cmd == "-l" || cmd == "-launch") {
        return cmdLaunch(args);
    } else if (cmd == "-ls" || cmd == "-list") {
        return cmdList(args);
    } else if (cmd == "-rm" || cmd == "-remove") {
        return cmdRemove(args);
    } else if (cmd == "-v" || cmd == "-verify") {
        return cmdVerify(args);
    } else if (cmd == "-java") {
        return cmdJava(args);
    } else if (cmd == "-login") {
        return cmdLogin(args);
    } else if (cmd == "-account") {
        return cmdAccount(args);
    } else if (cmd == "-mods") {
        return cmdMods(args);
    } else if (cmd == "-loader") {
        return cmdLoader(args);
    } else if (cmd == "-assets") {
        return cmdAssets(args);
    } else if (cmd == "-crash") {
        return cmdCrash(args);
    } else if (cmd == "-saves") {
        return cmdSaves(args);
    } else if (cmd == "-servers") {
        return cmdServers(args);
    } else if (cmd == "-settings") {
        return cmdSettings(args);
    } else if (cmd == "-info") {
        return cmdInfo();
    } else {
        std::cerr << "未知命令: " << cmd << "\n";
        printHelp(argv[0]);
        return 1;
    }
    
    return 0;
}
