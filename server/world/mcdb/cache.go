package mcdb

import (
	"github.com/df-mc/goleveldb/leveldb"
	"sync"
)

var (
	// cache holds a map of path => *leveldb.DB. It is used to cache providers that are re-used for different dimensions
	// with the same underlying levelDB database.
	cache sync.Map
	// refc is a map that serves as reference counter for the *leveldb.DB instances stored in the cache variable above. A
	// *leveldb.DB instance is removed from the cache is the ref counter reaches 0.
	refc sync.Map
)

func cacheLoad(k string) (*leveldb.DB, bool) {
	if v, ok := refc.Load(k); ok {
		refc.Store(k, v.(int)+1)
		db, _ := cache.Load(k)
		return db.(*leveldb.DB), true
	}
	return nil, false
}

func cacheStore(k string, db *leveldb.DB) {
	cache.Store(k, db)
	if v, ok := refc.LoadOrStore(k, 1); ok {
		refc.Store(k, v.(int)+1)
	}
}

func cacheDelete(k string) int {
	v, _ := refc.Load(k)
	if v == 1 {
		refc.Delete(k)
		return 0
	}
	refc.Store(k, v.(int)-1)
	return v.(int) - 1
}
