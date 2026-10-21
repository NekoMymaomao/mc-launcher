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

# nlohmann_json
find_package(nlohmann_json REQUIRED)

# curl
find_package(CURL REQUIRED)

# zlib
find_package(ZLIB REQUIRED)

# ============================================================
# 源码收集
# ============================================================
file(GLOB_RECURSE SOURCES "src/*.cpp")

add_executable(mc_launcher ${SOURCES})

target_include_directories(mc_launcher PRIVATE
    ${PROJECT_SOURCE_DIR}/include
)

# 链接库
target_link_libraries(mc_launcher PRIVATE
    nlohmann_json::nlohmann_json
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

# ============================================================
# 测试
# ============================================================
enable_testing()
add_executable(test_mc_launcher tests/test_main.cpp)
target_link_libraries(test_mc_launcher PRIVATE mc_launcher)
add_test(NAME AllTests COMMAND test_mc_launcher)
