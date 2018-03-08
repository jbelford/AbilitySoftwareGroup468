package tools

import (
	"errors"
	"fmt"
	"time"

	"github.com/mattpaletta/AbilitySoftwareGroup468/networks"

	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
	gcache "github.com/patrickmn/go-cache"
)

const (
	notExistsKey     = "UserNotExists:%s"
	userKey          = "User:%s"
	trigNotExistsKey = "TriggerNotExists:%s:%s:%s"
	userNoTrigKey    = "UserNoTrig:%s"
	triggerKey       = "Trigger:%s:%s:%s"
	txnKey           = "Transaction:%s"
)

// CacheDB a caching middleware for all regular DB operations
type CacheDB struct {
	db           *networks.MongoDB
	Users        networks.UsersCollection
	Triggers     networks.TriggersCollection
	Transactions networks.TransactionsCollection
	Logs         networks.LogsCollection
}

func (c *CacheDB) Close() {
	c.db.Close()
}

type cacheUsers struct {
	*gcache.Cache
	cln networks.UsersCollection
}

// Sets the user in cache if error is nil and also performs the extra function if it exists
func (u *cacheUsers) setUser(user *common.User, err error, ifGood func()) (*common.User, error) {
	if err != nil {
		return nil, err
	} else if ifGood != nil {
		ifGood()
	}
	key := fmt.Sprintf(userKey, user.UserId)
	u.Set(key, user, time.Minute)
	return user, nil
}

func (u *cacheUsers) AddUserMoney(userId string, amount int64) (*common.User, error) {
	user, err := u.cln.AddUserMoney(userId, amount)
	return u.setUser(user, err, func() {
		// Mark that the user exists
		key := fmt.Sprintf(notExistsKey, userId)
		u.Delete(key)
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
	if _, found := u.Get(checkKey); found {
		return nil, errors.New("User does not exist")
	}

	key := fmt.Sprintf(userKey, userId)
	userI, found := u.Get(key)
	user, ok := userI.(*common.User)
	if !ok || !found {
		var err error
		if user, err = u.cln.GetUser(userId); err != nil {
			// Mark that the user doesn't exist for next time
			u.Set(checkKey, true, time.Minute)
			return nil, err
		}
		u.Set(key, user, time.Minute)
	}
	return user, nil
}

func (u *cacheUsers) BulkTransaction(txns []*common.PendingTxn, wasCached bool) error {
	for _, txn := range txns {
		key := fmt.Sprintf(userKey, txn.UserId)
		u.Delete(key)
	}
	return u.cln.BulkTransaction(txns, wasCached)
}

func (u *cacheUsers) ProcessTxn(txn *common.PendingTxn, wasCached bool) (*common.User, error) {
	user, err := u.cln.ProcessTxn(txn, wasCached)
	return u.setUser(user, err, nil)
}

type cacheTrig struct {
	cache *gcache.Cache
	cln   networks.TriggersCollection
}

func (ct *cacheTrig) setTrigger(trig *common.Trigger, err error, ifGood func()) (*common.Trigger, error) {
	if err != nil {
		return nil, err
	} else if ifGood != nil {
		ifGood()
	}
	key := fmt.Sprintf(triggerKey, trig.UserId, trig.Type, trig.Stock)
	ct.cache.Set(key, trig, time.Minute)
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
		ct.cache.Delete(key)
		key = fmt.Sprintf(userNoTrigKey, trig.UserId)
		ct.cache.Delete(key)
	})
}

func (ct *cacheTrig) Cancel(userId string, stock string, trigType string) (*common.Trigger, error) {
	checkKey := fmt.Sprintf(trigNotExistsKey, userId, trigType, stock)
	if _, found := ct.cache.Get(checkKey); found {
		return nil, errors.New("Trigger does not exist")
	}

	trig, err := ct.cln.Cancel(userId, stock, trigType)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf(triggerKey, userId, trigType, stock)
	ct.cache.Delete(key)
	return trig, err
}

func (ct *cacheTrig) Get(userId string, stock string, trigType string) (*common.Trigger, error) {
	checkKey := fmt.Sprintf(trigNotExistsKey, userId, trigType, stock)
	if _, found := ct.cache.Get(checkKey); found {
		return nil, errors.New("Trigger does not exist")
	}

	key := fmt.Sprintf(triggerKey, userId, trigType, stock)
	trigI, found := ct.cache.Get(key)
	trig, ok := trigI.(*common.Trigger)
	if !ok || !found {
		var err error
		if trig, err = ct.cln.Get(userId, stock, trigType); err != nil {
			ct.cache.Set(checkKey, true, time.Minute)
			return nil, err
		}
		ct.cache.Set(key, trig, time.Minute)
	}
	return trig, nil
}

func (ct *cacheTrig) GetAllUser(userId string) ([]common.Trigger, error) {
	checkKey := fmt.Sprintf(userNoTrigKey, userId)
	if _, found := ct.cache.Get(checkKey); found {
		return []common.Trigger{}, nil
	}

	trigs, err := ct.cln.GetAllUser(userId)
	if err != nil {
		return nil, err
	} else if len(trigs) == 0 {
		ct.cache.Set(checkKey, true, time.Minute)
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
		ct.cache.Delete(key)
	}
	return ct.cln.BulkClose(txn)
}

type cacheTxns struct {
	cache *gcache.Cache
	cln   networks.TransactionsCollection
}

func (ct *cacheTxns) LogTxn(txn *common.PendingTxn, triggered bool) (*common.Transactions, error) {
	newTxn, err := ct.cln.LogTxn(txn, triggered)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf(txnKey, newTxn.UserId)
	ct.cache.Set(key, newTxn, time.Minute)
	return newTxn, nil
}

func (ct *cacheTxns) BulkLog(txns []*common.PendingTxn, triggered bool) error {
	for _, t := range txns {
		key := fmt.Sprintf(txnKey, t.UserId)
		ct.cache.Delete(key)
	}
	return ct.cln.BulkLog(txns, triggered)
}

func (ct *cacheTxns) Get(userId string) (*common.Transactions, error) {
	key := fmt.Sprintf(txnKey, userId)
	txnI, found := ct.cache.Get(key)
	txn, ok := txnI.(*common.Transactions)
	if !ok || !found {
		var err error
		if txn, err = ct.cln.Get(userId); err != nil {
			return nil, err
		}
		ct.cache.Set(key, txn, time.Minute)
	}
	return txn, nil
}

type cacheLogs struct {
	cache *gcache.Cache
	cln   networks.LogsCollection
}

func (cl *cacheLogs) LogEvent(e *common.EventLog) {
	cl.LogEvent(e)
}

func (cl *cacheLogs) GetLogs(userid string) ([]common.EventLog, error) {
	return cl.cln.GetLogs(userid)
}

func NewCacheDB() *CacheDB {
	db := networks.GetMongoDatabase()
	cache := gcache.New(time.Minute, 10*time.Minute)
	return &CacheDB{
		db:           db,
		Users:        &cacheUsers{cache, db.Users},
		Triggers:     &cacheTrig{cache, db.Triggers},
		Transactions: &cacheTxns{cache, db.Transactions},
		Logs:         &cacheLogs{cache, db.Logs},
	}
}
