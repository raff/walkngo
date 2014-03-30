#ifndef _GO_RUNTIME_FMT_H
#define _GO_RUNTIME_FMT_H 1

#include <iostream>

namespace fmt {

template<typename... T> void Print(T... args) {
    int dummy[sizeof...(T)] = { (std::cout << args, 0)... };
}

template<typename... T> void Println(T... args) {
    int dummy[sizeof...(T)] = { (std::cout << args << " ", 0)... };
    std::cout << std::endl;
}

}

#endif
