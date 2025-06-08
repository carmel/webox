package cache_test

import (
	"log"
	"testing"
	"time"

	"webox/cache"
)

// TestMapCache_SetWithTTL ...
func TestMapCache_SetWithTTL(t *testing.T) {

	cache.Set("hello", "nihao", time.Duration(time.Second*3))
	cache.Set("key", "sdsf", 0)
	log.Println(cache.Get("hello"))
	time.Sleep(time.Duration(3) * time.Second)
	log.Println(cache.Get("hello"))
	log.Println(cache.Get("key"))
}
