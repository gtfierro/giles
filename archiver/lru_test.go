package archiver

import (
	"testing"
)

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
