package tools

import (
	"log"
	"sync"
	"time"

	"github.com/allegro/bigcache"

	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
)

type Cache interface {
	// Thread locks access to the object at the key
	// Optional: setFunc can be used to set a new value before releasing the lock
	Get(key string, obj interface{}, setFunc func(hit bool, result interface{}) (interface{}, error)) error
	Set(key string, obj interface{})
	Delete(key string)
}

type cache struct {
	bcache *bigcache.BigCache
	locks  map[string]*sync.Mutex
	mtx    *sync.RWMutex
}

func (c *cache) getLock(key string) *sync.Mutex {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if c.locks[key] == nil {
		var newMutex sync.Mutex
		c.locks[key] = &newMutex
	}
	return c.locks[key]
}

func (c *cache) get(key string, obj interface{}) error {
	data, err := c.bcache.Get(key)
	if err == nil && obj != nil {
		err = common.DecodeData(data, obj)
	}
	return err
}

func (c *cache) set(key string, obj interface{}) {
	if encoded, err := common.EncodeData(obj); err == nil {
		if err = c.bcache.Set(key, encoded); err != nil {
			log.Println(err)
		}
	}
}

func (c *cache) Get(key string, obj interface{}, setFunc func(hit bool, result interface{}) (interface{}, error)) error {
	lock := c.getLock(key)
	lock.Lock()
	defer lock.Unlock()

	err := c.get(key, obj)
	if setFunc != nil {
		hit := err == nil
		err = nil
		newObj, err := setFunc(hit, obj)
		if err == nil && newObj != nil {
			c.set(key, newObj)
			c.get(key, obj)
		}
	}
	return err
}

func (c *cache) Set(key string, obj interface{}) {
	lock := c.getLock(key)
	lock.Lock()
	defer lock.Unlock()
	c.set(key, obj)
}

func (c *cache) Delete(key string) {
	lock := c.getLock(key)
	lock.Lock()
	defer lock.Unlock()
	c.bcache.Delete(key)
}

func NewCache() Cache {
	rwMtx := sync.RWMutex{}
	locks := make(map[string]*sync.Mutex)
	cfg := bigcache.DefaultConfig(time.Minute)
	cfg.OnRemove = func(key string, data []byte) {
		rwMtx.Lock()
		defer rwMtx.Unlock()
		delete(locks, key)
	}
	c, _ := bigcache.NewBigCache(cfg)
	return &cache{c, locks, &rwMtx}
}

type CacheUtil interface {
	GetQuote(symbol string, userId string, tid int64) (*common.QuoteData, error)
	GetReserved(userId string) int64
	GetReservedShares(userId string) map[string]int
	PushPendingTxn(pending common.PendingTxn)
	PopPendingTxn(userId string, txnType string) *common.PendingTxn
}

type cacheUtil struct {
	Cache
	logger Logger
}

func (c *cacheUtil) GetQuote(symbol string, userId string, tid int64) (*common.QuoteData, error) {
	key := "Quote:" + symbol
	quote := &common.QuoteData{}
	err := c.Get(key, quote, func(hit bool, result interface{}) (interface{}, error) {
		if !hit {
			newQuote, err := common.GetQuote(symbol, userId)
			if err != nil {
				return nil, err
			}
			go c.logger.QuoteServer(newQuote, tid)
			return newQuote, nil
		}
		return nil, nil
	})
	return quote, err
}

// GetReserved returns the sum of valid pending BUYs for the user
// Pending BUY's are stored in a list and are each valid for 60s
func (c *cacheUtil) GetReserved(userId string) int64 {
	key := userId + ":BUY"
	buys := []common.PendingTxn{}
	err := c.Get(key, &buys, nil)
	if err != nil {
		return 0
	}
	now := time.Now()
	var total int64
	for i := len(buys) - 1; i >= 0; i-- {
		txn := buys[i]
		if txn.Expiry.After(now) {
			total += txn.Reserved
		} else {
			break
		}
	}
	return total
}

// GetReservedShares returns a mapping where the keys are stock symbols
// The values of this mapping are the sum of shares pending to be sold
func (c *cacheUtil) GetReservedShares(userId string) map[string]int {
	key := userId + ":SELL"
	sells := []common.PendingTxn{}
	err := c.Get(key, &sells, nil)
	if err != nil {
		return nil
	}
	now := time.Now()
	mapping := make(map[string]int)
	for i := len(sells) - 1; i >= 0; i-- {
		txn := sells[i]
		if txn.Expiry.After(now) {
			mapping[txn.Stock] += txn.Shares
		} else {
			break
		}
	}
	return mapping
}

// PushPendingTxn adds a pending transaction (BUY or SELL) to the cache
// The txn is given a time-to-live of 60s
func (c *cacheUtil) PushPendingTxn(pending common.PendingTxn) {
	key := pending.UserId + ":" + pending.Type
	buys := []common.PendingTxn{}
	c.Get(key, &buys, func(hit bool, result interface{}) (interface{}, error) {
		var txns []common.PendingTxn
		if !hit {
			txns = []common.PendingTxn{pending}
		} else {
			list := result.(*[]common.PendingTxn)
			txns = append(*list, pending)
		}
		return txns, nil
	})
}

// PopPendingTxn removes the most recent pending transaction of the specified type (BUY or SELL)
// Returns nil if none exists
func (c *cacheUtil) PopPendingTxn(userId string, txnType string) *common.PendingTxn {
	key := userId + ":" + txnType
	buys := []common.PendingTxn{}
	err := c.Get(key, &buys, nil)
	if err != nil {
		return nil
	}
	c.Delete(key)
	now := time.Now()
	n := len(buys)
	recent := buys[n-1]
	if recent.Expiry.Before(now) {
		return nil
	}
	if n > 1 && buys[n-2].Expiry.After(now) {
		c.Set(key, buys[:n-1])
	}
	return &recent
}

func NewCacheUtil(l Logger) CacheUtil {
	c := NewCache()
	return &cacheUtil{c, l}
}
