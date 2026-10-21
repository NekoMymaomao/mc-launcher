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
# 依赖配置
# ============================================================

# nlohmann_json: header-only
find_package(nlohmann_json QUIET)
if(NOT nlohmann_json_FOUND)
    message(STATUS "nlohmann_json not found, downloading from GitHub...")
    include(FetchContent)
    FetchContent_Declare(
        nlohmann_json
        GIT_REPOSITORY https://github.com/nlohmann/json.git
        GIT_TAG v3.11.3
    )
    FetchContent_MakeAvailable(nlohmann_json)
endif()

# curl
find_library(CURL_LIBRARY NAMES curl libcurl)
find_path(CURL_INCLUDE_DIR NAMES curl/curl.h)
if(CURL_LIBRARY AND CURL_INCLUDE_DIR)
    add_library(CURL::libcurl UNKNOWN IMPORTED)
    set_target_properties(CURL::libcurl PROPERTIES
        IMPORTED_LOCATION "${CURL_LIBRARY}"
        INTERFACE_INCLUDE_DIRECTORIES "${CURL_INCLUDE_DIR}"
    )
    message(STATUS "Using system libcurl")
else()
    message(STATUS "libcurl not found, downloading from GitHub...")
    include(FetchContent)
    FetchContent_Declare(
        curl
        GIT_REPOSITORY https://github.com/curl/curl.git
        GIT_TAG curl-8_10_1
    )
    set(BUILD_TESTING OFF CACHE BOOL "" FORCE)
    set(BUILD_CURL_EXE OFF CACHE BOOL "" FORCE)
    set(CURL_STATICLIB ON CACHE BOOL "" FORCE)
    set(ENABLE_CURL_MANUAL OFF CACHE BOOL "" FORCE)
    set(ENABLE_IPV6 ON CACHE BOOL "" FORCE)
    FetchContent_MakeAvailable(curl)
endif()

# miniz (zip handling)
find_library(MINIZ_LIBRARY NAMES miniz)
find_path(MINIZ_INCLUDE_DIR NAMES miniz.h)
if(MINIZ_LIBRARY AND MINIZ_INCLUDE_DIR)
    message(STATUS "Using system miniz")
else()
    message(STATUS "miniz not found, downloading from GitHub...")
    include(FetchContent)
    FetchContent_Declare(
        miniz
        GIT_REPOSITORY https://github.com/richgel999/miniz.git
        GIT_TAG 3.0.2
    )
    FetchContent_MakeAvailable(miniz)
endif()

# zlib
find_package(ZLIB QUIET)
if(NOT ZLIB_FOUND)
    message(STATUS "zlib not found, downloading...")
    include(FetchContent)
    FetchContent_Declare(
        zlib
        GIT_REPOSITORY https://github.com/madler/zlib.git
        GIT_TAG v1.3.1
    )
    FetchContent_MakeAvailable(zlib)
endif()

# ============================================================
# 源码收集
# ============================================================
file(GLOB_RECURSE SOURCES "src/*.cpp")

add_executable(mc_launcher ${SOURCES})

target_include_directories(mc_launcher PRIVATE
    ${PROJECT_SOURCE_DIR}/include
    ${nlohmann_json_SOURCE_DIR}
    ${CURL_INCLUDE_DIR}
    ${MINIZ_INCLUDE_DIR}
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

# 安装启动器目录模板
install(DIRECTORY config/launcher_template/ DESTINATION share/mc_launcher)

# ============================================================
# 测试
# ============================================================
enable_testing()
add_executable(test_mc_launcher tests/test_main.cpp)
target_link_libraries(test_mc_launcher PRIVATE mc_launcher)
add_test(NAME AllTests COMMAND test_mc_launcher)
