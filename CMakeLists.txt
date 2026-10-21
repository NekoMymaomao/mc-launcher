cmake_minimum_required(VERSION 3.16)
project(mc_launcher VERSION 1.0.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# 平台定义
if(WIN32)
    add_compile_definitions(MC_WINDOWS)
    # 源码含中文，强制 MSVC 以 UTF-8 解析
    add_compile_options(/utf-8)
    add_compile_definitions(_CRT_SECURE_NO_WARNINGS)
elseif(UNIX AND NOT APPLE)
    add_compile_definitions(MC_LINUX)
endif()

# 查找 curl
find_package(CURL REQUIRED)

# 核心库
file(GLOB_RECURSE CORE_SOURCES "src/core/*.cpp")
add_library(mc_launcher_core ${CORE_SOURCES})
target_include_directories(mc_launcher_core PUBLIC ${PROJECT_SOURCE_DIR}/include)
target_link_libraries(mc_launcher_core PRIVATE CURL::libcurl)

if(WIN32)
    target_link_libraries(mc_launcher_core PRIVATE ws2_32 crypt32)
endif()

# 主程序
add_executable(mc_launcher src/main.cpp)
target_link_libraries(mc_launcher PRIVATE mc_launcher_core)

# 测试
enable_testing()
add_executable(test_mc_launcher tests/test_main.cpp)
target_link_libraries(test_mc_launcher PRIVATE mc_launcher_core)
add_test(NAME AllTests COMMAND test_mc_launcher)

# 安装
install(TARGETS mc_launcher RUNTIME DESTINATION bin)
