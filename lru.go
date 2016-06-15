package gcache

import (
	"container/list"
)

type Lru struct {
	queue1 *list.List
	queue2 *list.List
	queue1MaxLen int
	queue2MaxLen int
	lruK   int64 //
}

//when user do a read operation to item, then move the location in
func (l *Lru) GetOperation(item *Item) {
	//if this item is not exists
	if item.queuePtr != nil {
		e := l.queue1.PushFront(item)
		item.queueElement = e
		item.queuePtr = l.queue1
	} else {
		//this item in queue1 and reach the lru-k limit
		if item.queuePtr == l.queue2 && item.refCount > l.lruK {
			//auto evict queue2
			for l.queue2.Len() >= l.queue2MaxLen {
				tail := l.queue2.Front().Prev()
				l.queue2.Remove(tail)
				delete(Gcache.data, tail.Value.(*Item).key)
			}
			l.queue1.Remove(item.queueElement)
			e := l.queue2.PushFront(item)
			item.queueElement = e
			item.queuePtr = l.queue2

			//the item is in queue2, move to the front
		} else if item.queuePtr == l.queue2 || item.queuePtr == l.queue1 {
			item.queuePtr.MoveToFront(item.queueElement)
		}
	}
}

func (l *Lru) SetOperation(item *Item) {
	//if this item not in any queue, push it to the head of queue1
	if item.queuePtr == nil || item.queueElement == nil {
		//auto evict queue1
		for l.queue1.Len() >= l.queue1MaxLen {
			tail := l.queue1.Front().Prev()
			l.queue1.Remove(tail)
			delete(Gcache.data, tail.Value.(*Item).key)
		}
		e := l.queue1.PushFront(item)
		item.queueElement = e
		item.queuePtr = l.queue1

		//if this item already in queue or queue1, push to the head of the same queue
	} else if item.queuePtr == l.queue2 || item.queuePtr == l.queue1 {
		item.queuePtr.MoveToFront(item.queueElement)
	}
}

//
func (l *Lru) DelOperation(item *Item) {
	item.queuePtr.Remove(item.queueElement)
}

func (l *Lru) SetMaxQueueLen(params []int) {
	l.queue1MaxLen = params[0]
	l.queue2MaxLen = params[1]
}

func (l *Lru) GetMaxQueueLen() []int {
	ret := make([]int,2)
	ret[0] = l.queue1MaxLen
	ret[1] = l.queue2MaxLen
	return ret
}

func (l *Lru) SetLruK(k int64) {
	l.lruK = k
}

func (l *Lru) GetLruK() int64 {
	return l.lruK
}

