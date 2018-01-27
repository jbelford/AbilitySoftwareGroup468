package tools

import (
	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
	"log"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MongoDB struct {
	session      *mgo.Session
	Users        UsersCollection
	Triggers     TriggersCollection
	Transactions TransactionsCollection
}

func (db *MongoDB) Close() {
	db.session.Close()
}

type UsersCollection interface {
	AddUserMoney(userId string, amount int64) error
	UnreserveMoney(userId string, amount int64) error
	ReserveMoney(userId string, amount int64) error
	UnreserveShares(userId string, stock string, shares int) error
	ReserveShares(userId string, stock string, shares int) error
	GetUser(userId string) (common.User, error)
	BulkTransaction(txns []*common.PendingTxn) error
	ProcessBuy(buy *common.PendingTxn) error
	ProcessSell(sell *common.PendingTxn) error
}

type users struct {
	*mgo.Collection
}

func (c *users) AddUserMoney(userId string, amount int64) error {
	info, err := c.Upsert(
		bson.M{"_id": userId},
		bson.M{
			"$setOnInsert": bson.M{"_id": userId},
			"$inc":         bson.M{"balance": amount},
		})

	if info != nil && info.UpsertedId != nil {
		log.Printf("Created user '%s'", userId)
	}
	return err
}

func (c *users) UnreserveMoney(userId string, amount int64) error {
	return c.Update(
		bson.M{"_id": userId},
		bson.M{"$inc": bson.M{
			"balance":  amount,
			"reserved": -amount,
		}})
}

func (c *users) ReserveMoney(userId string, amount int64) error {
	return c.Update(
		bson.M{"_id": userId, "balance": bson.M{"$gte": amount}},
		bson.M{"$inc": bson.M{
			"balance":  -amount,
			"reserved": amount,
		}})
}

func (c *users) UnreserveShares(userId string, stock string, shares int) error {
	return c.Update(
		bson.M{"_id": userId, "stock." + stock + ".reserved": bson.M{"$gte": shares}},
		bson.M{"$inc": bson.M{
			"stock." + stock + ".real":     shares,
			"stock." + stock + ".reserved": -shares,
		}})
}

func (c *users) ReserveShares(userId string, stock string, shares int) error {
	return c.Update(
		bson.M{"_id": userId, "stock." + stock + ".real": bson.M{"$gte": shares}},
		bson.M{"$inc": bson.M{
			"stock." + stock + ".real":     -shares,
			"stock." + stock + ".reserved": shares,
		}})
}

func (c *users) GetUser(userId string) (common.User, error) {
	var user common.User
	err := c.Find(bson.M{"_id": userId}).One(&user)
	return user, err
}

func (c *users) BulkTransaction(txns []*common.PendingTxn) error {
	bulk := c.Bulk()
	for _, txn := range txns {
		var selector, update map[string]interface{}
		if txn.Type == "BUY" {
			selector = bson.M{"_id": txn.UserId, "reserved": bson.M{"$gte": txn.Price}}
			update = bson.M{"$inc": bson.M{
				"reserved":                     -txn.Reserved,
				"stock." + txn.Stock + ".real": txn.Shares,
			}}
		} else {
			selector = bson.M{"_id": txn.UserId, "stock." + txn.Stock + ".reserved": bson.M{"$gte": txn.Shares}}
			update = bson.M{"$inc": bson.M{
				"balance":                          txn.Price,
				"stock." + txn.Stock + ".reserved": -txn.Shares,
			}}
		}
		bulk.Update(selector, update)
	}
	_, err := bulk.Run()
	return err
}

func (c *users) ProcessBuy(buy *common.PendingTxn) error {
	return c.Update(
		bson.M{"_id": buy.UserId, "reserved": bson.M{"$gte": buy.Price}},
		bson.M{"$inc": bson.M{
			"reserved":                     -buy.Price,
			"stock." + buy.Stock + ".real": buy.Shares,
		}})
}

func (c *users) ProcessSell(sell *common.PendingTxn) error {
	return c.Update(
		bson.M{"_id": sell.UserId, "stock." + sell.Stock: bson.M{"$gte": sell.Shares}},
		bson.M{"$inc": bson.M{
			"balance":                       sell.Price,
			"stock." + sell.Stock + ".real": -sell.Shares,
		}})
}

type TriggersCollection interface {
	GetAll() ([]common.Trigger, error)
	Set(t *common.Trigger) error
	Cancel(userId string, stock string, trigType string) (*common.Trigger, error)
	Get(userId string, stock string, trigType string) (*common.Trigger, error)
	GetAllUser(userId string) ([]common.Trigger, error)
	BulkClose(txn []*common.PendingTxn) error
}

type triggers struct {
	*mgo.Collection
}

func (c *triggers) GetAll() ([]common.Trigger, error) {
	var result []common.Trigger
	err := c.Find(bson.M{"when": bson.M{"$gt": 0}}).All(&result)
	return result, err
}

func (c *triggers) GetAllUser(userId string) ([]common.Trigger, error) {
	var result []common.Trigger
	err := c.Find(bson.M{"userId": userId}).All(&result)
	return result, err
}

func (c *triggers) Set(t *common.Trigger) error {
	_, err := c.Upsert(
		bson.M{"stock": t.Stock, "type": t.Type, "userId": t.UserId},
		bson.M{
			"$setOnInsert": bson.M{"stock": t.Stock, "type": t.Type, "userId": t.UserId, "shares": t.Shares},
			"$set":         bson.M{"amount": t.Amount, "when": t.When},
		})
	return err
}

func (c *triggers) Get(userId string, stock string, trigType string) (*common.Trigger, error) {
	var t *common.Trigger
	err := c.Find(bson.M{"userId": userId, "stock": stock, "type": trigType}).One(&t)
	return t, err
}

func (c *triggers) Cancel(userId string, stock string, trigType string) (*common.Trigger, error) {
	var t *common.Trigger
	_, err := c.Find(
		bson.M{"userId": userId, "stock": stock, "type": trigType},
	).Apply(mgo.Change{Remove: true}, &t)
	return t, err
}

func (c *triggers) BulkClose(txn []*common.PendingTxn) error {
	bulk := c.Bulk()
	for _, txn := range txn {
		bulk.Remove(bson.M{"userId": txn.UserId, "stock": txn.Stock, "type": txn.Type})
	}
	_, err := bulk.Run()
	return err
}

type TransactionsCollection interface {
	LogTxn(txn *common.PendingTxn, triggered bool) error
	BulkLog(txns []*common.PendingTxn, triggered bool) error
	Get(userId string) ([]common.Transaction, error)
}

type transactions struct {
	*mgo.Collection
}

func (c *transactions) LogTxn(txn *common.PendingTxn, triggered bool) error {
	return c.Insert(bson.M{
		"type":      txn.Type,
		"triggered": triggered,
		"stock":     txn.Stock,
		"amount":    txn.Price,
		"shares":    txn.Shares,
		"timestamp": time.Now().Unix(),
	})
}

func (c *transactions) BulkLog(txns []*common.PendingTxn, triggered bool) error {
	timestamp := time.Now().Unix()
	bulk := c.Bulk()
	for _, txn := range txns {
		bulk.Insert(bson.M{
			"type":      txn.Type,
			"triggered": triggered,
			"stock":     txn.Stock,
			"amount":    txn.Price,
			"shares":    txn.Shares,
			"timestamp": timestamp,
		})
	}
	_, err := bulk.Run()
	return err
}

func (c *transactions) Get(userId string) ([]common.Transaction, error) {
	var txns []common.Transaction
	err := c.Find(bson.M{"_id": userId}).All(&txns)
	return txns, err
}

func GetMongoDatabase() (*MongoDB, error) {
	log.Println("Connecting to db using", common.CFG.Database.Url)
	session, err := mgo.Dial(common.CFG.Database.Url)
	if err != nil {
		return nil, err
	}
	db := session.DB(common.CFG.Database.Name)
	return &MongoDB{
		session:      session,
		Users:        &users{db.C("Users")},
		Triggers:     &triggers{db.C("Triggers")},
		Transactions: &transactions{db.C("Transactions")},
	}, nil
}
