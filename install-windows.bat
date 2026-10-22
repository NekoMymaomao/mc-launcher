@echo off
chcp 65001 >nul 2>&1
setlocal EnableDelayedExpansion

echo ========================================
echo   MC Launcher WebUI - Windows 安装程序
echo ========================================
echo.

REM 检查 Git
where git >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ 错误: 未找到 git，请先安装 Git
    echo    下载地址: https://git-scm.com/download/win
    pause
    exit /b 1
)
echo ✅ Git 已安装

REM 设置安装目录
set "REPO_URL=https://github.com/NekoMymaomao/mc-launcher-webui.git"
set "TARGET_DIR=%USERPROFILE%\mc-launcher"

REM 检查是否已安装
if exist "%TARGET_DIR%" (
    echo 📁 检测到已有安装目录: %TARGET_DIR%
    set /p OVERWRITE="是否覆盖现有安装? (Y/N): "
    if /i not "!OVERWRITE!"=="Y" (
        echo 取消安装
        pause
        exit /b 0
    )
    rmdir /s /q "%TARGET_DIR%"
)

echo.
echo 📥 正在克隆仓库...
git clone --depth 1 "%REPO_URL%" "%TARGET_DIR%"

if not exist "%TARGET_DIR%" (
    echo ❌ 克隆失败
    pause
    exit /b 1
)

echo ✅ 仓库已克隆到: %TARGET_DIR%

REM 检查架构并下载对应二进制
cd /d "%TARGET_DIR%"

echo.
echo 🔍 检测系统架构...
for /f "tokens=*" %%a in ('wmic os get osarchitecture ^| findstr /i "64"') do set "ARCH=%%a"

if "%ARCH%"=="64-bit" (
    set "BIN=mc-launcher-windows-amd64.exe"
    echo ✅ 检测到 64 位系统
) else (
    echo ⚠️  未检测到 64 位系统，将使用默认二进制
    set "BIN=mc-launcher-windows-amd64.exe"
)

REM 检查核心文件
if not exist "Core\mc-launcher\%BIN%" (
    echo ⚠️  未找到预编译二进制: Core\mc-launcher\%BIN%
    echo    请手动下载或从 GitHub Releases 获取
)

REM 创建启动脚本
echo @echo off > launcher.bat
echo chcp 65001 >nul 2>&1 >> launcher.bat
echo title MC Launcher WebUI >> launcher.bat
echo echo ======================================== >> launcher.bat
echo echo   MC Launcher WebUI >> launcher.bat
echo echo ======================================== >> launcher.bat
echo echo. >> launcher.bat
echo echo 🚀 启动 MC Launcher 服务器... >> launcher.bat
echo echo    地址: http://localhost:9876 >> launcher.bat
echo echo    按 Ctrl+C 停止 >> launcher.bat
echo echo. >> launcher.bat
echo cd /d "%%~dp0" >> launcher.bat
echo start "" "Core\mc-launcher\%BIN%" -addr 0.0.0.0:9876 >> launcher.bat
echo timeout /t 2 >nul >> launcher.bat
echo start "" "http://localhost:9876" >> launcher.bat
echo echo. >> launcher.bat
echo echo ✅ 服务器已启动，浏览器应该已打开 >> launcher.bat
echo echo    如果浏览器未打开，请手动访问: http://localhost:9876 >> launcher.bat
echo echo. >> launcher.bat
echo pause >> launcher.bat

echo ✅ 启动脚本已创建: launcher.bat

REM 创建 README
(
echo # MC Launcher WebUI - Windows 安装说明
echo.
echo ## 安装完成
echo.
echo ✅ 已成功安装到: %TARGET_DIR%
echo.
echo ## 使用方法
echo.
echo ### 启动服务器
echo ```
echo cd %TARGET_DIR%
echo .\launcher.bat
echo ```
echo.
echo ### 访问 Web UI
echo 启动服务器后，在浏览器打开:
echo - http://localhost:9876
echo - http://127.0.0.1:9876
echo.
echo ### 更新版本
echo ```
echo cd %TARGET_DIR%
echo git pull
echo ```
echo.
echo ## 文件说明
echo - launcher.bat - 一键启动脚本
echo - Core/mc-launcher/ - 核心二进制文件
echo - frontend/ - Web UI 前端文件
echo - android/ - Android 项目
echo - .github/workflows/ - CI/CD 配置
echo.
echo ## 卸载
echo 删除目录: %TARGET_DIR%
) > README-INSTALL.md

echo.
echo ========================================
echo   ✅ 安装完成!
echo ========================================
echo.
echo 📁 安装位置: %TARGET_DIR%
echo 🚀 启动命令: cd %TARGET_DIR% ^&^& .\launcher.bat
echo 🌐 访问地址: http://localhost:9876
echo.
echo ========================================
pause
