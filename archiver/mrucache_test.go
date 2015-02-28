package archiver

import (
	"fmt"
	"testing"
)

func TestMRUCache(t *testing.T) {
	var cache *Cache = NewCache(8)

	// Fill cache with values 0,1,2,3,4,5,6,9
	for i := 0; i < 10; i++ {
		cache.Set(fmt.Sprintf("testkey%d", i), i)
	}
	fmt.Println(cache.cache)

	h := cache.head
	for {
		if h == cache.tail {
			break
		}
		fmt.Printf("at element k %v v %v\n", h.key, h.value)
		h = h.prev
	}

	// Verify
	for i := 2; i < 10; i++ {
		val, success := cache.Get(fmt.Sprintf("testkey%d", i))
		if val != i || !success {
			t.Errorf("Missing value %d in cache", i)
		}
	}

	for i := 0; i < 2; i++ {
		_, success := cache.Get(fmt.Sprintf("testkey%d", i))
		if success {
			t.Errorf("Value %d should not be in cache", i)
		}
	}
}

func TestMRUCacheLinkHeadToTail(t *testing.T) {
	var cache *Cache = NewCache(8)
	var numValues = 8

	for i := 0; i < numValues; i++ {
		cache.Set(fmt.Sprintf("%d", i), i)
	}
	current := cache.head
	index := numValues - 1
	for {
		if current == nil {
			t.Errorf("At index %v, current is nil", index)
		}
		if current == cache.tail {
			if cache.tail.prev != nil {
				t.Errorf("Tail's prev should be nil but is key %v", cache.tail.prev.key)
			}
			if index != 0 {
				t.Errorf("Index was %v, but should be 0", index)
			}
			break
		}
		if current.value != index {
			t.Errorf("Value with key %v at index %v should have value %v but has value %v", current.key, index, index, current.value)
		}
		current = current.prev
		index--
	}
}

func TestMRUCacheLinkTailToHead(t *testing.T) {
	var cache *Cache = NewCache(8)
	var numValues = 8

	for i := 0; i < numValues; i++ {
		cache.Set(fmt.Sprintf("%d", i), i)
	}
	current := cache.tail
	index := 0

	for {
		if current == nil {
			t.Errorf("At index %v, current is nil", index)
		}
		if current == cache.head {
			if cache.head.next != nil {
				t.Errorf("Head's next should be nil but is key %v", cache.head.next.key)
			}
			if index+1 == len(cache.cache) {
				return
			} else {
				t.Errorf("Index was %v, but length of cache is %v", index+1, len(cache.cache))
			}
		}
		if current.value != index {
			t.Errorf("Value with key %v at index %v should have value %v but has value %v", current.key, index, index, current.value)
		}
		current = current.next
		index++
	}
}

func TestMRUCacheTestFirstEviction(t *testing.T) {
	var cache *Cache = NewCache(8)
	var numValues = 9

	for i := 0; i < numValues; i++ {
		cache.Set(fmt.Sprintf("%d", i), i)
	}
	current := cache.tail
	index := 0

	for {
		if current == nil {
			t.Errorf("At index %v, current is nil", index)
		}
		if current == cache.head {
			if cache.head.next != nil {
				t.Errorf("Head's next should be nil but is key %v", cache.head.next.key)
			}
			if index+1 == len(cache.cache) {
				return
			} else {
				t.Errorf("Index was %v, but length of cache is %v", index+1, len(cache.cache))
			}
		}
		if current.value != index+1 {
			t.Errorf("Value with key %v at index %v should have value %v but has value %v", current.key, index, index+1, current.value)
		}
		current = current.next
		index++
	}
}
