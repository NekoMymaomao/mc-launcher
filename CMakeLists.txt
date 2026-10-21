cmake_minimum_required(VERSION 3.20)
project(mc_launcher VERSION 1.0.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# ============================================================
# 平台定义
# ============================================================
if(WIN32)
    add_compile_definitions(MC_WINDOWS)
    message(STATUS "Windows build")
elseif(UNIX AND NOT APPLE)
    add_compile_definitions(MC_LINUX)
    message(STATUS "Linux build")
elseif(APPLE)
    add_compile_definitions(MC_MACOS)
    message(STATUS "macOS build")
endif()

# ============================================================
# FetchContent - 所有依赖单体外置
# ============================================================
include(FetchContent)

# nlohmann_json (header-only)
FetchContent_Declare(
    nlohmann_json
    GIT_REPOSITORY https://github.com/nlohmann/json.git
    GIT_TAG v3.11.3
)
FetchContent_MakeAvailable(nlohmann_json)

# curl (static)
FetchContent_Declare(
    curl
    GIT_REPOSITORY https://github.com/curl/curl.git
    GIT_TAG curl-8_10_1
)
set(BUILD_TESTING OFF CACHE BOOL "" FORCE)
set(BUILD_CURL_EXE OFF CACHE BOOL "" FORCE)
set(CURL_STATICLIB ON CACHE BOOL "" FORCE)
set(ENABLE_CURL_MANUAL OFF CACHE BOOL "" FORCE)
FetchContent_MakeAvailable(curl)

# miniz (zip handling, header-only-ish)
FetchContent_Declare(
    miniz
    GIT_REPOSITORY https://github.com/richgel999/miniz.git
    GIT_TAG 3.0.2
)
FetchContent_MakeAvailable(miniz)

# zlib for curl
if(NOT TARGET zlib::zlib)
    find_package(ZLIB REQUIRED)
endif()

# ============================================================
# 源码收集
# ============================================================
file(GLOB_RECURSE SOURCES "src/*.cpp")
file(GLOB_RECURSE HEADERS "include/**/*.hpp")

add_executable(mc_launcher ${SOURCES} ${HEADERS})

target_include_directories(mc_launcher PRIVATE
    ${PROJECT_SOURCE_DIR}/include
    ${miniz_SOURCE_DIR}
    ${nlohmann_json_SOURCE_DIR}
)

# 链接库
target_link_libraries(mc_launcher PRIVATE
    CURL::libcurl
    ZLIB::zlib
    ${CMAKE_DL_LIBS}
)

# Windows 特定
if(WIN32)
    target_link_libraries(mc_launcher PRIVATE ws2_32 crypt32)
    set_target_properties(mc_launcher PROPERTIES WIN32_EXECUTABLE OFF)
endif()

# macOS 特定
if(APPLE)
    target_link_libraries(mc_launcher PRIVATE "-framework CoreFoundation" "-framework Security")
endif()

# ============================================================
# 安装配置
# ============================================================
install(TARGETS mc_launcher RUNTIME DESTINATION bin)

# 安装启动器目录结构说明（在 README 中描述）

# ============================================================
# 测试
# ============================================================
enable_testing()
add_executable(test_mc_launcher tests/test_main.cpp)
target_link_libraries(test_mc_launcher PRIVATE mc_launcher)
add_test(NAME AllTests COMMAND test_mc_launcher)
