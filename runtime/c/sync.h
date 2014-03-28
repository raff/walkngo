#include <mutex>
#include <condition_variable>

using namespace std;

class Mutex {
    friend class Cond;
private:
    mutex m;
public:
    void Lock() {
        m.lock();
    }

    void Unlock() {
        m.unlock();
    }
};

class Cond {
private:
    condition_variable cv;
public:
    Mutex *L;

    Cond(Mutex &m) : L(&m) {}

    void Wait() {
        unique_lock<mutex> lk(L->m);
        cv.wait(lk);
    }

    void Signal() {
        cv.notify_one();
    }

    void Broadcast() {
        cv.notify_all();
    }
};
