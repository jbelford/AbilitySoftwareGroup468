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
	// , setFunc func(hit bool, result interface{}) (interface{}, error)
	GetLock(key string) *sync.RWMutex
	// Convenience methods that lock before operation
	GetSync(key string, obj interface{}) error
	SetSync(key string, obj interface{})
	DeleteSync(key string)
	// Normal non-locking operations
	Get(key string, obj interface{}) error
	Set(key string, obj interface{})
	Delete(key string)
}

type cache struct {
	bcache *bigcache.BigCache
	locks  *sync.Map
}

func (c *cache) GetLock(key string) *sync.RWMutex {
	// log.Printf("Cache: Getting lock for '%s'\n", key)
	lockI, _ := c.locks.LoadOrStore(key, &sync.RWMutex{})
	lock := lockI.(*sync.RWMutex)
	// log.Printf("Cache: Got lock for '%s'\n", key)
	return lock
}

func (c *cache) GetSync(key string, obj interface{}) error {
	lock := c.GetLock(key)
	lock.RLock()
	defer lock.RUnlock()
	return c.Get(key, obj)
}

func (c *cache) SetSync(key string, obj interface{}) {
	lock := c.GetLock(key)
	lock.Lock()
	defer lock.Unlock()
	c.Set(key, obj)
}

func (c *cache) DeleteSync(key string) {
	lock := c.GetLock(key)
	lock.Lock()
	defer lock.Unlock()
	c.Delete(key)
}

func (c *cache) Get(key string, obj interface{}) error {
	data, err := c.bcache.Get(key)
	if err == nil && obj != nil {
		err = common.DecodeData(data, obj)
	}
	return err
}

func (c *cache) Set(key string, obj interface{}) {
	encoded, err := common.EncodeData(obj)
	if err == nil {
		err = c.bcache.Set(key, encoded)
		if err == nil {
			return
		}
	}
	log.Printf("Cache: '%s' Failed encoding: '%s'\n", key, err.Error())
}

func (c *cache) Delete(key string) {
	c.bcache.Delete(key)
	c.locks.Delete(key)
}

func NewCache() Cache {
	locks := sync.Map{}
	cfg := bigcache.DefaultConfig(time.Minute)
	cfg.OnRemove = func(key string, data []byte) {
		locks.Delete(key)
	}
	c, _ := bigcache.NewBigCache(cfg)
	return &cache{c, &locks}
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
	// log.Printf("CacheUtil:'%d' Getting quote '%s'\n", tid, symbol)
	key := "Quote:" + symbol
	lock := c.GetLock(key)
	quote := &common.QuoteData{}
	lock.Lock()
	defer lock.Unlock()
	err := c.Get(key, quote)
	if err != nil {
		quote, err = common.GetQuote(symbol, userId)
		if err != nil {
			// log.Printf("CacheUtil:'%d' Failed to get quote '%s'\n", tid, symbol)
			return nil, err
		}
		go c.logger.QuoteServer(quote, tid)
		c.Set(key, quote)
	}
	// log.Printf("CacheUtil:'%d' Got quote for '%s' - %d\n", tid, symbol, quote.Quote)
	return quote, nil
}

// GetReserved returns the sum of valid pending BUYs for the user
// Pending BUY's are stored in a list and are each valid for 60s
func (c *cacheUtil) GetReserved(userId string) int64 {
	key := userId + ":BUY"
	buys := []common.PendingTxn{}
	err := c.GetSync(key, &buys)
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
	err := c.GetSync(key, &sells)
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
// The txn is given a time-to-live of 10s
func (c *cacheUtil) PushPendingTxn(pending common.PendingTxn) {
	key := pending.UserId + ":" + pending.Type
	lock := c.GetLock(key)
	buys := []common.PendingTxn{}
	lock.Lock()
	defer lock.Unlock()
	err := c.Get(key, &buys)
	if err != nil {
		buys = []common.PendingTxn{pending}
	} else {
		buys = append(buys, pending)
	}
	c.Set(key, buys)
}

// PopPendingTxn removes the most recent pending transaction of the specified type (BUY or SELL)
// Returns nil if none exists
func (c *cacheUtil) PopPendingTxn(userId string, txnType string) *common.PendingTxn {
	key := userId + ":" + txnType
	buys := []common.PendingTxn{}
	err := c.GetSync(key, &buys)
	if err != nil {
		return nil
	}
	lock := c.GetLock(key)
	lock.Lock()
	defer lock.Unlock()

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
