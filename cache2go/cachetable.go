package cache2go

import (
	"log"
	"sort"
	"sync"
	"time"
)

type CacheTable struct {
	// sync mutex
	sync.RWMutex
	//table name
	name string
	//table item
	items map[interface{}]*CacheItem

	//timer clean up
	cleanupTimer *time.Timer
	//current timer duration
	cleanupInterval time.Duration

	//logger use for table
	logger *log.Logger

	//load non-exsiting key
	loadData func(key interface{}, args ...interface{}) *CacheItem
	//callback method when add new cacheitem
	addedItem func(item *CacheItem)
	//callback method when delete cacheitem
	aboutToDeleteItem func(item *CacheItem)
}

//return cache items length
func (table *CacheTable) Count() int {
	table.RLock()
	defer table.RUnlock()
	return len(table.items)
}

//foreach cache items
func (table *CacheTable) Foreach(trans func(key interface{}, item *CacheItem)) {
	table.RLock()
	defer table.RUnlock()

	for key, item := range table.items {
		trans(key, item)
	}
}

//a callback when add new item
func (table *CacheTable) SetAddedItemCallback(f func(*CacheItem)) {
	table.RLock()
	defer table.RUnlock()

	table.addedItem = f
}

// a callback when delete item
func (table *CacheTable) SetAboutToDeleteItemCallback(f func(*CacheItem)) {
	table.RLock()
	defer table.RUnlock()

	table.aboutToDeleteItem = f
}

// set the logger this cache table
func (table *CacheTable) SetLogger(logger *log.Logger) {
	table.RLock()
	defer table.RUnlock()

	table.logger = logger
}

// Expiration check loop, triggered by self-adjusting timer
func (table *CacheTable) expirationCheck() {
	table.RLock()

	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}

	if table.cleanupInterval > 0 {
		table.log("Expiration check triggered after", table.cleanupInterval, "from table", table.name)
	} else {
		table.log("Expiration check installed for table", table.name)
	}

	now := time.Now()
	smallestDuration := 0 * time.Second
	for key, item := range table.items {
		item.RLock()
		lifeSpan := item.lifeSpan
		accessedOn := item.accessedOn
		item.RUnlock()

		if lifeSpan == 0 {
			continue
		}
		if now.Sub(accessedOn) >= lifeSpan {
			table.deleteInterval(key)
		} else {
			if smallestDuration == 0 || lifeSpan-now.Sub(accessedOn) < smallestDuration {
				smallestDuration = lifeSpan - now.Sub(accessedOn)
			}
		}
	}

	table.cleanupInterval = smallestDuration
	if smallestDuration > 0 {
		table.cleanupTimer = time.AfterFunc(smallestDuration, func() {
			go table.expirationCheck()
		})
	}

	table.RUnlock()
}

func (table *CacheTable) addInternal(item *CacheItem) {
	table.log("Adding item with key", item.key, "and lifespan of", item.lifeSpan, "to table", table.name)
	table.items[item.key] = item

	expDur := table.cleanupInterval
	addedItem := table.addedItem

	table.Unlock()
	// callback add a item
	if addedItem != nil {
		addedItem(item)
	}
	//If we haven't set up any expiration check timer
	if item.lifeSpan > 0 && (expDur == 0 || item.lifeSpan < expDur) {
		table.expirationCheck()
	}
}

// add key/data item
func (table *CacheTable) Add(key interface{}, lifeSpan time.Duration, data interface{}) *CacheItem {
	item := NewCacheItem(key, lifeSpan, data)

	table.Lock()
	table.addInternal(item)

	return item
}

func (table *CacheTable) deleteInterval(key interface{}) (*CacheItem, error) {
	r, ok := table.items[key]
	if !ok {
		return nil, ErrKeyNotFound
	}

	aboutToDeleteItem := table.aboutToDeleteItem
	table.Unlock()

	if aboutToDeleteItem != nil {
		aboutToDeleteItem(r)
	}

	r.RLock()
	defer r.RUnlock()
	if r.aboutToExpire != nil {
		r.aboutToExpire(key)
	}

	table.Lock()
	table.log("Deleting item with key", key, "created on", r.createdOn, "and hit", r.accessCount, "times from table", table.name)
	delete(table.items, key)

	return r, nil
}

//delete a item from the cache
func (table *CacheTable) Delete(key interface{}) (*CacheItem, error) {
	table.Lock()
	defer table.Unlock()

	return table.deleteInterval(key)
}

//is exist item in cache
func (table *CacheTable) Exists(key interface{}) bool {
	table.RLock()
	defer table.RUnlock()
	_, ok := table.items[key]

	return ok
}

//isexist ? false : add item
func (table *CacheTable) NotFoundAdd(key interface{}, lifeSpan time.Duration, data interface{}) bool {
	table.Lock()

	if _, ok := table.items[key]; ok {
		table.Unlock()
		return false
	}

	item := NewCacheItem(key, lifeSpan, data)
	table.addInternal(item)

	return true
}

// set load data
func (table *CacheTable) SetDataLoader(f func(interface{}, ...interface{}) *CacheItem) {
	table.Lock()
	defer table.Unlock()

	table.loadData = f
}

//Value returns an item from the cache
func (table *CacheTable) Value(key interface{}, args ...interface{}) (*CacheItem, error) {
	table.Lock()
	defer table.Unlock()

	loadData := table.loadData
	r, ok := table.items[key]
	if ok {
		r.KeepAlive()
		return r, nil
	}

	if loadData != nil {
		item := loadData(key, args...)
		if item != nil {
			table.Add(key, item.lifeSpan, item.data)
			return item, nil
		}

		return nil, ErrKeyNotFoundOrLoadable
	}

	return nil, ErrKeyNotFound
}

// flush delete all items from this cache table
func (table *CacheTable) Flush() {
	table.Lock()
	defer table.Unlock()

	table.log("flush table cache", table.name)

	table.items = make(map[interface{}]*CacheItem)
	table.cleanupInterval = 0
	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}
}

// logging func
func (table *CacheTable) log(v ...interface{}) {
	if table.logger == nil {
		return
	}

	table.logger.Println(v...)
}

type CacheItemPair struct {
	Key         interface{}
	AccessCount int64
}

type CacheItemPairList []CacheItemPair

func (p CacheItemPairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p CacheItemPairList) Len() int           { return len(p) }
func (p CacheItemPairList) Less(i, j int) bool { return p[i].AccessCount > p[j].AccessCount }

func (table *CacheTable) MostAccessed(count int64) []*CacheItem {
	table.Lock()
	defer table.Unlock()

	p := make(CacheItemPairList, len(table.items))
	i := 0

	for key, item := range table.items {
		p[i] = CacheItemPair{key, item.accessCount}
		i++
	}

	sort.Sort(p)

	var r []*CacheItem
	c := int64(0)

	for _, v := range p {
		if c >= count {
			break
		}

		item, ok := table.items[v.Key]
		if ok {
			r = append(r, item)
		}
		c++
	}

	return r
}
