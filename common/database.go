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
	ProcessBuy(userId string, buy *PendingTxn) error
	ProcessSell(userId string, sell *PendingTxn) error
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

func (c *users) ProcessBuy(userId string, buy *PendingTxn) error {
	return c.Update(
		bson.M{"_id": userId, "balance": bson.M{"$gte": buy.Price}},
		bson.M{"$inc": bson.M{
			"balance":            -buy.Price,
			"stock." + buy.Stock: buy.Shares,
		}})
}

func (c *users) ProcessSell(userId string, sell *PendingTxn) error {
	return c.Update(
		bson.M{"_id": userId, "stock." + sell.Stock: bson.M{"$gte": sell.Shares}},
		bson.M{"$inc": bson.M{
			"balance":             sell.Price,
			"stock." + sell.Stock: -sell.Shares,
		}})
}

type TriggersCollection interface{}

type triggers struct {
	*mgo.Collection
}

type TransactionsCollection interface{}

type transactions struct {
	*mgo.Collection
}

func GetMongoDatabase() (*MongoDB, error) {
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
