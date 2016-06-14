package gcache

type Lru struct {
	queue1  *Queue
	queue2  *Queue
	lruK    int64      //
	getChan chan *Node //
	setChan chan *Node //
	delChan chan *Node //
}

type Queue struct {
	id     int
	Len    int64
	MaxLen int64
	Head   *Node
	Tail   *Node
}

type Node struct {
	prev     *Node
	next     *Node
	item     *Item
	queueNo  int
	refCount int64
}

const NOQueue1 = 1
const NOQueue2 = 2


// all operation to the lru queue are being serialized to make sure the operation is concurrency safe
func (l *Lru)daemonA() {
	go func() {
		for {
			select {
			case node := <-l.getChan:
				{
					if node.refCount == l.lruK {
						removeNodeFromQueue(l.queue1, node)
						putNodeToHead(l.queue2, node)
					} else if node.refCount > l.lruK {
						putNodeToHead(l.queue2, node)
					} else {
						putNodeToHead(l.queue1, node)
					}
					node.refCount++
				}
			case node := <-l.setChan: {
				putNodeToHead(l.queue1, node)
				removeNodeFromQueue(l.queue2, node)
			}
			case node := <-l.delChan:
				{
					if node.queueNo == NOQueue1 {
						removeNodeFromQueue(l.queue1, node)
					} else if node.queueNo == NOQueue2 {
						removeNodeFromQueue(l.queue2, node)
					}
				}
			}
		}
	}()
}

func putNodeToHead(q *Queue, node *Node) {
	// if node already in this queue
	if node.queueNo == q.id {
		//if node already in the head of the queue
		if q.Head == node || node.prev == nil {
			return
		}
		node.next = q.Head
		node.prev = nil
		q.Head = node

		if node != q.Tail {
			// if node not in the tail of queue
			node.prev.next = node.next
			node.next.prev = node.prev

		} else {
			// node at the tail of the queue
			node.prev.next = nil
			q.Tail = node.prev
		}
	} else {
		//if node not in this queue
		//if queue is empty
		node.queueNo = q.id
		if q.Head == nil && q.Tail == nil {
			q.Head = node
			q.Tail = node
			q.Len = 1
		} else {
			//if queue is not empty
			node.next = q.Head
			q.Head = node
			q.Len++
		}
	}
}

func removeNodeFromQueue(q *Queue, node *Node) {
	if node.queueNo != q.id {
		return
	}
	//if only one node in this queue
	if node.next == nil && node.prev == nil {
		q.Head = nil
		q.Tail = nil
		return
	}

	//if the current node is the first one of this queue
	if node.prev == nil {
		q.Head = node.next
		node.next.prev = nil
	} else if node.next == nil {
		//if the current node is the last one of this queue
		q.Tail = node.prev
		node.prev.next = nil
	} else {
		// this node in the mid of this queue
		node.prev.next = node.next
		node.next.prev = node.prev
	}

	q.Len--
	node.next = nil
	node.prev = nil
}

