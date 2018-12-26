package cache2go

import (
	"sync"
	"time"
)

type CacheItem struct {
	sync.RWMutex

	//item key
	key interface{}
	//item data
	data interface{}

	//How long will the item live in the cache when not being accessed/kept alive
	lifeSpan time.Duration
	//created timestamp
	createdOn time.Time
	//last access timestamp
	accessedOn time.Time
	// access count
	accessCount int64
	// callback remove item
	aboutToExpire func(key interface{})
}

// create new item
func NewCacheItem(key interface{}, lifeSpan time.Duration, data interface{}) *CacheItem {
	t := time.Now()

	return &CacheItem{
		key:      key,
		lifeSpan: lifeSpan,
		data:     data,

		createdOn:     t,
		accessedOn:    t,
		accessCount:   0,
		aboutToExpire: nil,
	}
}

// keep alive
func (item *CacheItem) KeepAlive() {
	item.Lock()
	defer item.Unlock()

	item.accessedOn = time.Now()
	item.accessCount++
}

// get item lifeSpan
func (item *CacheItem) LifeSpan() time.Duration {
	return item.lifeSpan
}

// get item accsessedon
func (item *CacheItem) AccessedOn() time.Time {
	item.Lock()
	defer item.Unlock()

	return item.accessedOn
}

// get item createdon
func (item *CacheItem) CreatedOn() time.Time {
	return item.createdOn
}

// get item accessCount
func (item *CacheItem) AccessCount() int64 {
	item.Lock()
	defer item.Unlock()

	return item.accessCount
}

// get item key
func (item *CacheItem) Key() interface{} {
	return item.key
}

// get item value
func (item *CacheItem) Data() interface{} {
	return item.data
}

// remove key callback
func (item *CacheItem) SetAboutToExpireCallback(f func(interface{})) {
	item.Lock()
	defer item.Unlock()

	item.aboutToExpire = f
}
