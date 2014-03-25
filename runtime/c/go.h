#include <iostream>
#include <map>
#include <string>

// for string, map, etc.
using namespace std;

typedef unsigned char byte;

inline void panic(string &arg) {
    cerr << "panic: " << arg << endl;
    char *paniker = 0;
    *paniker = 0;
}
