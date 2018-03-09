package tools

import (
	"log"
	"time"

	"github.com/allegro/bigcache"

	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
)

type Cache interface {
	Get(key string, obj interface{}) error
	Set(key string, obj interface{})
	Delete(key string)
}

type cache struct {
	bcache *bigcache.BigCache
}

func (c *cache) Get(key string, obj interface{}) error {
	data, err := c.bcache.Get(key)
	if err == nil && obj != nil {
		err = common.DecodeData(data, obj)
	}
	return err
}

func (c *cache) Set(key string, obj interface{}) {
	if encoded, err := common.EncodeData(obj); err == nil {
		if err = c.bcache.Set(key, encoded); err != nil {
			log.Println(err)
		}
	}
}

func (c *cache) Delete(key string) {
	c.bcache.Delete(key)
}

func NewCache() Cache {
	c, _ := bigcache.NewBigCache(bigcache.DefaultConfig(time.Minute))
	return &cache{c}
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
	err := c.Get(key, quote)
	if err != nil {
		quote, err = common.GetQuote(symbol, userId)
		if err != nil {
			return nil, err
		}
		go c.logger.QuoteServer(quote, tid)
		c.Set(key, quote)
	}
	return quote, nil
}

// GetReserved returns the sum of valid pending BUYs for the user
// Pending BUY's are stored in a list and are each valid for 60s
func (c *cacheUtil) GetReserved(userId string) int64 {
	key := userId + ":BUY"
	buys := []common.PendingTxn{}
	err := c.Get(key, &buys)
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
	err := c.Get(key, &sells)
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
	err := c.Get(key, &buys)
	if err != nil {
		c.Set(key, []common.PendingTxn{pending})
	} else {
		c.Set(key, append(buys, pending))
	}
}

// PopPendingTxn removes the most recent pending transaction of the specified type (BUY or SELL)
// Returns nil if none exists
func (c *cacheUtil) PopPendingTxn(userId string, txnType string) *common.PendingTxn {
	key := userId + ":" + txnType
	buys := []common.PendingTxn{}
	err := c.Get(key, &buys)
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
