package common

import (
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var dbConfig DatabaseConfig

type MongoDB struct {
	url      string
	name     string
	session  *mgo.Session
	database *mgo.Database
}

func (db *MongoDB) AddUserMoney(userId string, amount int64) error {
	info, err := db.database.C("Users").Upsert(
		bson.M{"userId": userId},
		bson.M{"$setOnInsert": bson.M{"userId": userId}, "$inc": bson.M{"balance": amount}})

	if info != nil && info.UpsertedId != nil {
		log.Printf("Created user s'%s'", userId)
	}
	return err
}

func (db *MongoDB) Close() {
	db.session.Close()
}

func GetMongoDatabase() (*MongoDB, error) {
	session, err := mgo.Dial(dbConfig.Url)
	if err != nil {
		return nil, err
	}
	return &MongoDB{dbConfig.Url, dbConfig.Name, session, session.DB(dbConfig.Name)}, nil
}

func init() {
	config, err := GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	dbConfig = config.Database
}
