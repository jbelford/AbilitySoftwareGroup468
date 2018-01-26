package common

import (
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var dbConfig DatabaseConfig

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
	GetUser(userId string) (User, error)
	BulkTransaction(txns []*PendingTxn) error
	ProcessBuy(buy *PendingTxn) error
	ProcessSell(sell *PendingTxn) error
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

func (c *users) GetUser(userId string) (User, error) {
	var user User
	err := c.Find(bson.M{"_id": userId}).One(&user)
	return user, err
}

func (c *users) BulkTransaction(txns []*PendingTxn) error {
	bulk := c.Bulk()
	for _, txn := range txns {
		var selector map[string]interface{}
		if txn.Type == "BUY" {
			selector = bson.M{"_id": txn.UserId, "balance": bson.M{"$gte": txn.Price}}
			txn.Price *= -1
		} else {
			selector = bson.M{"_id": txn.UserId, "stock." + txn.Stock: bson.M{"$gte": txn.Shares}}
			txn.Shares *= -1
		}
		update := bson.M{"$inc": bson.M{
			"balance":            txn.Price,
			"stock." + txn.Stock: txn.Shares,
		}}
		bulk.Update(selector, update)
	}
	_, err := bulk.Run()
	return err
}

func (c *users) ProcessBuy(buy *PendingTxn) error {
	return c.Update(
		bson.M{"_id": buy.UserId, "balance": bson.M{"$gte": buy.Price}},
		bson.M{"$inc": bson.M{
			"balance":            -buy.Price,
			"stock." + buy.Stock: buy.Shares,
		}})
}

func (c *users) ProcessSell(sell *PendingTxn) error {
	return c.Update(
		bson.M{"_id": sell.UserId, "stock." + sell.Stock: bson.M{"$gte": sell.Shares}},
		bson.M{"$inc": bson.M{
			"balance":             sell.Price,
			"stock." + sell.Stock: -sell.Shares,
		}})
}

type TriggersCollection interface {
	GetAll() ([]Trigger, error)
}

type triggers struct {
	*mgo.Collection
}

func (c *triggers) GetAll() ([]Trigger, error) {
	var result []Trigger
	err := c.Find(bson.M{}).All(&result)
	return result, err
}

type TransactionsCollection interface{}

type transactions struct {
	*mgo.Collection
}

func GetMongoDatabase() (*MongoDB, error) {
	log.Println("Connecting to db using", dbConfig.Url)
	session, err := mgo.Dial(dbConfig.Url)
	if err != nil {
		return nil, err
	}
	db := session.DB(dbConfig.Name)
	return &MongoDB{
		session:      session,
		Users:        &users{db.C("Users")},
		Triggers:     &triggers{db.C("Triggers")},
		Transactions: &transactions{db.C("Transactions")},
	}, nil
}

func init() {
	config, err := GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	dbConfig = config.Database
}
