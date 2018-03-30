package tools

import (
	"errors"
	"fmt"

	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
)

const (
	notExistsKey     = "UserNotExists:%s"
	userKey          = "User:%s"
	trigNotExistsKey = "TriggerNotExists:%s:%s:%s"
	userNoTrigKey    = "UserNoTrig:%s"
	triggerKey       = "Trigger:%s:%s:%s"
	txnKey           = "Transaction:%s"
)

type CacheMongoSession struct {
	session *MongoSession
	cache   Cache
}

func (ms *CacheMongoSession) GetSharedInstance() *CacheDB {
	mongodb := ms.session.GetSharedInstance()
	return getCacheDB(mongodb, ms.cache)
}

func (ms *CacheMongoSession) GetUniqueInstance() *CacheDB {
	mongodb := ms.session.GetUniqueInstance()
	return getCacheDB(mongodb, ms.cache)
}

func (ms *CacheMongoSession) Close() {
	ms.session.Close()
}

func GetCacheMongoSession() *CacheMongoSession {
	session := GetMongoSession()
	c := NewCache()
	return &CacheMongoSession{session, c}
}

// CacheDB a caching middleware for all regular DB operations
type CacheDB struct {
	db           *MongoDB
	Users        UsersCollection
	Triggers     TriggersCollection
	Transactions TransactionsCollection
	Logs         LogsCollection
}

func (c *CacheDB) Close() {
	c.db.Close()
}

type cacheUsers struct {
	Cache
	cln UsersCollection
}

// Sets the user in cache if error is nil and also performs the extra function if it exists
func (u *cacheUsers) setUser(user *common.User, err error, ifGood func()) (*common.User, error) {
	if err != nil {
		return nil, err
	} else if ifGood != nil {
		ifGood()
	}
	key := fmt.Sprintf(userKey, user.UserId)
	u.SetSync(key, user)
	return user, nil
}

func (u *cacheUsers) AddUserMoney(userId string, amount int64) (*common.User, error) {
	user, err := u.cln.AddUserMoney(userId, amount)
	return u.setUser(user, err, func() {
		// Mark that the user exists
		key := fmt.Sprintf(notExistsKey, userId)
		u.DeleteSync(key)
	})
}

func (u *cacheUsers) UnreserveMoney(userId string, amount int64) (*common.User, error) {
	user, err := u.cln.UnreserveMoney(userId, amount)
	return u.setUser(user, err, nil)
}

func (u *cacheUsers) ReserveMoney(userId string, amount int64) (*common.User, error) {
	user, err := u.cln.ReserveMoney(userId, amount)
	return u.setUser(user, err, nil)
}

func (u *cacheUsers) UnreserveShares(userId string, stock string, shares int) (*common.User, error) {
	user, err := u.cln.UnreserveShares(userId, stock, shares)
	return u.setUser(user, err, nil)
}

func (u *cacheUsers) ReserveShares(userId string, stock string, shares int) (*common.User, error) {
	user, err := u.cln.ReserveShares(userId, stock, shares)
	return u.setUser(user, err, nil)
}

func (u *cacheUsers) GetUser(userId string) (*common.User, error) {
	checkKey := fmt.Sprintf(notExistsKey, userId)
	if err := u.GetSync(checkKey, nil); err == nil {
		return nil, errors.New("User does not exist")
	}

	key := fmt.Sprintf(userKey, userId)
	lock := u.GetLock(key)
	lock.Lock()
	defer lock.Unlock()

	user := &common.User{}
	err := u.Get(key, &user)
	if err != nil {
		user, err = u.cln.GetUser(userId)
		if err != nil {
			u.SetSync(checkKey, true)
			return nil, err
		}
		u.Set(key, user)
	}
	return user, nil
}

func (u *cacheUsers) BulkTransaction(txns []*common.PendingTxn, wasCached bool) error {
	for _, txn := range txns {
		key := fmt.Sprintf(userKey, txn.UserId)
		u.DeleteSync(key)
	}
	return u.cln.BulkTransaction(txns, wasCached)
}

func (u *cacheUsers) ProcessTxn(txn *common.PendingTxn, wasCached bool) (*common.User, error) {
	user, err := u.cln.ProcessTxn(txn, wasCached)
	return u.setUser(user, err, nil)
}

type cacheTrig struct {
	cache Cache
	cln   TriggersCollection
}

func (ct *cacheTrig) setTrigger(trig *common.Trigger, err error, ifGood func()) (*common.Trigger, error) {
	if err != nil {
		return nil, err
	} else if ifGood != nil {
		ifGood()
	}
	key := fmt.Sprintf(triggerKey, trig.UserId, trig.Type, trig.Stock)
	ct.cache.SetSync(key, trig)
	return trig, nil
}

func (ct *cacheTrig) GetAll() ([]common.Trigger, error) {
	return ct.cln.GetAll()
}

func (ct *cacheTrig) Set(t *common.Trigger) (*common.Trigger, error) {
	trig, err := ct.cln.Set(t)
	return ct.setTrigger(trig, err, func() {
		// Mark that the trigger exists
		key := fmt.Sprintf(trigNotExistsKey, trig.UserId, trig.Type, trig.Stock)
		ct.cache.DeleteSync(key)
		key = fmt.Sprintf(userNoTrigKey, trig.UserId)
		ct.cache.DeleteSync(key)
	})
}

func (ct *cacheTrig) Cancel(userId string, stock string, trigType string) (*common.Trigger, error) {
	checkKey := fmt.Sprintf(trigNotExistsKey, userId, trigType, stock)
	if err := ct.cache.GetSync(checkKey, nil); err == nil {
		return nil, errors.New("Trigger does not exist")
	}

	trig, err := ct.cln.Cancel(userId, stock, trigType)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf(triggerKey, userId, trigType, stock)
	ct.cache.DeleteSync(key)
	return trig, err
}

func (ct *cacheTrig) Get(userId string, stock string, trigType string) (*common.Trigger, error) {
	checkKey := fmt.Sprintf(trigNotExistsKey, userId, trigType, stock)
	if err := ct.cache.GetSync(checkKey, nil); err == nil {
		return nil, errors.New("Trigger does not exist")
	}

	key := fmt.Sprintf(triggerKey, userId, trigType, stock)
	lock := ct.cache.GetLock(key)
	lock.Lock()
	defer lock.Unlock()

	t := &common.Trigger{}
	err := ct.cache.Get(key, &t)
	if err != nil {
		t, err = ct.cln.Get(userId, stock, trigType)
		if err != nil {
			ct.cache.SetSync(checkKey, true)
			return nil, err
		}
		ct.cache.Set(key, t)
	}
	return t, nil
}

func (ct *cacheTrig) GetAllUser(userId string) ([]common.Trigger, error) {
	checkKey := fmt.Sprintf(userNoTrigKey, userId)
	if err := ct.cache.GetSync(checkKey, nil); err == nil {
		return []common.Trigger{}, nil
	}

	trigs, err := ct.cln.GetAllUser(userId)
	if err != nil {
		return nil, err
	} else if len(trigs) == 0 {
		ct.cache.SetSync(checkKey, true)
	} else {
		for _, t := range trigs {
			ct.setTrigger(&t, nil, nil)
		}
	}
	return trigs, nil
}

func (ct *cacheTrig) BulkClose(txn []*common.PendingTxn) error {
	for _, t := range txn {
		key := fmt.Sprintf(triggerKey, t.UserId, t.Type, t.Stock)
		ct.cache.DeleteSync(key)
	}
	return ct.cln.BulkClose(txn)
}

type cacheTxns struct {
	cache Cache
	cln   TransactionsCollection
}

func (ct *cacheTxns) LogTxn(txn *common.PendingTxn, triggered bool) (*common.Transactions, error) {
	newTxn, err := ct.cln.LogTxn(txn, triggered)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf(txnKey, newTxn.UserId)
	ct.cache.SetSync(key, newTxn)
	return newTxn, nil
}

func (ct *cacheTxns) BulkLog(txns []*common.PendingTxn, triggered bool) error {
	for _, t := range txns {
		key := fmt.Sprintf(txnKey, t.UserId)
		ct.cache.DeleteSync(key)
	}
	return ct.cln.BulkLog(txns, triggered)
}

func (ct *cacheTxns) Get(userId string) (*common.Transactions, error) {
	key := fmt.Sprintf(txnKey, userId)
	lock := ct.cache.GetLock(key)
	lock.Lock()
	txns := &common.Transactions{}
	err := ct.cache.Get(key, &txns)
	if err != nil {
		txns, err = ct.cln.Get(userId)
		if err != nil {
			txns = &common.Transactions{UserId: userId, Logged: []common.Transaction{}}
		}
		ct.cache.Set(key, txns)
	}
	lock.Unlock()
	return txns, nil
}

type cacheLogs struct {
	cache Cache
	cln   LogsCollection
}

func (cl *cacheLogs) LogEvents(e []*common.EventLog) {
	cl.LogEvents(e)
}

func (cl *cacheLogs) GetLogs(userid string) ([]common.EventLog, error) {
	return cl.cln.GetLogs(userid)
}

func getCacheDB(db *MongoDB, c Cache) *CacheDB {
	return &CacheDB{
		db:           db,
		Users:        &cacheUsers{c, db.Users},
		Triggers:     &cacheTrig{c, db.Triggers},
		Transactions: &cacheTxns{c, db.Transactions},
		Logs:         &cacheLogs{c, db.Logs},
	}
}
