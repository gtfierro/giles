package archiver

import (
	"fmt"
	"testing"
)

func TestMRUCache(t *testing.T) {
	var cache *Cache = NewCache(8)

	for i := 0; i < 10; i++ {
		cache.Set(fmt.Sprintf("testkey%d", i), i)
		cache.Get(fmt.Sprintf("testkey%d", i))
	}
	// Cache should have values 2,3,4,5,6,7,8,9
	fmt.Println(cache.cache)

	h := cache.head
	for {
		fmt.Printf("at element k %v v %v\n", h.key, h.value)
		if h == cache.tail {
			break
		}
		h = h.prev
	}

	// Verify
	for i := 2; i < 10; i++ {
		val, success := cache.Get(fmt.Sprintf("testkey%d", i))
		if val != i || !success {
			t.Errorf("value %d should be in cache, but is not", i)
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

func TestGet(t *testing.T) {
	lru := NewCache(uint32(4))

	val, ok := lru.Get("asdf")
	if ok != false {
		t.Error("ok should be false but is", ok)
	}
	lru.Set("asdf", "asdfvalue")
	val, ok = lru.Get("asdf")
	if val != "asdfvalue" {
		t.Error("LRU.Get does not return correct value", val)
	}

	if lru.cache["asdf"].value.(string) != "asdfvalue" {
		t.Error("LRU.cache does not contain key/value after Get")
	}

}

func TestEviction(t *testing.T) {
	lru := NewCache(2)
	val1, ok := lru.Get("a")
	if ok != false {
		t.Error("ok should be false")
	}
	lru.Set("a", "avalue")
	val1, ok = lru.Get("a")
	if ok != true {
		t.Error("ok should be true")
	}
	if val1 != "avalue" {
		t.Error("lru.Get: val1 should be avalue but is", val1)
	}

	lru.Set("b", "bvalue")
	lru.Set("c", "cvalue")

	if len(lru.cache) != 2 {
		t.Error("lru.Cache size should be 2 but is", len(lru.cache))
	}

	val1, ok = lru.Get("a")
	if ok == true {
		t.Error("a should have been evicted")
	}

	val1, ok = lru.Get("b")
	if ok != true {
		t.Error("b should have been retained")
	}

	val1, ok = lru.Get("c")
	if ok != true {
		t.Error("c should have been retained")
	}
}

/**
 * Should do the following benchmarks:
 * Insert with no Reuse (no repeats)
 * Insert with Resue (with repeats -- moving LRU item to top)
 * Parallel versions of the above, to test the effect of the Lock
 * Insert with LRU size = 1
 * Insert+Get with LRU size = 1
 * Get with LRU size = 1
**/

func BenchmarkSetSize1NoReuse(b *testing.B) {
	l := NewCache(uint32(3))
	for i := 0; i < b.N; i++ {
		l.Set(string(i), i)
	}
}

func BenchmarkSetSize1000Reuse(b *testing.B) {
	l := NewCache(uint32(1000))
	for i := 0; i < b.N; i++ {
		l.Set(string(i), i)
	}
}

func BenchmarkSetSize1Reuse(b *testing.B) {
	l := NewCache(uint32(1))
	l.Set("1", 1)
	for i := 0; i < b.N; i++ {
		l.Set("1", 1)
	}
}

func BenchmarkGetSize1(b *testing.B) {
	l := NewCache(uint32(1))
	l.Set("1", 1)
	for i := 0; i < b.N; i++ {
		l.Get("1")
	}
}
