package gcache

type QueueMeta struct {
	id      int
	addChan chan *Item
	delChan chan *Item
	Len     int64
	MaxLen  int64
	Head    *Item
	Tail    *Item
}


//serialize to add the a item to he head of each queue
func (q *QueueMeta) daemonPutItemToHead() {
	go func() {
		for {
			item := <-q.addChan
			putItemToHead(q, item)
		}
	}()
}

func putItemToHead(q *QueueMeta, item *Item) {
	// if item already in this queue
	if item.queueNo == q.id {
		//if item already in the head of the queue
		if q.Head == item {
			return
		}
		item.next = q.Head
		q.Head = item

		if item != q.Tail { // if item not in the tail of queue
			item.prev.next = item.next
			item.next.prev = item.prev

		} else { // item at the tail of the queue
			item.prev.next = nil
			q.Tail = item.prev
		}
	} else {  //if item not in this queue
		//if queue is empty
		item.queueNo = q.id
		if q.Head == nil && q.Tail == nil {
			q.Head = item
			q.Tail = item
			q.Len = 1
		} else { //if queue is not empty
			item.next = q.Head
			q.Head = item
			q.Len++
		}
	}
}

// serialize remove item from each queue
func (q *QueueMeta) daemonRemoveItemFromQueue() {
	go func() {
		for {
			item := <-q.delChan
			removeItemFromQueue(q, item)

		}
	}()
}

func removeItemFromQueue(q *QueueMeta, item *Item) {
	if item.queueNo != q.id {
		return
	}
	//if only one item in this queue
	if item.next == nil && item.prev == nil {
		q.Head = nil
		q.Tail = nil
		return
	}

	//if the current item is the first one of this queue
	if item.prev == nil {
		q.Head = item.next
		item.next.prev = nil
	} else if item.next == nil { //if the current item is the last one of this queue
		q.Tail = item.prev
		item.prev.next = nil
	} else {
		// this item in the mid of this queue
		item.prev.next = item.next
		item.next.prev = item.prev
	}
	// todo this is not concurrency safe
	q.Len--
	item.next = nil
	item.prev = nil
}





