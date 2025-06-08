package cache

import (
	"sync"
	"time"
)

/*MapCache MapCache */
type MapCache struct {
	cache sync.Map
}

type cachedData struct {
	val any
	ttl time.Time
}

// NewMapCache ...
func NewMapCache() *MapCache {
	return &MapCache{}
}

/*Set check exist */
func (m *MapCache) Add(key string, val any) {
	m.Set(key, val, 0)
}

/*GetD get interface with default */
func (m *MapCache) Get(key string) any {
	if v, b := m.cache.Load(key); b {
		switch vv := v.(type) {
		case *cachedData:
			if !vv.ttl.IsZero() && vv.ttl.Before(time.Now()) {
				m.cache.Delete(key)
				return nil
			}
			return vv.val
		}
	}
	return nil
}

/*Set set interface with ttl */
func (m *MapCache) Set(key string, val any, duration time.Duration) {
	var ttl time.Time
	if duration != 0 {
		ttl = time.Now().Add(duration)
	}
	m.cache.Store(key, &cachedData{
		val: val,
		ttl: ttl,
	})
}

/*Has check exist */
func (m *MapCache) Has(key string) bool {
	if v, b := m.cache.Load(key); b {
		switch vv := v.(type) {
		case *cachedData:
			if !vv.ttl.IsZero() && vv.ttl.Before(time.Now()) {
				return false
			}
			return true
		}
	}
	return false
}

/*Delete one value */
func (m *MapCache) Delete(key string) {
	m.Delete(key)
}

/*Clear delete all values */
func (m *MapCache) Clear() {
	*m = MapCache{}
}
