package common

import (
	"time"

	gcache "github.com/patrickmn/go-cache"
)

type Cache interface {
	PushPendingTxn(userId string, pending PendingTxn)
	PopPendingTxn(userId string, txnType string) *PendingTxn
}

type cache struct {
	*gcache.Cache
}

func (c *cache) PushPendingTxn(userId string, pending PendingTxn) {
	key := userId + ":" + pending.Type
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
