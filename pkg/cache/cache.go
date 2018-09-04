/*
Package cache is a simple global cache implementation
call StartCache() to start the cache thread. then use
	AddToCache, FlushCache, RemoveFromCache and FetchFromCache as necessary
*/
package cache

import (
	"sync"
	"time"
)

//cache settings
const (
	cacheChannelBufferSize = 100
	//if the cache is idle for this period of time, update cache objects
	cacheInactiveUpdateTime = 1 * time.Millisecond
	//if no cache update has happened in this amount of time, force a cache update
	//even if the cache is not idle
	cacheMaxUpdateInterval = 1 * time.Second
)

// variable cache settings
var cacheSettingsLock = sync.Mutex{}
var (
	cacheTTL = 60 * time.Second
)

var cacheMap map[string]*cacheObject
var cacheChannel chan cacheChannelMsg

// cache messaging format
// operation:
//	the function pointed to by operation is run on the cache goroutine
// callbackChan:
//	if not nil then the return value of operation is
// 	sent through the callback channel
type cacheChannelMsg struct {
	operation    func() interface{}
	callbackChan chan interface{}
}

type cacheObject struct {
	object    interface{}
	timeStamp time.Time
	ttl       time.Duration
}

//-------------------- front end methods --------------------------------

/*
StartCache starts cache management thread. CALL THIS FIRST
*/
func StartCache() {
	cacheChannel = make(chan cacheChannelMsg, cacheChannelBufferSize)
	cacheMap = make(map[string]*cacheObject)

	go cacheMain(cacheChannel)
}

/*
AddToCache caches a object under the given cacheType + name combination.
*/
func AddToCache(cacheType string, name string, object interface{}) {
	msg := cacheChannelMsg{}
	msg.operation = func() interface{} {
		cacheSettingsLock.Lock()
		defer cacheSettingsLock.Unlock()
		addToCache(cacheType, name, cacheTTL, object)
		return nil
	}
	cacheChannel <- msg
}

/*
AddToCacheTTLOverride is like AddToCache but allows overriding of global ttl value.
*/
func AddToCacheTTLOverride(cacheType string, name string, ttl time.Duration, object interface{}) {
	msg := cacheChannelMsg{}
	msg.operation = func() interface{} {
		addToCache(cacheType, name, ttl, object)
		return nil
	}
	cacheChannel <- msg
}

/*
UpdateCacheTTL set a new ttl for cache objects (objects already in cache uneffected)
*/
func UpdateCacheTTL(newTTL time.Duration) {
	cacheSettingsLock.Lock()
	defer cacheSettingsLock.Unlock()
	cacheTTL = newTTL
}

/*
FlushCache removes all objects from the cache
*/
func FlushCache() {
	msg := cacheChannelMsg{}
	msg.operation = func() interface{} {
		flushCache()
		return nil
	}
	cacheChannel <- msg
}

/*
RemoveFromCache removes the object denoted by cacheType + name from the cache.
*/
func RemoveFromCache(cacheType string, name string) {
	msg := cacheChannelMsg{}
	msg.operation = func() interface{} {
		removeFromCache(cacheType, name)
		return nil
	}
	cacheChannel <- msg
}

/*
FetchFromCache attempts to fetch the object denoted by cacheType + name from the cache.
If such an object does not exist in the cache nil is returned.
*/
func FetchFromCache(cacheType string, name string) interface{} {
	msg := cacheChannelMsg{}
	msg.operation = func() interface{} {
		return fetchFromCache(cacheType, name)
	}
	msg.callbackChan = make(chan interface{})
	cacheChannel <- msg
	return <-msg.callbackChan
}

// ------------------------- back end methods -----------------------------------
func cacheMain(cacheChan <-chan cacheChannelMsg) {
	lstUpdateTime := time.Now()
	for {
		select {
		case op := <-cacheChan:
			doCacheOp(op)
		case <-time.After(cacheInactiveUpdateTime):
			updateTTL()
			lstUpdateTime = time.Now()
		}

		if time.Since(lstUpdateTime) > cacheMaxUpdateInterval {
			updateTTL()
			lstUpdateTime = time.Now()
		}
	}
}

func doCacheOp(op cacheChannelMsg) {
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

func createCacheObject(ttl time.Duration, object interface{}) *cacheObject {
	return &cacheObject{object, time.Now(), ttl}
}

func addToCache(cacheType string, name string, ttl time.Duration, object interface{}) {
	if cacheMap[cacheType+name] == nil {
		cacheMap[cacheType+name] = createCacheObject(ttl, object)
	}
}

func flushCache() {
	cacheMap = make(map[string]*cacheObject)
}

func removeFromCache(cacheType string, name string) {
	delete(cacheMap, cacheType+name)
}

func fetchFromCache(cacheType string, name string) interface{} {
	obj := cacheMap[cacheType+name]
	if obj != nil {
		obj.timeStamp = time.Now()
		return obj.object
	}
	return nil
}
