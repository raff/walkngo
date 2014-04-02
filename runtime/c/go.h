#ifndef _GO_RUNTIME_H
#define _GO_RUNTIME_H 1

#include <iostream>
#include <map>
#include <string>
#include <tuple>
#include <queue>
#include <thread>
#include <mutex>
#include <condition_variable>

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

    std::string Error() {
        return s;
    }
};

inline void panic(std::string &arg) {
    std::cerr << "panic: " << arg << std::endl;
    char *paniker = 0;
    *paniker = 0;
}

inline void Goroutine(std::function<void()> const& fun) {
    std::thread t(fun);
    t.detach();
}

class Deferred {
private:
    std::function<void()> const& deferred_call;

public:
    Deferred(std::function<void()> const& fun) : deferred_call(fun) {
    }

    ~Deferred() {
        deferred_call();
    }
};

template<class T> class Chan {
private:
    std::queue<T> buffer;
    int size;
    std::mutex m;
    std::condition_variable send_cond;
    std::condition_variable recv_cond;

public:
    Chan(int n=1) : size(n) {
    }

    void Send(T value) {
        std::unique_lock<std::mutex> lk(m);

        while (buffer.size() >= size) {
            send_cond.wait(lk);
        }

        buffer.push(value);
        recv_cond.notify_one();
    }

    T Receive() {
        std::unique_lock<std::mutex> lk(m);

        while (buffer.empty()) {
            recv_cond.wait(lk);
        }

        T ret = buffer.front();
        buffer.pop();
        send_cond.notify_one();
        return ret;
    }
};
#endif
