cmake_minimum_required(VERSION 3.16)
project(mc_launcher VERSION 1.0.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# 平台定义
if(WIN32)
    add_compile_definitions(MC_WINDOWS)
elseif(UNIX AND NOT APPLE)
    add_compile_definitions(MC_LINUX)
endif()

# 查找 curl
find_package(CURL REQUIRED)

# 源码
file(GLOB_RECURSE SOURCES "src/*.cpp")

add_executable(mc_launcher ${SOURCES})

target_include_directories(mc_launcher PRIVATE ${PROJECT_SOURCE_DIR}/include)
target_link_libraries(mc_launcher PRIVATE CURL::libcurl)

if(WIN32)
    target_link_libraries(mc_launcher PRIVATE ws2_32 crypt32)
endif()

# 测试
enable_testing()
add_executable(test_mc_launcher tests/test_main.cpp)
target_link_libraries(test_mc_launcher PRIVATE mc_launcher)
add_test(NAME AllTests COMMAND test_mc_launcher)

# 安装
install(TARGETS mc_launcher RUNTIME DESTINATION bin)
