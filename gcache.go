package gcache

import (
	"sync"
	"time"
)

//this package implement a cache simuliar to memcached, it use 2Q  as its evict method
type cache struct {
	data       map[string]*Item
	queue1     *QueueMeta //the meta data of the first queue
	queue2     *QueueMeta // the meta data of the second queue
	deleteChan chan *Item //
	moveChan chan *Item //
	rwLock     sync.RWMutex
	lruK int64 //
}


type Item struct {
	key        string
	value      interface{}
	expiration int64 //second of expiration , 0 means never expire
	prev       *Item
	next       *Item
	rwLock     sync.RWMutex
	queueNo    int   //indicate queue this item in
	refCount int64 //引用次数
}

const NOQueue1 = 1
const NOQueue2 = 2

var Gcache = &cache{}

//init the hash map and the two queue
func init() {
	Gcache.data = make(map[string]*Item)
	Gcache.lruK = 1
	Gcache.moveChan = make(chan *Item, 1024)
	Gcache.deleteChan = make(chan *Item, 1024)
	Gcache.queue1 = &QueueMeta{id:NOQueue1, MaxLen:102400}
	Gcache.queue2 = &QueueMeta{id:NOQueue2, MaxLen:102400}
	Gcache.queue1.Len = 0
	Gcache.queue1.addChan = make(chan *Item, 10240)
	Gcache.queue1.delChan = make(chan *Item, 10240)
	Gcache.queue2.Len = 0
	Gcache.queue2.addChan = make(chan *Item, 10240)
	Gcache.queue2.delChan = make(chan *Item, 10240)
	Gcache.queue1.daemonPutItemToHead()
	Gcache.queue2.daemonPutItemToHead()
	Gcache.daemonDelete()
	Gcache.daemonMoveItem()
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
	item.rwLock.Unlock()
	c.queue1.addChan <- item
	return nil
}

func (c *cache)Get(key string)(interface{}) {
	item := c.getItem(key)

	//expired
	if int64(time.Now().Unix()) > item.expiration {
		return nil
	}

	//remove from the first queue, then add to the second queue
	if item.refCount >= c.lruK{
		c.moveChan <- item
	} else if item.refCount > c.lruK{
		c.queue2.addChan <- item
	} else { //add this to queue1
		c.queue1.addChan <- item
	}
	item.rwLock.Lock()
	item.refCount++
	item.rwLock.Unlock()


	//add to the head of the second queue
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

//the delete operation is serialized
//receive ptr from chan and do remove node from map ,also send info to queue to del
//todo
func (c *cache) daemonDelete() {
	go func(){
		for {
			item := <-c.deleteChan
			if _, ok := c.data[item.key]; ok {
				if item.queueNo == NOQueue1 {
					c.queue1.delChan <- item
				} else if item.queueNo == NOQueue2 {
					c.queue2.delChan <- item
				}
				delete(c.data, item.key)
			}
		}
	}()
}

func (c *cache) daemonMoveItem() {
	go func() {
		for {
			item := <- c.moveChan
			removeItemFromQueue(c.queue1, item)
			putItemToHead(c.queue2, item)
		}
	}()
}
