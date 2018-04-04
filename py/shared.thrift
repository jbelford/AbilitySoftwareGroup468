namespace py shared


struct Response {
    1:bool Success
    2:optional string Message
    3:optional string Stock
    4:optional i64 Quote
    5:optional i64 ReqAmount
	6:optional i64 RealAmount
	7:optional i32 Shares
	8:optional i64 Expiration
	9:optional i64 Paid
	10:optional i64 Received
	11:optional i32 SharesAfford
	12:optional i64 AffordAmount
	13:optional UserInfo Status
	14:optional list<Transaction>Transactions
	15:optional list<Trigger>Triggers
	16:optional list<binary> File
}

struct Command {
    1:i32 C_type
	2:i64 TransactionID
	3:string UserId
	4:optional i64 Amount
	5:optional string StockSymbol
	6:optional string FileName
	7:optional i64 Timestamp
}

struct UserInfo {
    1: i64 Balance
    2: i64 Reserved
    3: STOCK Stock
}

struct User {
    1: string User
    2: i64 Balance
    3: i64 Reserved
    4: STOCK stock
}

struct STOCK {
    1: i64 Real
    2: i64 Reserved
}

struct Trigger {
    1: string UserId
    2: string Stock
    3: i64 TransactionID
    4: string Type
    5: i32 Shares
    6: i64 Amount
    7: i64 When
}

struct Transaction {
    1: string Type
    2: bool Triggered
    3: string Stock
    4: i64 Amount
    5: i32 Shares
    6: i64 TimeStamp
}

struct PendingTxn {
    1: string UserId
    2: string Type
    3: string Stock
    4: i64 Reserved
    5: i64 Price
    6: i32 Shares
    7: i64 Expiry
    8: error error
}

struct QuoteData {
    1: i64 Quote
    2: string Symbol
    3: string UserId
    4: i64 Timestamp
    5: string Cryptokey
    6: error error
}


typedef string error