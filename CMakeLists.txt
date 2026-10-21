cmake_minimum_required(VERSION 3.20)
project(mc_launcher VERSION 1.0.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# 平台检测
if(WIN32)
    add_compile_definitions(MC_WINDOWS)
    message(STATUS "Building for Windows")
elseif(UNIX AND NOT APPLE)
    add_compile_definitions(MC_LINUX)
    message(STATUS "Building for Linux")
else()
    message(WARNING "Unsupported platform")
endif()

# ============================================================
# 依赖配置
# ============================================================

# nlohmann_json: header-only
set(nlohmann_json_DIR "" CACHE PATH "nlohmann_json install path")
find_package(nlohmann_json REQUIRED)

# curl: 优先使用静态库
set(CURL_STATICLIB ON CACHE BOOL "Link curl statically")
find_package(CURL REQUIRED)

# Boost (可选，仅用于filesystem)
option(USE_BOOST "Use Boost filesystem" ON)
if(USE_BOOST)
    find_package(Boost REQUIRED COMPONENTS filesystem)
endif()

# ============================================================
# 源码
# ============================================================
file(GLOB_RECURSE SOURCES
    "src/*.cpp"
)

add_executable(mc_launcher ${SOURCES})

target_include_directories(mc_launcher PRIVATE
    ${PROJECT_SOURCE_DIR}/include
    ${PROJECT_SOURCE_DIR}/vendor
)

target_link_libraries(mc_launcher PRIVATE
    nlohmann_json::nlohmann_json
    CURL::libcurl
)

if(USE_BOOST)
    target_link_libraries(mc_launcher PRIVATE
        Boost::filesystem
        Boost::system
    )
endif()

# Windows 链接选项
if(WIN32)
    target_link_options(mc_launcher PRIVATE
        /SUBSYSTEM:CONSOLE
    )
endif()

# ============================================================
# 测试
# ============================================================
enable_testing()
add_executable(test_mc_launcher tests/test_main.cpp)
target_link_libraries(test_mc_launcher PRIVATE
    mc_launcher
    nlohmann_json::nlohmann_json
    CURL::libcurl
)
add_test(NAME McLauncherTests COMMAND test_mc_launcher)

# ============================================================
# 安装
# ============================================================
install(TARGETS mc_launcher RUNTIME DESTINATION bin)
