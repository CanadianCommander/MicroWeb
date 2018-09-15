package cache_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/CanadianCommander/MicroWeb/pkg/cache"
)

func TestCacheAdd(t *testing.T) {
	cache.StartCache()

	var magicNum = 42
	cache.AddToCache(cache.CacheTypeResource, "test123", &magicNum)

	magicNumFromCache := cache.FetchFromCache(cache.CacheTypeResource, "test123")
	if magicNumFromCache == nil {
		t.Fail()
	}
}

func TestCacheRemove(t *testing.T) {
	TestCacheAdd(t)

	cache.RemoveFromCache(cache.CacheTypeResource, "test123")

	magicNumFromCache := cache.FetchFromCache(cache.CacheTypeResource, "test123")
	if magicNumFromCache != nil {
		t.Fail()
	}
}

func TestCacheFlush(t *testing.T) {
	TestCacheAdd(t)

	cache.FlushCache()
	magicNumFromCache := cache.FetchFromCache(cache.CacheTypeResource, "test123")
	if magicNumFromCache != nil {
		t.Fail()
	}
}

func TestCacheTTL(t *testing.T) {
	cache.StartCache()
	cache.UpdateCacheTTL(250 * time.Millisecond)

	var magicNum = 42
	cache.AddToCache(cache.CacheTypeResource, "test123", &magicNum)

	time.Sleep(300 * time.Millisecond)

	if cache.FetchFromCache(cache.CacheTypeResource, "test123") != nil {
		t.Fail()
	}
}

func TestCacheManyItems(t *testing.T) {
	cache.StartCache()
	cache.UpdateCacheTTL(10 * time.Millisecond)

	sTime := time.Now()
	testTime := 1 * time.Second
	for time.Since(sTime) < testTime {

		item := "foobar"
		cache.AddToCache(cache.CacheTypeResource, string(rand.Int()), &item)
	}
}

func TestCacheRemoveByType(t *testing.T){
	cache.StartCache()

	for i:=0; i<100;i++ {
		cache.AddToCache(cache.CacheTypeResource, string(i), 42)
		cache.AddToCache(cache.CacheTypeDatabase, string(i), 43)
	}

	allDB := cache.FetchAllOfType(cache.CacheTypeDatabase)
	for _, dbObj := range allDB {
		if (dbObj.(int)) != 43 {
			t.Fail()
		}
	}

	cache.FlushCacheByType(cache.CacheTypeResource)
	for i:=0; i < 100; i++ {
		if cache.FetchFromCache(cache.CacheTypeResource,  string(i)) != nil {
			t.Fail()
		}
		if cache.FetchFromCache(cache.CacheTypeDatabase, string(i)) == nil {
			t.Fail()
		}
	}
}
