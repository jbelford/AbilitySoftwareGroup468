namespace py log

include "shared.thrift"

service LoggerRpc {
    oneway void UserCommand(1: shared.Command cmd),
    oneway void QuoteServer(1: shared.QuoteData quote, 2: i64 tid),
    oneway void AccountTransaction(1:string userid, 2: double funds, 3: string action, 4: i64 tid),
    oneway void SystemEvent(1: shared.Command cmd),
    oneway void ErrorEvent(1: shared.Command cmd, 2: string e),
    oneway void DebugEvent(1: shared.Command cmd, 2: string debug),
    string DumpLogUser(1: string userid),
    string DumpLog(),
}