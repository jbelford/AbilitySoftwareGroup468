package common

import (
	"time"

	gcache "github.com/patrickmn/go-cache"
)

type Cache interface {
	GetQuote(symbol string) (*QuoteData, error)
	PushPendingTxn(pending PendingTxn)
	PopPendingTxn(userId string, txnType string) *PendingTxn
}

type cache struct {
	*gcache.Cache
}

func (c *cache) GetQuote(symbol string) (*QuoteData, error) {
	key := "Quote:" + symbol
	quoteI, found := c.Get(key)
	quote, ok := quoteI.(*QuoteData)
	if !ok || !found {
		quote, err := getQuote(symbol)
		if err != nil {
			return nil, err
		}
		c.Set(key, quote, time.Minute)
	}
	return quote, nil
}

func (c *cache) PushPendingTxn(pending PendingTxn) {
	key := pending.UserId + ":" + pending.Type
	buysI, found := c.Get(key)
	if buys, ok := buysI.([]PendingTxn); !ok || !found {
		c.Set(key, []PendingTxn{pending}, time.Minute)
	} else {
		c.Set(key, append(buys, pending), time.Minute)
	}
}

func (c *cache) PopPendingTxn(userId string, txnType string) *PendingTxn {
	key := userId + ":" + txnType
	buysI, found := c.Get(key)
	if !found {
		return nil
	}
	buys, ok := buysI.([]PendingTxn)
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

func NewCache() Cache {
	return &cache{gcache.New(time.Minute, 10*time.Minute)}
}
