package cache2go

import "sync"

var cache = make(map[string]*CacheTable)
var mutex sync.RWMutex

func Cache(table string) *CacheTable {
	mutex.RLock()
	t, ok := cache[table]
	mutex.RUnlock()

	if !ok {
		mutex.Lock()
		t, ok = cache[table]

		if !ok {
			t = &CacheTable{
				name:  table,
				items: make(map[interface{}]*CacheItem),
			}
		}
		mutex.Unlock()
	}

	return t
}
