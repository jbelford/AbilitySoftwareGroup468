package common

import (
	"log"
	"time"
)

type TriggerManager struct {
	c  Cache
	db *MongoDB
}

func (tm *TriggerManager) Start() {
	go func() {
		for {
			log.Println("Executing triggers...")
			trigs, err := tm.db.Triggers.GetAll()
			if err == nil {
				txns := make([]*PendingTxn, 0)
				for _, trig := range trigs {
					txn := tm.processTrigger(trig)
					if txn != nil {
						txns = append(txns, txn)
					}
				}
				if len(txns) > 0 {
					log.Printf("Resolving %d transactions\n", len(txns))
					err = tm.db.Users.BulkTransaction(txns)
					if err != nil {
						log.Println(err)
					}
					err = tm.db.Transactions.BulkLog(txns, true)
					if err != nil {
						log.Println(err)
					}
					err = tm.db.Triggers.BulkClose(txns)
					if err != nil {
						log.Println(err)
					}
				}
			}
			time.Sleep(time.Minute)
		}
	}()
}

func (tm *TriggerManager) processTrigger(t Trigger) *PendingTxn {
	quote, err := tm.c.GetQuote(t.Stock)
	if err != nil {
		return nil
	}
	isTriggered := (t.Type == "BUY" && quote.Quote <= t.When) ||
		(t.Type == "SELL" && quote.Quote >= t.When)
	if !isTriggered {
		return nil
	}
	shares := t.Shares
	if t.Type == "BUY" {
		shares = int(t.Amount / quote.Quote)
	}
	price := int64(shares) * quote.Quote
	return &PendingTxn{
		UserId:   t.UserId,
		Price:    price,
		Reserved: t.Amount,
		Shares:   shares,
		Type:     t.Type,
		Stock:    t.Stock,
	}
}

func NewTrigMan(c Cache, db *MongoDB) *TriggerManager {
	return &TriggerManager{c, db}
}
