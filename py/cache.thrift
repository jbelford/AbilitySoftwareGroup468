namespace py cacheRPC


include "shared.thrift"

service Cache {
    shared.QuoteData GetQuote(1:string symbol, 2: string userId, 3: i64 tid),
}