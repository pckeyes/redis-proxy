package cache

import (
	"context"
	"sync"
	"time"
)

// Cache a thread-safe TTL map
type Cache struct {
	data        map[string]string
	capacity    int
	expiry      time.Duration
	mutex       *sync.Mutex
	expiryQueue *ExpiryQueue
}

// NewCache creates an empty Cache
func NewCache(capacity int, expiry time.Duration) *Cache {
	return &Cache{
		data:        make(map[string]string),
		capacity:    capacity,
		expiry:      expiry,
		mutex:       &sync.Mutex{},
		expiryQueue: newExpiryQueue(),
	}
}

// Put puts a key value pair in the Cache's map
func (c *Cache) Put(key string, value string) {
	if c.Size() == c.capacity {
		oldestKey := c.expiryQueue.pop().key
		c.deleteKey(oldestKey)
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data[key] = value
	expiration := time.Now().Add(c.expiry)
	keyExpiry := &KeyExpiry{
		key:        key,
		expiration: expiration,
		next:       nil,
	}
	c.expiryQueue.push(keyExpiry)
}

// Get gets a value from the Cache's map
func (c *Cache) Get(key string) (string, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	value, exists := c.data[key]
	return value, exists
}

// ExpiryChecker checks for and removes expired cache values
func (c *Cache) ExpiryChecker(ctx context.Context) {
	timer := time.NewTicker(checkInterval)
	defer timer.Stop()

	for {
		select {
		case <-timer.C: //timer has talked to us
			c.removeExpiredKeys()

		case <-ctx.Done():
			return
		}
	}
}

// Size returns the number of entries in the Cache's data map
func (c *Cache) Size() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return len(c.data)
}

// Clear clears all values from the cache
func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.expiryQueue.clear()
	c.data = make(map[string]string)
}

func (c *Cache) removeExpiredKeys() {
	head := c.expiryQueue.peek()
	if head != nil {
		//remove all expired keys
		for {
			if head.expiration.After(time.Now()) {
				return
			}
			expiredKey := c.expiryQueue.pop().key
			c.deleteKey(expiredKey)
			if c.expiryQueue.peek() == nil {
				return
			}
		}
	}
}

func (c *Cache) deleteKey(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.data, key)
}
