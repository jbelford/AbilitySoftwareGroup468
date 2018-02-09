package networks

import (
	"log"
	"net/rpc"

	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
)

type TxnConn interface {
	Send(cmd common.Command) *common.Response
	Close() error
}

type txnServe struct {
	client *rpc.Client
}

func (t *txnServe) connect() {
	var err error
	for {
		t.client, err = rpc.Dial("tcp", common.CFG.TxnServer.Url)
		if err != nil {
			log.Println("FAILED TO CONNECT TO TRANSACTION SERVER")
			continue
		}
		break
	}
}

func (t *txnServe) Send(cmd common.Command) *common.Response {
	for {
		var resp common.Response
		err := t.client.Call("TxnRPC."+common.Commands[cmd.C_type], cmd, &resp)
		if err == rpc.ErrShutdown {
			t.client, err = rpc.Dial("tcp", common.CFG.TxnServer.Url)
			continue
		}
		return &resp
	}
}

func (t *txnServe) Close() error {
	return t.client.Close()
}

func GetTxnConn() TxnConn {
	t := &txnServe{}
	t.connect()
	return t
}
