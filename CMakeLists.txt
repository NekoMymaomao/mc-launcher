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
# 依赖配置 - 优先使用系统库
# ============================================================

# nlohmann_json
find_package(nlohmann_json QUIET)
if(NOT nlohmann_json_FOUND)
    find_path(NLOHMANN_JSON_INCLUDE_DIR NAMES nlohmann/json.hpp)
    if(NOT NLOHMANN_JSON_INCLUDE_DIR)
        message(FATAL_ERROR "nlohmann_json not found. Install: apt-get install nlohmann-json3-dev")
    endif()
    message(STATUS "Found nlohmann_json at: ${NLOHMANN_JSON_INCLUDE_DIR}")
endif()

# curl
find_package(CURL QUIET)
if(NOT CURL_FOUND)
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
        message(FATAL_ERROR "curl not found. Install: apt-get install libcurl4-openssl-dev")
    endif()
endif()

# zlib
find_package(ZLIB QUIET)
if(NOT ZLIB_FOUND)
    find_library(ZLIB_LIBRARY NAMES z)
    find_path(ZLIB_INCLUDE_DIR NAMES zlib.h)
    if(ZLIB_LIBRARY AND ZLIB_INCLUDE_DIR)
        add_library(ZLIB::zlib UNKNOWN IMPORTED)
        set_target_properties(ZLIB::zlib PROPERTIES
            IMPORTED_LOCATION "${ZLIB_LIBRARY}"
            INTERFACE_INCLUDE_DIRECTORIES "${ZLIB_INCLUDE_DIR}"
        )
        message(STATUS "Using system zlib")
    else()
        message(FATAL_ERROR "zlib not found. Install: apt-get install zlib1g-dev")
    endif()
endif()

# ============================================================
# 源码收集
# ============================================================
file(GLOB_RECURSE SOURCES "src/*.cpp")

add_executable(mc_launcher ${SOURCES})

target_include_directories(mc_launcher PRIVATE
    ${PROJECT_SOURCE_DIR}/include
    ${NLOHMANN_JSON_INCLUDE_DIR}
    ${CURL_INCLUDE_DIR}
    ${ZLIB_INCLUDE_DIR}
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
install(DIRECTORY config/launcher_template/ DESTINATION share/mc_launcher)

# ============================================================
# 测试
# ============================================================
enable_testing()
add_executable(test_mc_launcher tests/test_main.cpp)
target_link_libraries(test_mc_launcher PRIVATE mc_launcher)
add_test(NAME AllTests COMMAND test_mc_launcher)
