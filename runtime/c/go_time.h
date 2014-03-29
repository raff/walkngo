#ifndef _GO_RUNTIME_TIME_H
#define _GO_RUNTIME_TIME_H 1

#include <thread>
#include <chrono>

namespace go_time {

const std::chrono::duration<double> Nanosecond = std::chrono::nanoseconds(1);
const std::chrono::duration<double> Microsecond = std::chrono::microseconds(1);
const std::chrono::duration<double, std::milli> Millisecond = std::chrono::milliseconds(1);
const std::chrono::duration<double> Second = std::chrono::seconds(1);
const std::chrono::duration<double> Minute = std::chrono::minutes(1);
const std::chrono::duration<double> Hour = std::chrono::hours(1);

typedef std::chrono::duration<double> Duration;

inline void Sleep(Duration d) {
    std::this_thread::sleep_for(d);
}

}

#endif
