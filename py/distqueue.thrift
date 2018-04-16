include "shared.thrift"

service DistQueueRPC {

    shared.Response GetCompletedItem(1: i64 QueueInst, 2: shared.Command cmd),
    shared.Command GetItem(1: i64 QueueInst),
    void MarkComplete(1: i64 QueueInst, 2: shared.Command cmd, 3: shared.Response res),
    shared.Response PutItem(1: i64 QueueInst, 2: shared.Command cmd),

}