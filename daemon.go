package gcache


//serialize to add the a item to he head of each queue
func (c *cache) daemonPutItemToHead() {
	//change the queue1
	go func(){
		for {
			item := <- c.q1Meta.addChan
			if item.queueNo != 1 {
				continue
			}
			item.next = c.q1Meta.Head
			item.next.prev = item
			c.q1Meta.Head = item
			c.q1Meta.Len ++

			if c.q1Meta.Tail == nil {
				c.q1Meta.Tail = item
			}
		}
	}()
	//change the queue2
	go func(){
		for {
			item := <- c.q2Meta.addChan
			if item.queueNo != NOQueue1 {
				continue
			}
			item.next = c.q2Meta.Head
			item.next.prev = item
			c.q2Meta.Head = item
			c.q2Meta.Len ++

			if c.q2Meta.Tail == NOQueue2 {
				c.q2Meta.Tail = item
			}
		}
	}()
}

// serialize remove item from each queue
func (c *cache) daemonRemoveItemFromQueue(){
	go func() {
		for {
			item := <- c.q1Meta.delChan
			if item.queueNo != NOQueue1 {
				continue
			}
			item.prev.next = item.next
			item.next.prev = item.prev
		}

	}()

	go func() {
		for {
			item := <- c.q2Meta.delChan
			if item.queueNo != NOQueue2 {
				continue
			}
			item.prev.next = item.next
			item.next.prev = item.prev
		}
	}()
}


//receive ptr from chan and do remove node from queue and map
func (c *cache) daemonDelete() {
	go func(){
		item := <- c.deleteChan
		if _,ok := c.data[item.key]; ok {
			if item.queueNo == NOQueue1 {
				c.q1Meta.delChan <- item
			} else if item.queueNo == NOQueue2 {
				c.q2Meta.delChan <- item
			}
			delete(c.data, item.key)
		}
	}()
}


