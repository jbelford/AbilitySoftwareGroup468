namespace py transactionRPC


include "shared.thrift"

service Transaction {
    shared.Response ADD(1:shared.Command cmd),
    shared.Response QUOTE(1:shared.Command cmd),
    shared.Response BUY(1:shared.Command cmd),
    shared.Response COMMIT_BUY(1:shared.Command cmd),
    shared.Response CANCEL_BUY(1:shared.Command cmd),
    shared.Response SELL(1:shared.Command cmd),
    shared.Response COMMIT_SELL(1:shared.Command cmd),
    shared.Response CANCEL_SELL(1:shared.Command cmd),
    shared.Response SET_BUY_AMOUNT(1:shared.Command cmd),
    shared.Response CANCEL_SET_BUY(1:shared.Command cmd),
    shared.Response SET_BUY_TRIGGER(1:shared.Command cmd),
    shared.Response SET_SELL_AMOUNT(1:shared.Command cmd),
    shared.Response SET_SELL_TRIGGER(1:shared.Command cmd),
    shared.Response CANCEL_SET_SELL(1:shared.Command cmd),
    shared.Response DUMPLOG(1:shared.Command cmd),
    shared.Response DISPLAY_SUMMARY(1:shared.Command cmd),

}

service TriggerManRpc {
    shared.PendingTxn ProcessTrigger(1: shared.Trigger t),

}