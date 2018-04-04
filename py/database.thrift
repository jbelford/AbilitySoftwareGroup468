include "shared.thrift"

namespace py databaseRPC


service Database {
    DBResponse AddUserMoney(1: string userId, 2: i64 amount),
    DBResponse UnreserveMoney(1: string userId, 2: i64 amount),
    DBResponse ReserveMoney(1: string userId, 2: i64 amount),
    i32 GetReserveMoney(1: string userId),
    DBResponse UnreserveShares(1: string userId, 2: string stock, 3: i32 shares),
    DBResponse ReservedShares(1: string userId, 2: string stock, 3: i32 shares),
    i32 GetReservedShares(1: string userId, 2: string stock),
    DBResponse GetUser(1: string userId),
    DBResponse BulkTransactions(1: list<shared.PendingTxn> txns, 2: bool wasCached),
    DBResponse ProcessTxn(1: shared.PendingTxn txn, 2: bool wasCached)

    shared.Trigger PopPendingTxn(1: string userId, 2: string txnType),
    void PushPendingTxn(1: shared.PendingTxn pending),
    DBResponse AddNewTrigger(1: shared.Trigger trigger),
    shared.Trigger CancelTrigger(1: string userId, 2: string stock, 3: string trigger_type),
    shared.Trigger GetTrigger(1: string userId, 2: string stock, 3: string trigger_type),
}


struct DBResponse {
    1: optional shared.error error
    2: optional shared.User user
}