package pokecache

import (
	"sync"
	"time"
)

type Cache struct {
	entry map[string]cacheEntry
	mu    *sync.RWMutex
	done  chan bool
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

func NewCache(interval time.Duration) *Cache {
	c := &Cache{
		entry: map[string]cacheEntry{},
		mu:    &sync.RWMutex{},
		done:  make(chan bool),
	}

	go c.reapLoop(interval)
	return c
}

func (c Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entry[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
}

func (c Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entry[key]
	return entry.val, ok
}

func (c *Cache) Stop() {
	c.done <- true
}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.mu.Lock()
			for k, v := range c.entry {
				if interval < time.Since(v.createdAt) {
					delete(c.entry, k)
				}
			}
			c.mu.Unlock()
		}
	}
}
