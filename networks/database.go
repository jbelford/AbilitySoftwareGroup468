package networks

import (
	"log"
	"time"

	"github.com/mattpaletta/AbilitySoftwareGroup468/common"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MongoDB struct {
	session      *mgo.Session
	Users        UsersCollection
	Triggers     TriggersCollection
	Transactions TransactionsCollection
	Logs         LogsCollection
}

func (db *MongoDB) Close() {
	db.session.Close()
}

type UsersCollection interface {
	// Add specified amount to a user's account
	// If the user does not exist they will also be created
	AddUserMoney(userId string, amount int64) (*common.User, error)

	// Place the amount of money reserved back into users balance
	// Returns error if user does not have the amount reserved
	UnreserveMoney(userId string, amount int64) (*common.User, error)

	// Place the amount of money from balance into reserved
	// Returns error if user does not have amount in balance
	ReserveMoney(userId string, amount int64) (*common.User, error)

	// Place number of shares from reserved for specific stock back into account
	// Returns error if user does not have that amount of shares reserved
	UnreserveShares(userId string, stock string, shares int) (*common.User, error)

	// Reserve the number of shares for a specific stock
	// Returns error if user does not have the amount of shares
	ReserveShares(userId string, stock string, shares int) (*common.User, error)

	// Returns a user or an error if not found
	GetUser(userId string) (*common.User, error)

	// Processes several pending transaction updates
	// If the transactions were cached then users' reserved fields won't be updated
	BulkTransaction(txns []*common.PendingTxn, wasCached bool) error

	// Processes one pending transaction
	// If the transaction was cached then the user's reserved fields won't be updated
	ProcessTxn(txn *common.PendingTxn, wasCached bool) (*common.User, error)
}

type users struct {
	*mgo.Collection
	s *mgo.Session
}

func (c *users) findAndModify(query interface{}, update interface{}) (*common.User, error) {
	change := mgo.Change{
		Update:    update,
		ReturnNew: true,
	}
	var result *common.User
	_, err := c.Find(query).Apply(change, &result)
	return result, err
}

func (c *users) AddUserMoney(userId string, amount int64) (*common.User, error) {
	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"balance": amount}},
		ReturnNew: true,
		Upsert:    true,
	}

	var result *common.User
	info, err := c.FindId(userId).Apply(change, &result)

	if info != nil && info.UpsertedId != nil {
		log.Printf("Created user '%s'", userId)
	}
	return result, err
}

func (c *users) UnreserveMoney(userId string, amount int64) (*common.User, error) {
	return c.findAndModify(
		bson.M{"_id": userId, "reserved": bson.M{"$gte": amount}},
		bson.M{"$inc": bson.M{
			"balance":  amount,
			"reserved": -amount,
		}})
}

func (c *users) ReserveMoney(userId string, amount int64) (*common.User, error) {
	return c.findAndModify(
		bson.M{"_id": userId, "balance": bson.M{"$gte": amount}},
		bson.M{"$inc": bson.M{
			"balance":  -amount,
			"reserved": amount,
		}})
}

func (c *users) UnreserveShares(userId string, stock string, shares int) (*common.User, error) {
	return c.findAndModify(
		bson.M{"_id": userId, "stock." + stock + ".reserved": bson.M{"$gte": shares}},
		bson.M{"$inc": bson.M{
			"stock." + stock + ".real":     shares,
			"stock." + stock + ".reserved": -shares,
		}})
}

func (c *users) ReserveShares(userId string, stock string, shares int) (*common.User, error) {
	return c.findAndModify(
		bson.M{"_id": userId, "stock." + stock + ".real": bson.M{"$gte": shares}},
		bson.M{"$inc": bson.M{
			"stock." + stock + ".real":     -shares,
			"stock." + stock + ".reserved": shares,
		}})
}

func (c *users) GetUser(userId string) (*common.User, error) {
	var user *common.User
	err := c.Find(bson.M{"_id": userId}).One(&user)
	return user, err
}

func (c *users) BulkTransaction(txns []*common.PendingTxn, wasCached bool) error {
	bulk := c.Bulk()
	for _, txn := range txns {
		var selector, update bson.M
		if txn.Type == "BUY" {
			selector, update = buyParams(txn, wasCached)
		} else {
			selector, update = sellParams(txn, wasCached)
		}
		bulk.Update(selector, update)
	}
	_, err := bulk.Run()
	return err
}

func (c *users) ProcessTxn(txn *common.PendingTxn, wasCached bool) (*common.User, error) {
	var selector, update bson.M
	if txn.Type == "BUY" {
		selector, update = buyParams(txn, wasCached)
	} else {
		selector, update = sellParams(txn, wasCached)
	}
	return c.findAndModify(selector, update)
}

func buyParams(buy *common.PendingTxn, wasCached bool) (selector bson.M, update bson.M) {
	if wasCached {
		selector = bson.M{"_id": buy.UserId, "balance": bson.M{"$gte": buy.Price}}
		update = bson.M{"$inc": bson.M{
			"balance":                      -buy.Price,
			"stock." + buy.Stock + ".real": buy.Shares,
		}}
	} else {
		selector = bson.M{"_id": buy.UserId, "reserved": bson.M{"$gte": buy.Price}}
		update = bson.M{"$inc": bson.M{
			"balance":                      buy.Reserved - buy.Price,
			"reserved":                     -buy.Reserved,
			"stock." + buy.Stock + ".real": buy.Shares,
		}}
	}
	return selector, update
}

func sellParams(sell *common.PendingTxn, wasCached bool) (selector bson.M, update bson.M) {
	realOrReserved := "real"
	if !wasCached {
		realOrReserved = "reserved"
	}
	selector = bson.M{"_id": sell.UserId, "stock." + sell.Stock + "." + realOrReserved: bson.M{"$gte": sell.Shares}}
	update = bson.M{"$inc": bson.M{
		"balance": sell.Price,
		"stock." + sell.Stock + "." + realOrReserved: -sell.Shares,
	}}
	return selector, update
}

type TriggersCollection interface {
	// Returns all the configured triggers
	GetAll() ([]common.Trigger, error)

	// Sets a trigger. If a trigger for the stock & type (buy or sell)
	// is already configured then the trigger's t.Amount will be updated.
	// t.When will be updated if the value is greater than 0
	Set(t *common.Trigger) (*common.Trigger, error)

	// Removes a trigger from the database
	// Returns an error if no trigger exists
	Cancel(userId string, stock string, trigType string) (*common.Trigger, error)

	// Gets a trigger for the user, stock, and type (buy or sell)
	// Returns error is none exists
	Get(userId string, stock string, trigType string) (*common.Trigger, error)

	// Gets all triggers of a specified user
	GetAllUser(userId string) ([]common.Trigger, error)

	// Removes several triggers corresponding to the transactions that executed
	BulkClose(txn []*common.PendingTxn) error
}

type triggers struct {
	*mgo.Collection
	s *mgo.Session
}

func (c *triggers) findAndModify(query interface{}, update interface{}) (*common.Trigger, error) {
	change := mgo.Change{
		Update:    update,
		ReturnNew: true,
		Upsert:    true,
	}
	var result *common.Trigger
	_, err := c.Find(query).Apply(change, &result)
	return result, err
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

func (c *triggers) Set(t *common.Trigger) (*common.Trigger, error) {
	set := bson.M{"amount": t.Amount}
	if t.When > 0 {
		set["when"] = t.When
	}
	return c.findAndModify(
		bson.M{"stock": t.Stock, "type": t.Type, "userId": t.UserId},
		bson.M{
			"$setOnInsert": bson.M{"stock": t.Stock, "type": t.Type, "userId": t.UserId,
				"shares": t.Shares, "transactionId": t.TransactionID},
			"$set": set,
		})
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
	// Store transaction information in the database
	LogTxn(txn *common.PendingTxn, triggered bool) (*common.Transactions, error)

	// Store bulk of transaction info in the database
	BulkLog(txns []*common.PendingTxn, triggered bool) error

	// Gets a list of transactions for a users account
	Get(userId string) (*common.Transactions, error)
}

type transactions struct {
	*mgo.Collection
	s *mgo.Session
}

func (c *transactions) findAndModify(query interface{}, update interface{}) (*common.Transactions, error) {
	change := mgo.Change{
		Update:    update,
		ReturnNew: true,
		Upsert:    true,
	}
	var result *common.Transactions
	_, err := c.Find(query).Apply(change, &result)
	return result, err
}

func (c *transactions) LogTxn(txn *common.PendingTxn, triggered bool) (*common.Transactions, error) {
	return c.findAndModify(
		bson.M{"_id": txn.UserId},
		bson.M{"$push": bson.M{"Logged": bson.M{
			"type":      txn.Type,
			"triggered": triggered,
			"stock":     txn.Stock,
			"amount":    txn.Price,
			"shares":    txn.Shares,
			"timestamp": time.Now().Unix(),
		}}})
}

func (c *transactions) BulkLog(txns []*common.PendingTxn, triggered bool) error {
	timestamp := time.Now().Unix()
	bulk := c.Bulk()
	for _, txn := range txns {
		bulk.Update(
			bson.M{"_id": txn.UserId},
			bson.M{"$push": bson.M{"Logged": bson.M{
				"type":      txn.Type,
				"triggered": triggered,
				"stock":     txn.Stock,
				"amount":    txn.Price,
				"shares":    txn.Shares,
				"timestamp": timestamp,
			}}})
	}
	_, err := bulk.Run()
	return err
}

func (c *transactions) Get(userId string) (*common.Transactions, error) {
	var txns *common.Transactions
	err := c.Find(bson.M{"_id": userId}).One(&txns)
	return txns, err
}

type LogsCollection interface {
	LogEvent(e *common.EventLog)
	GetLogs(userid string) ([]common.EventLog, error)
}

type logs struct {
	*mgo.Collection
	work  chan *common.EventLog
	flush chan bool
}

func (c *logs) wait(bulk *mgo.Bulk, limit int) bool {
	for i := 0; i < limit; i++ {
		select {
		case l := <-c.work:
			bulk.Insert(l)
		case <-c.flush:
			return true
		}
	}
	return false
}

func (c *logs) initBulkProcessing() {
	limit := 100
	c.work = make(chan *common.EventLog, limit)
	c.flush = make(chan bool)
	go func() {
		for {
			bulk := c.Bulk()
			flushed := c.wait(bulk, limit)
			bulk.Run()
			if flushed {
				c.flush <- true
			}
		}
	}()
}

func (c *logs) LogEvent(e *common.EventLog) {
	if c.work == nil {
		c.initBulkProcessing()
	}
	c.work <- e
}

func (c *logs) GetLogs(userid string) ([]common.EventLog, error) {
	if c.work == nil {
		c.initBulkProcessing()
	} else {
		c.flush <- true
		<-c.flush // Wait for flush
	}

	query := bson.M{"Username": userid}
	if userid == "admin" {
		query = bson.M{}
	}
	var result []common.EventLog
	err := c.Find(query).All(&result)
	return result, err
}

func GetMongoDatabase() *MongoDB {
	log.Println("Connecting to db using", common.CFG.Database.Url)
	for {
		session, err := mgo.Dial(common.CFG.Database.Url)
		if err != nil {
			log.Println(err)
			continue
		}
		db := session.DB(common.CFG.Database.Name)
		return &MongoDB{
			session:      session,
			Users:        &users{db.C("Users"), session},
			Triggers:     &triggers{db.C("Triggers"), session},
			Transactions: &transactions{db.C("Transactions"), session},
			Logs:         &logs{db.C("Logs"), nil, nil},
		}
	}
}
