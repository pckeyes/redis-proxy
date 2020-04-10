package cache

import (
	"sync"
	"time"
)

var checkInterval = time.Duration(10 * time.Millisecond)

// ExpiryQueue a queue containing structs of key and expiration time
type ExpiryQueue struct {
	head  *KeyExpiry
	tail  *KeyExpiry
	mutex *sync.Mutex
}

// KeyExpiry contains data for expiring keys and pointer to next
type KeyExpiry struct {
	key        string
	expiration time.Time
	next       *KeyExpiry
}

// newExpiryQueue creates an empty ExpiryQueue
func newExpiryQueue() *ExpiryQueue {
	return &ExpiryQueue{
		mutex: &sync.Mutex{},
	}
}

func (e *ExpiryQueue) peek() *KeyExpiry {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	return e.head
}

// pop removes and returns head of linked list
func (e *ExpiryQueue) pop() *KeyExpiry {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	currHead := e.head
	if currHead != nil {
		e.head = e.head.next
		if e.head == nil {
			e.tail = nil
		}
	}
	return currHead
}

func (e *ExpiryQueue) push(keyExpiry *KeyExpiry) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.head == nil { //if the linked list is empty
		e.head = keyExpiry
		e.tail = keyExpiry
	} else { //if the linked list is not empty
		oldTail := e.tail
		oldTail.next = keyExpiry
		e.tail = keyExpiry
	}

}

func (e *ExpiryQueue) clear() {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.head = nil
	e.tail = nil
}
