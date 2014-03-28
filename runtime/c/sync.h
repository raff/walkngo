#ifndef _GO_RUNTIME_SYNC_H
#define _GO_RUNTIME_SYNC_H

#include <mutex>
#include <condition_variable>

namespace sync {
    class Mutex : private std::mutex {
        friend class Cond;
    public:
        void Lock() {
            std::mutex::lock();
        }

        void Unlock() {
            std::mutex::unlock();
        }
    };

    class Cond {
    private:
        std::condition_variable_any cv;
    public:
        Mutex *L;
        Cond(Mutex &m) : L(&m) {}

        void Wait() {
            cv.wait(static_cast<std::mutex&>(*L));
        }

        void Signal() {
            cv.notify_one();
        }

        void Broadcast() {
            cv.notify_all();
        }
    };
}

#endif
