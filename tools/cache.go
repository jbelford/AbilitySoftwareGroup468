package tools

import (
	"time"

	"github.com/mattpaletta/AbilitySoftwareGroup468/common"

	gcache "github.com/patrickmn/go-cache"
)

type Cache interface {
	GetQuote(symbol string, userId string, tid int64) (*common.QuoteData, error)
	GetReserved(userId string) int64
	GetReservedShares(userId string) map[string]int
	PushPendingTxn(pending common.PendingTxn)
	PopPendingTxn(userId string, txnType string) *common.PendingTxn
}

type cache struct {
	*gcache.Cache
	logger Logger
}

func (c *cache) GetQuote(symbol string, userId string, tid int64) (*common.QuoteData, error) {
	key := "Quote:" + symbol
	quoteI, found := c.Get(key)
	quote, ok := quoteI.(*common.QuoteData)
	if !ok || !found {
		var err error
		quote, err = common.GetQuote(symbol, userId)
		if err != nil {
			return nil, err
		}
		go c.logger.QuoteServer(quote, tid)
		c.Set(key, quote, time.Minute)
	}
	return quote, nil
}

// GetReserved returns the sum of valid pending BUYs for the user
// Pending BUY's are stored in a list and are each valid for 60s
func (c *cache) GetReserved(userId string) int64 {
	key := userId + ":BUY"
	buysI, found := c.Get(key)
	if !found {
		return 0
	}
	now := time.Now()
	var total int64
	buys, ok := buysI.([]common.PendingTxn)
	if ok {
		for i := len(buys) - 1; i >= 0; i-- {
			txn := buys[i]
			if txn.Expiry.After(now) {
				total += txn.Reserved
			} else {
				break
			}
		}
	}
	return total
}

// GetReservedShares returns a mapping where the keys are stock symbols
// The values of this mapping are the sum of shares pending to be sold
func (c *cache) GetReservedShares(userId string) map[string]int {
	key := userId + ":SELL"
	sellsI, found := c.Get(key)
	if !found {
		return nil
	}
	now := time.Now()
	mapping := make(map[string]int)
	sells, ok := sellsI.([]common.PendingTxn)
	if ok {
		for i := len(sells) - 1; i >= 0; i-- {
			txn := sells[i]
			if txn.Expiry.After(now) {
				mapping[txn.Stock] += txn.Shares
			} else {
				break
			}
		}
	}
	return mapping
}

// PushPendingTxn adds a pending transaction (BUY or SELL) to the cache
// The txn is given a time-to-live of 60s
func (c *cache) PushPendingTxn(pending common.PendingTxn) {
	key := pending.UserId + ":" + pending.Type
	buysI, found := c.Get(key)
	if buys, ok := buysI.([]common.PendingTxn); !ok || !found {
		c.Set(key, []common.PendingTxn{pending}, time.Minute)
	} else {
		c.Set(key, append(buys, pending), time.Minute)
	}
}

// PopPendingTxn removes the most recent pending transaction of the specified type (BUY or SELL)
// Returns nil if none exists
func (c *cache) PopPendingTxn(userId string, txnType string) *common.PendingTxn {
	key := userId + ":" + txnType
	buysI, found := c.Get(key)
	if !found {
		return nil
	}
	buys, ok := buysI.([]common.PendingTxn)
	if !ok {
		return nil
	}
	c.Delete(key)
	now := time.Now()
	n := len(buys)
	recent := buys[n-1]
	if recent.Expiry.Before(now) {
		return nil
	}
	if n > 1 {
		newExpiry := buys[n-2].Expiry.Sub(now)
		if newExpiry >= 0 {
			c.Set(key, buys[:n-1], newExpiry)
		}
	}
	return &recent
}

func NewCache(l Logger) Cache {
	return &cache{gcache.New(time.Minute, 10*time.Minute), l}
}
