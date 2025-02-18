package pokecache

import (
	"fmt"
	"testing"
	"time"
)

func TestAddGet(t *testing.T) {
	const interval = 5 * time.Second
	cases := []struct {
		key string
		val []byte
	}{
		{
			key: "https://example.com",
			val: []byte("testdata"),
		},
		{
			key: "https://example.com/path",
			val: []byte("moretestdata"),
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			cache := NewCache(interval)
			cache.Add(c.key, c.val)
			val, ok := cache.Get(c.key)
			if !ok {
				t.Errorf("expected to find key")
				return
			}
			if string(val) != string(c.val) {
				t.Errorf("expected to find value")
				return
			}
		})
	}
}

func TestReapLoop(t *testing.T) {
	const baseTime = 5 * time.Millisecond
	const waitTime = baseTime + 5*time.Millisecond
	cache := NewCache(baseTime)
	cache.Add("https://example.com", []byte("testdata"))

	_, ok := cache.Get("https://example.com")
	if !ok {
		t.Errorf("expected to find key")
		return
	}

	time.Sleep(waitTime)

	_, ok = cache.Get("https://example.com")
	if ok {
		t.Errorf("expected to not find key")
		return
	}
}

func TestCacheStop(t *testing.T) {
	interval := 100 * time.Millisecond
	cache := NewCache(interval)

	// Add something to the cache
	cache.Add("test", []byte("data"))

	// Verify it's there
	_, ok := cache.Get("test")
	if !ok {
		t.Error("Expected to find data in cache")
	}

	// Wait longer than interval
	time.Sleep(interval * 2)
	// Verify it's not there
	_, ok = cache.Get("test")
	if ok {
		t.Error("Expected to not find the data in cache")
	}

	// Stop the cache
	cache.Stop()

	// Try adding something new
	cache.Add("test2", []byte("more data"))

	// Wait longer than interval
	time.Sleep(interval * 2)

	// The reaping should have stopped, so data should still be there
	_, ok = cache.Get("test2")
	if !ok {
		t.Error("Expected data to still be in cache after stop")
	}
}
