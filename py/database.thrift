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

   void PushPendingTxn(1: shared.PendingTxn pending),

}


struct DBResponse {
    1: optional shared.error error
    2: optional shared.User user
}