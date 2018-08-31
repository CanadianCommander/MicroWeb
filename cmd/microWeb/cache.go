// cache singleton. (think of this file as a object) (PS Just Found golang sync.Pool in docs....)
/*
	call StartCache() to start the cache thread. then use
		AddToCache, FlushCache, RemoveFromCache and FetchFromCache as necessary
*/
package main

import (
	"time"

	"github.com/CanadianCommander/MicroWeb/pkg/logger"
)

//cache settings
const (
	CACHE_CHANNLE_BUFFER_SIZE = 100
	//if the cache is idle for this period of time, update cache objects
	CACHE_INACTIVE_UPDATE_TIME = 1 * time.Millisecond
	//if no cache update has happened in this amount of time, force a cache update
	//even if the cache is not idle
	CACHE_UPDATE_INTERVAL_MAX = 1 * time.Second
	CACHE_OBJECT_TTL          = 60 * time.Second
)

//cache object "types"
const (
	CACHE_RESOURCE        = "static:"
	CACHE_TEMPLATE_PLUGIN = "tPlugin:"
	CACHE_API_PLUGIN      = "aPlugin:"
)

var cacheMap map[string]*CacheObject
var cacheChannel chan CacheChannelMsg = nil

// cache messaging format
// operation:
//	the function pointed to by operation is run on the cache goroutine
// callbackChan:
//	if not nil then the return value of operation is
// 	sent through the callback channel
type CacheChannelMsg struct {
	operation    func() interface{}
	callbackChan chan interface{}
}

type CacheObject struct {
	object    interface{}
	timeStamp time.Time
	ttl       time.Duration
}

//-------------------- front end methods --------------------------------
func StartCache(maxItems int) {
	cacheChannel = make(chan CacheChannelMsg, CACHE_CHANNLE_BUFFER_SIZE)
	cacheMap = make(map[string]*CacheObject)

	logger.LogVerbose("Cache Started")
	go cacheMain(maxItems, cacheChannel)
}

func AddToCache(cacheType string, name string, object interface{}) {
	msg := CacheChannelMsg{}
	msg.operation = func() interface{} {
		addToCache(cacheType, name, globalSettings.GetCacheTTL(), object)
		return nil
	}
	cacheChannel <- msg
}

func AddToCacheTTLOverride(cacheType string, name string, ttl time.Duration, object interface{}) {
	msg := CacheChannelMsg{}
	msg.operation = func() interface{} {
		addToCache(cacheType, name, ttl, object)
		return nil
	}
	cacheChannel <- msg
}

func FlushCache() {
	msg := CacheChannelMsg{}
	msg.operation = func() interface{} {
		flushCache()
		return nil
	}
	cacheChannel <- msg
}

func RemoveFromCache(cacheType string, name string) {
	msg := CacheChannelMsg{}
	msg.operation = func() interface{} {
		removeFromCache(cacheType, name)
		return nil
	}
	cacheChannel <- msg
}

func FetchFromCache(cacheType string, name string) interface{} {
	msg := CacheChannelMsg{}
	msg.operation = func() interface{} {
		return fetchFromCache(cacheType, name)
	}
	msg.callbackChan = make(chan interface{})
	cacheChannel <- msg
	return <-msg.callbackChan
}

// ------------------------- back end methods -----------------------------------
func cacheMain(maxItems int, cacheChan <-chan CacheChannelMsg) {
	lstUpdateTime := time.Now()
	for {
		select {
		case op := <-cacheChan:
			doCacheOp(op)
		case <-time.After(CACHE_INACTIVE_UPDATE_TIME):
			updateTTL()
			lstUpdateTime = time.Now()
		}

		if time.Since(lstUpdateTime) > CACHE_UPDATE_INTERVAL_MAX {
			updateTTL()
			lstUpdateTime = time.Now()
		}
	}
}

func doCacheOp(op CacheChannelMsg) {
	ret := op.operation()
	if op.callbackChan != nil {
		op.callbackChan <- ret
	}
}

//remove cache items that have out lived there ttl
func updateTTL() {
	for key, obj := range cacheMap {
		if time.Since(obj.timeStamp) > obj.ttl {
			delete(cacheMap, key)
		}
	}
}

func createCacheObject(ttl time.Duration, object interface{}) *CacheObject {
	return &CacheObject{object, time.Now(), ttl}
}

func addToCache(cacheType string, name string, ttl time.Duration, object interface{}) {
	if cacheMap[cacheType+name] == nil {
		cacheMap[cacheType+name] = createCacheObject(ttl, object)
	}
}

func flushCache() {
	cacheMap = make(map[string]*CacheObject)
}

func removeFromCache(cacheType string, name string) {
	delete(cacheMap, cacheType+name)
}

func fetchFromCache(cacheType string, name string) interface{} {
	obj := cacheMap[cacheType+name]
	if obj != nil {
		obj.timeStamp = time.Now()
		return obj.object
	} else {
		return nil
	}
}
