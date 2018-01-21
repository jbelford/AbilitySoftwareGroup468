package transaction

import (
	"log"

    "net"
    //"fmt"
    "bufio"
    //"strings" // only needed below for sample processing
)

type TransactionServer struct{}

func handle_add(userid string, amount float32) {
    log.Println("handle_add")
}

func handle_quote(userid string, amount float32) {
    log.Println("handle_quote")
}

func handle_buy(userid string, stocksymbol string, amount float32) {
    log.Println("handle_buy")
}

func handle_commit_buy(userid string) {
    log.Println("handle_commit_buy")
}

func handle_cancel_buy(userid string) {
    log.Println("handle_cancel_buy")
}

func handle_sell(userid string, stocksymbol string, amount float32) {
    log.Println("handle_sell")
}

func handle_commit_sell(userid string) {
    log.Println("handle_commit_sell")
}

func handle_cancel_sell(userid string) {
    log.Println("handle_cancel_sell")
}

func handle_set_buy_amount(userid string, stocksymbol string, amount float32) {
    log.Println("handle_set_buy_amount")
}

func handle_cancel_set_buy(userid string, stocksymbol string) {
    log.Println("handle_cancel_set_buy")
}

func handle_set_buy_trigger(userid string, stocksymbol string, amount float32) {
    log.Println("handle_set_buy_trigger")
}

func handle_set_sell_amount(userid string, stocksymbol string, amount float32) {
    log.Println("handle_set_sell_amount")
}

func handle_set_sell_trigger(userid string, stocksymbol string, amount float32) {
    log.Println("handle_set_sell_trigger")
}

func handle_cancel_set_sell(userid string, stocksymbol string) {
    log.Println("handle_cancel_set_sell")
}

func handle_admin_dumplog(userid string, filename string) {
    log.Println("handle_admin_dumplog")
}

func handle_dumplog(filename string) {
    log.Println("handle_dumplog")
}

func handle_display_summary(userid string) {
    log.Println("handle_display_summary")
}

func (ts *TransactionServer) Start() {
    conn, _ := net.Dial("tcp", "127.0.0.1:8081")
    for {
        message, _ := bufio.NewReader(conn).ReadString('\n')
        log.Println("Received: ", string(message))

        //so.On("add", handle_add)
        //so.On("quote", handle_quote)
		//so.On("buy", handle_buy)
		//so.On("commmit_buy", handle_commit_buy)
		//so.On("cancel_buy", handle_cancel_buy)
		//so.On("sell", handle_sell)
		//so.On("commmit_sell", handle_commit_sell)
		//so.On("cancel_sell", handle_cancel_sell)
		//so.On("set_buy_amount", handle_set_buy_amount)
		//so.On("cancel_set_buy", handle_cancel_set_buy)
		//so.On("set_buy_trigger", handle_set_buy_trigger)
		//so.On("set_sell_amount", handle_set_sell_amount)
        //so.On("set_sell_trigger", handle_set_sell_trigger)
        //so.On("cancel_set_sell", handle_cancel_set_sell)
        //so.On("dumplog", handle_dumplog)
        //so.On("admin_dumplog", handle_admin_dumplog)
        //so.On("display_summary", handle_display_summary)
	}
}

func main() {

}
