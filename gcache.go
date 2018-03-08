package gcache

import (
	"sync"
	"time"
	"container/list"
)

//this package implement a cache simuliar to memcached, it use 2Q  as its evict method
type cache struct {
	data   map[string]*Item
	rwLock sync.RWMutex
	lru    *Lru
}

type Item struct {
	key        string
	value      interface{}
	expiration int64 //second of expiration , 0 means never Expire
	rwLock     sync.RWMutex
	queuePtr   *list.List   //indicate queue this item in
	queueElement *list.Element
	refCount int64
}


var Gcache = &cache{}

//init the hash map and the two queue
func init() {
	Gcache.data = make(map[string]*Item)
	Gcache.lru = &Lru{queue1:&list.List{}, queue2:&list.List{}, lruK:1, queue1MaxLen:1024, queue2MaxLen:1024}
	Gcache.lru.queue2.Init()
	Gcache.lru.queue1.Init()

	/**
	实现主动的cache清除
	 */
	go func() {
		t := time.NewTicker(60 * time.Second)
		for {
			select {
			case <- t.C:
				Gcache.autoExpire()
			}
		}
	}()
}

//get the ptr of the item we want to operate
//
func (c *cache)getItem(key string) (*Item) {
	c.rwLock.Lock()
	defer c.rwLock.Unlock()

	if item, ok := c.data[key]; ok {
		return item
	}
	c.data[key] = &Item{key:key, expiration:0, refCount:0}
	return c.data[key]
}

func Set(key string, value interface{}, expiration int64) error {
	return Gcache.set(key, value, expiration)
}

func Get(key string) (interface{}) {
	return Gcache.get(key)
}

func MGet(keys []string) ([]interface{}) {
	return Gcache.mGet(keys)
}

func Delete(key string) error {
	return Gcache.delete(key)
}

func (c *cache) set(key string, value interface{}, expiration int64) error {
	item := c.getItem(key)
	c.rwLock.Lock()
	defer c.rwLock.Unlock()
	item.value = value
	item.expiration = int64(time.Now().Unix()) + expiration
	c.lru.SetOperation(item)
	return nil
}

func (c *cache)get(key string) (interface{}) {
	item := c.getItem(key)
	c.rwLock.Lock()
	defer c.rwLock.Unlock()
	//expired
	if int64(time.Now().Unix()) > item.expiration {
		c.delete(key)
		return nil
	}

	item.refCount++
	c.lru.GetOperation(item)
	return item.value
}


func (c *cache)mGet(keys []string) ([]interface{}) {
	var ret = make([]interface{}, 0)
	for _, key := range keys {
		val := c.get(key)
		ret = append(ret, val)
	}
	return ret
}

func (c *cache) delete(key string) error {
	c.rwLock.Lock()
	defer c.rwLock.Unlock()

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
	for key,item := range c.data {
		if int64(time.Now().Unix()) > item.expiration {
			c.delete(key)
		}
	}
}
