#include <iostream>
#include <map>
#include <string>
#include <tuple>

// for string, map, etc.
using namespace std;

typedef unsigned char byte;

typedef string error;

inline void panic(string &arg) {
    cerr << "panic: " << arg << endl;
    char *paniker = 0;
    *paniker = 0;
}
