#include "core/utils.hpp"
#include <iostream>

using namespace mc::utils;

void testUtils() {
    assert(StringUtils::trim("  hello  ") == "hello");
    assert(StringUtils::toLower("HELLO") == "hello");
    assert(StringUtils::startsWith("hello world", "hello"));

    auto parts = StringUtils::split("a:b:c", ':');
    assert(parts.size() == 3);

    std::string uuid = Uuid::generateV4();
    assert(uuid.size() == 36);
    assert(uuid[8] == '-');

    std::string hash = Sha1::compute("hello");
    assert(hash == "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d");

    std::cout << "[PASS] testUtils\n";
}

void testJson() {
    Json::Value v = Json::parse("{\"name\":\"test\",\"value\":42,\"flag\":true}");
    assert(v["name"].getString() == "test");
    assert(v["value"].getNumber() == 42);
    assert(v["flag"].getBool() == true);

    Json::Value arr = Json::parse("[1,2,3]");
    assert(arr.size() == 3);
    assert(arr[0].getNumber() == 1);

    std::cout << "[PASS] testJson\n";
}

int main() {
    std::cout << "Running tests...\n";
    testUtils();
    testJson();
    std::cout << "All tests passed!\n";
    return 0;
}
