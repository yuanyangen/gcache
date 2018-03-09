package gcache

import (
	"container/list"
	"strings"
	"sync"
	"time"
)

//this package implement a cache simuliar to memcached, it use 2Q  as its evict method
type cache struct {
	data map[string]*Item
	lock sync.Mutex
	lru  *Lru
}

type Item struct {
	key          string
	value        interface{}
	expiration   int64 //second of expiration , 0 means never Expire
	queuePtr     *list.List //indicate queue this item in
	queueElement *list.Element
	refCount     int64
}

var Gcache = &cache{}

//init the hash map and the two queue
func init() {
	Gcache.data = make(map[string]*Item)
	Gcache.lru = &Lru{queue1: &list.List{}, queue2: &list.List{}, lruK: 1, queue1MaxLen: 1024, queue2MaxLen: 1024}
	Gcache.lru.queue2.Init()
	Gcache.lru.queue1.Init()

	/**
	实现主动的cache清除
	*/
	go func() {
		t := time.NewTicker(60 * time.Second)
		for {
			select {
			case <-t.C:
				Gcache.autoExpire()
			}
		}
	}()
}

func Set(key string, value interface{}, expiration int64) error {
	return Gcache.set(key, value, expiration)
}

func Get(key string) interface{} {
	return Gcache.concurrentGet(key)
}

func MGet(keys []string) []interface{} {
	return Gcache.mGet(keys)
}

func ScanWithPrefix(prefix string) []interface{} {
	return Gcache.scanWithPrefix(prefix)
}

func Delete(key string) error {
	return Gcache.concurrentDelete(key)
}

func Dump() map[string]interface{} {
	return Gcache.dump()
}

func (c *cache) set(key string, value interface{}, expiration int64) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.data[key]; !ok {
		c.data[key] = &Item{key: key, expiration: 0, refCount: 0}
	}
	item := c.data[key]
	item.value = value
	item.expiration = int64(time.Now().Unix()) + expiration
	c.lru.SetOperation(item)
	return nil
}


func (c *cache) concurrentGet(key string) interface{} {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.get(key)
}
func (c *cache) get(key string) interface{} {
	//key不存在
	if item, ok := c.data[key]; ok {
		//key已经过期
		if int64(time.Now().Unix()) > item.expiration {
			c.delete(key)
			return nil
		} else {
			//正常返回
			item.refCount++
			c.lru.GetOperation(item)
			return item.value
		}
	} else {
		return nil
	}
}

func (c *cache) mGet(keys []string) []interface{} {
	c.lock.Lock()
	defer c.lock.Unlock()

	var ret = make([]interface{}, 0)
	for _, key := range keys {
		val := c.get(key)
		ret = append(ret, val)
	}
	return ret
}

func (c *cache) scanWithPrefix(prefix string) []interface{} {
	c.lock.Lock()
	defer c.lock.Unlock()
	var ret = make([]interface{}, 0)
	for k := range c.data {
		if strings.HasPrefix(k, prefix) {
			v := c.get(k)
			if v != nil {
				ret = append(ret, v)
			}
		}
	}
	return ret
}

func (c *cache) concurrentDelete(key string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.delete(key)
}

func (c *cache) delete(key string) error {
	if item, ok := c.data[key]; ok {
		delete(c.data, key)
		c.lru.DelOperation(item)
	}
	return nil
}

/**
遍历所有的key， 实现代码的自动清除
*/
func (c *cache) autoExpire() {
	for key, item := range c.data {
		if int64(time.Now().Unix()) > item.expiration {
			c.lock.Lock()
			c.delete(key)
			c.lock.Unlock()
		}
	}
}

func (c *cache) dump() map[string]interface{} {
	ret := make(map[string]interface{})
	c.lock.Lock()
	defer c.lock.Unlock()
	for k := range c.data {
		v := c.get(k)
		if v != nil {
			ret[k] = v
		}
	}
	return ret
}
