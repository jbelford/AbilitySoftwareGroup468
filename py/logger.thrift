namespace py log

include "shared.thrift"

service LoggerRpc {
    shared.error UserCommand(1: shared.Command cmd),
    shared.error QuoteServer(1: shared.QuoteData quote, 2: i64 tid),
    shared.error AccountTransaction(1:string userid, 2: i64 funds, 3: string action, 4: i64 tid),
    shared.error SystemEvent(1: shared.Command cmd),
    shared.error ErrorEvent(1: shared.Command cmd, 2: string e),
    shared.error DebugEvent(1: shared.Command cmd, 2: string debug),
    list<binary> DumpLogUser(1: string userid),
    list<binary> DumpLog(),
}