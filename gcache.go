package gcache

import (

	"sync"
	"time"
)

//this package implement a cache simuliar to memcached, it use 2Q  as its evict method
type cache struct {
	data   map[string]*Item
	q1Meta *QueueMeta //the meta data of the first queue
	q2Meta *QueueMeta // the meta data of the second queue
	deleteChan chan *Item
	rwLock sync.RWMutex
}

type QueueMeta struct {
	addChan chan *Item
	delChan chan *Item
	Len    int64
	MaxLen int64
	Head   *Item
	Tail   *Item
}

type Item struct {
	key        string
	value      interface{}
	expiration int64 //second of expiration , 0 means never expire
	prev       *Item
	next       *Item
	rwLock     sync.RWMutex
	queueNo    int   //indicate queue this item in
}

const NOQueue1 = 1
const NOQueue2 = 2

var Gcache = &cache{}

func init() {
	Gcache.data = make(map[string]*Item)
	Gcache.q1Meta = &QueueMeta{}
	Gcache.q2Meta = &QueueMeta{}
	Gcache.q1Meta.Len = 0
	Gcache.q1Meta.addChan = make(chan *Item, 10240)
	Gcache.q1Meta.delChan = make(chan *Item, 10240)
	Gcache.q2Meta.Len = 0
	Gcache.q2Meta.addChan = make(chan *Item, 10240)
	Gcache.q2Meta.delChan = make(chan *Item, 10240)
	Gcache.daemonPutItemToHead()
	Gcache.daemonRemoveItemFromQueue()
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
	if item, ok := c.data[key]; ok {
		c.rwLock.Unlock()
		return item
	}

	c.data[key] = &Item{key:key, expiration:0}
	c.rwLock.Unlock()
	return c.data[key]
}

//todo auto evict when the list is full
func (c *cache)Set(key string, value interface{}, expiration int64) error {
	item := c.getItem(key)
	item.rwLock.Lock()
	item.value = value
	item.expiration = int64(time.Now().Unix()) + expiration
	item.queueNo = 1
	item.rwLock.Unlock()
	c.q1Meta.addChan <- item
	return nil
}

func (c *cache)Get(key string)(interface{}) {
	item := c.getItem(key)
	//if the item is in queue1 then move it to queue2
	if item.queueNo == 0 {
		return nil
	}

	//expired
	if int64(time.Now().Unix()) > item.expiration {
		return nil
	}

	//remove from the first queue
	if item.queueNo == 1 {
		c.q1Meta.delChan <- item
	}

	//add to the head of the second queue
	c.q2Meta.addChan <- item
	return item.value
}

func (c *cache)Gets(keys []string)( []interface{}) {
	var ret = make([]interface{},0)
	for _,key := range keys {
		val := c.Get(key)
		ret = append(ret, val)
	}
	return ret
}

func (c *cache) Delete(key string) error {
	//what if the key not exist in queue or map ??
	//so should first remove from the queue
	if item,ok := c.data[key]; ok {
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

