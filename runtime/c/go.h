#ifndef _GO_RUNTIME_H
#define _GO_RUNTIME_H 1

#include <iostream>
#include <map>
#include <string>
#include <tuple>

// for string, map, etc.
using namespace std;

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

typedef string error;

inline void panic(string &arg) {
    cerr << "panic: " << arg << endl;
    char *paniker = 0;
    *paniker = 0;
}

#endif
