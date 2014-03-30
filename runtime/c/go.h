#ifndef _GO_RUNTIME_H
#define _GO_RUNTIME_H 1

#include <iostream>
#include <map>
#include <string>
#include <tuple>
#include <thread>

typedef unsigned char      uint8;
typedef unsigned short int uint16;
typedef unsigned int       uint32;
typedef unsigned long int  uint64;

typedef signed char      int8;
typedef signed short int int16;
typedef signed int       int32;
typedef signed long int  int64;

typedef float  float32;
typedef double float64;

typedef uint8 byte;
typedef int32 rune;

class error {
private:
    std::string s;
public:
    error(std::string message) {
        s = message;
    }

    string Error() {
        return s;
    }
};

inline void panic(std::string &arg) {
    cerr << "panic: " << arg << endl;
    char *paniker = 0;
    *paniker = 0;
}

inline void GoCall(function<void()> const& fun) {
    thread t(fun);
    t.detach();
}

#endif
