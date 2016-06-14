package gcache

import (
	"sync"
	"time"
)

//this package implement a cache simuliar to memcached, it use 2Q  as its evict method
type cache struct {
	data       map[string]*Item
	deleteChan chan *Item //
	addChan chan *Item //
	rwLock     sync.RWMutex
	lru        *Lru
}

type Item struct {
	key        string
	value      interface{}
	expiration int64 //second of expiration , 0 means never expire
	rwLock     sync.RWMutex
	refCount   int64 //引用次数
	node       *Node
}

var Gcache = &cache{}

//init the hash map and the two queue
func init() {
	Gcache.data = make(map[string]*Item)
	Gcache.deleteChan = make(chan *Item, 1024)
	Gcache.addChan = make(chan *Item, 1024)
	Gcache.lru.delChan = make(chan *Node, 1024)
	Gcache.lru.getChan = make(chan *Node, 1024)
	Gcache.lru.setChan = make(chan *Node, 1024)
	Gcache.lru.queue1 = &Queue{id:NOQueue1, MaxLen:102400}
	Gcache.lru.queue2 = &Queue{id:NOQueue2, MaxLen:102400}
	Gcache.lru.queue1.Len = 0
	Gcache.lru.queue2.Len = 0
	/*
	Gcache.lru.queue1.addChan = make(chan *Node, 10240)
	Gcache.lru.queue1.delChan = make(chan *Node, 10240)
	Gcache.lru.queue2.addChan = make(chan *Node, 10240)
	Gcache.lru.queue2.delChan = make(chan *Node, 10240)
	*/
	Gcache.lru.daemonA()
	Gcache.lru.lruK = 1
	Gcache.daemonDelete()
}

//get the ptr of the item we want to operate
func (c *cache)getItem(key string) *Item {
	c.rwLock.RLock()
	if item, ok := c.data[key]; ok {
		c.rwLock.RUnlock()
		return item
	}
	c.rwLock.RUnlock()

	//here seems a little strange , the condition below is for thread safe
	//todo may be i can optimize this
	c.rwLock.Lock()

	// make this is concurrency safe
	if item, ok := c.data[key]; ok {
		c.rwLock.Unlock()
		return item
	}
	node := &Node{}
	item := &Item{key:key, expiration:0, node:node}
	node.item = item
	c.data[key] = item
	c.rwLock.Unlock()
	return c.data[key]
}

//todo auto evict when the list is full
func (c *cache)Set(key string, value interface{}, expiration int64) error {
	item := c.getItem(key)
	item.rwLock.Lock()
	item.value = value
	item.expiration = int64(time.Now().Unix()) + expiration
	item.rwLock.Unlock()
	c.lru.setChan <- item.node
	return nil
}

func (c *cache)Get(key string) (interface{}) {
	item := c.getItem(key)

	//expired
	if int64(time.Now().Unix()) > item.expiration {
		return nil
	}
	c.lru.getChan <- item.node
	return item.value
}

func (c *cache)Gets(keys []string) ([]interface{}) {
	var ret = make([]interface{}, 0)
	for _, key := range keys {
		val := c.Get(key)
		ret = append(ret, val)
	}
	return ret
}

func (c *cache) Delete(key string) error {
	//what if the key not exist in queue or map ??
	//so should first remove from the queue
	if item, ok := c.data[key]; ok {
		c.deleteChan <- item
	}
	return nil
}

func Append() {

}

func Replace() {

}

func Stats() {

}

func Flush() {

}

func Incr() {

}

func Decr() {

}

//the delete operation is serialized
//receive ptr from chan and do remove node from map ,also send info to queue to del
//todo
func (c *cache) daemonDelete() {
	go func() {
		for {
			select {
			case item := <-c.deleteChan :{
				if _, ok := c.data[item.key]; ok {
					delete(c.data, item.key)
				}
			}
			case item := <- c.addChan : {

			}

		}
	}
}()
}
