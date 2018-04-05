include "shared.thrift"

service DistQueueRPC {
    shared.Command GetItem(1: i64 QueueInst),
    void MarkComplete(1: i64 QueueInst, 2: shared.Command cmd),
    void PutItem(1: i64 QueueInst, 2: shared.Command cmd),

}