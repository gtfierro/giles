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

	// Verify
	for i := 0; i < 7; i++ {
		val, success := cache.Get(fmt.Sprintf("testkey%d", i))
		if val != i || !success {
			t.Errorf("Missing value %d in cache", i)
		}
	}
	for i := 7; i < 9; i++ {
		_, success := cache.Get(fmt.Sprintf("testkey%d", i))
		if success {
			t.Errorf("Value %d should not be in cache", i)
		}
	}
	_, success := cache.Get(fmt.Sprintf("testkey%d", 9))
	if !success {
		t.Errorf("Value %d should not be in cache", 9)
	}
}
