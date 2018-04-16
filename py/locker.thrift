namespace py lockerRPC

include "shared.thrift"

service LockerRPC {
    bool isLocked(1: string Key, 2: string Type),
    i64 requestLock(1: string Key, 2: string Type),
    void releaseLock(2: string Key),
}