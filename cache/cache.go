package cache

import "time"

/*Cache define an cache interface */
type Cache interface {
	Add(key string, val any)
	Set(key string, val any, timeout time.Duration)
	Get(key string) any
	Delete(prefix string)
	Clear()
	Has(key string) bool
}

// var cache sync.Map
var cache Cache

func init() {
	RegisterCache(NewMapCache())
}

/*RegisterCache register cache to map */
func RegisterCache(c Cache) {
	cache = c
}

// Get get value
func Get(key string) any {
	return cache.Get(key)
}

// Set set value
func Set(key string, val any, ttl time.Duration) {
	cache.Set(key, val, ttl)
}

// Has check value
func Has(key string) bool {
	return cache.Has(key)
}

// Delete delete value
func Delete(key string) {
	cache.Delete(key)
}

// Clear clear all
func Clear() {
	cache.Clear()
}
