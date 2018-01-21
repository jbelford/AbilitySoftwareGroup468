package transaction

import (
	"log"
	"net/http"

	"github.com/googollee/go-socket.io"
)

type TransactionServer struct{}

func handle_add(userid int, amount float32) {
    log.Println("handle_add")
}

func handle_quote(userid int, amount float32) {
    log.Println("handle_quote")
}

func handle_buy(userid int, stocksymbol string, amount float32) {
    log.Println("handle_buy")
}

func handle_commit_buy(userid int) {
    log.Println("handle_commit_buy")
}

func handle_cancel_buy(userid int) {
    log.Println("handle_cancel_buy")
}

func handle_sell(userid int, stocksymbol string, amount float32) {
    log.Println("handle_sell")
}

func handle_commit_sell(userid int) {
    log.Println("handle_commit_sell")
}

func handle_cancel_sell(userid int) {
    log.Println("handle_cancel_sell")
}

func handle_set_buy_amount(userid int, stocksymbol string, amount float32) {
    log.Println("handle_set_buy_amount")
}

func handle_cancel_set_buy(userid int, stocksymbol string) {
    log.Println("handle_cancel_set_buy")
}

func handle_set_buy_trigger(userid int, stocksymbol string, amount float32) {
    log.Println("handle_set_buy_trigger")
}

func handle_set_sell_amount(userid int, stocksymbol string, amount float32) {
    log.Println("handle_set_sell_amount")
}

func handle_set_sell_trigger(userid int, stocksymbol string, amount float32) {
    log.Println("handle_set_sell_trigger")
}

func handle_cancel_set_sell(userid int, stocksymbol string) {
    log.Println("handle_cancel_set_sell")
}

func handle_admin_dumplog(userid int, filename string) {
    log.Println("handle_admin_dumplog")
}

func handle_dumplog(filename string) {
    log.Println("handle_dumplog")
}

func handle_display_summary(userid int) {
    log.Println("handle_display_summary")
}

func (ts *TransactionServer) Start() {
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}
	server.On("connection", func(so socketio.Socket) {
		log.Println("on connection")
        so.Join("add")
        so.Join("quote")
        so.Join("buy")
        so.Join("commmit_buy")
        so.Join("cancel_buy")
        so.Join("sell")
        so.Join("commmit_sell")
        so.Join("cancel_sell")
        so.Join("set_buy_amount")
        so.Join("cancel_set_buy")
        so.Join("set_buy_trigger")
        so.Join("set_sell_amount")
        so.Join("set_sell_trigger")
        so.Join("cancel_set_sell")
        so.Join("dumplog")
        so.Join("admin_dumplog")
        so.Join("display_summary")

        so.On("add", handle_add)
        so.On("quote", handle_quote)
		so.On("buy", handle_buy)
		so.On("commmit_buy", handle_commit_buy)
		so.On("cancel_buy", handle_cancel_buy)
		so.On("sell", handle_sell)
		so.On("commmit_sell", handle_commit_sell)
		so.On("cancel_sell", handle_cancel_sell)
		so.On("set_buy_amount", handle_set_buy_amount)
		so.On("cancel_set_buy", handle_cancel_set_buy)
		so.On("set_buy_trigger", handle_set_buy_trigger)
		so.On("set_sell_amount", handle_set_sell_amount)
        so.On("set_sell_trigger", handle_set_sell_trigger)
        so.On("cancel_set_sell", handle_cancel_set_sell)
        so.On("dumplog", handle_dumplog)
        so.On("admin_dumplog", handle_admin_dumplog)
        so.On("display_summary", handle_display_summary)
	})
	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})

	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("./asset")))
	log.Println("Serving at localhost:5000...")
	log.Fatal(http.ListenAndServe(":5000", nil))
}

func main() {

}
