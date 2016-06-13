package gcache

type QueueMeta struct {
	addChan chan *Item
	delChan chan *Item
	Len    int64
	MaxLen int64
	Head   *Item
	Tail   *Item
}


//serialize to add the a item to he head of each queue
func (q *QueueMeta) daemonPutItemToHead() {
	go func(){
		for {
			item := <- q.addChan
			if item.queueNo != NOQueue1 {
				continue
			}

			//when queue is empty
			if q.Head == nil && q.Tail == nil {
				q.Tail = item
				q.Head = item
				q.Len = 1
			} else { //while the queue is not empty
				item.next.prev = item
				item.next = q.Head
				q.Len ++
			}
		}
	}()
}


// serialize remove item from each queue
func (q *QueueMeta) daemonRemoveItemFromQueue(){
	go func () {
		for {
			item := <-q.delChan

			//if only one item in this queue
			if item.next == nil && item.prev == nil {
				q.Head = nil
				q.Tail = nil
				continue
			}

			//if the current item is the first one of this queue
			if item.prev == nil {
				q.Head = item.next
				item.next.prev = nil
			} else if item.next == nil {
				//if the current item is the last one of this queue
				q.Tail = item.prev
				item.prev.next = nil
			} else {
				// this item in the mid of this queue
				item.prev.next = item.next
				item.next.prev = item.prev
			}
			if item.queueNo != NOQueue2 {
				continue
			}
			item.prev.next = item.next
			item.next.prev = item.prev
		}
	}()
}




