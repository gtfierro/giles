package main

import (
	"testing"
)

func refresh_string(key string) string {
	return key + "value"
}

func TestGet(t *testing.T) {
	lru := NewLRU(uint32(4), refresh_string)

	val := refresh_string("asdf")
	if val != "asdfvalue" {
		t.Error("refresh_string does not correctly append 'value'")
	}

	val = lru.Get("asdf")
	if val != "asdfvalue" {
		t.Error("LRU.Get does not return correct value")
	}

	if lru.cache["asdf"] != "asdfvalue" {
		t.Error("LRU.cache does not contain key/value after Get")
	}

	if lru.queue.Front().Value.(string) != "asdf" {
		t.Error("LRU.queue does not contain key: asdf")
	}
}

func TestEviction(t *testing.T) {
	lru := NewLRU(2, refresh_string)
	var val1, val2, val3 string
	val1 = lru.Get("a")
	if val1 != "avalue" {
		t.Error("lru.Get: val1 should be avalue but is", val1)
	}

	val2 = lru.Get("b")
	if val2 != "bvalue" {
		t.Error("lru.Get: val2 should be bvalue but is", val2)
	}

	val3 = lru.Get("c")
	if val3 != "cvalue" {
		t.Error("lru.Get: val3 should be cvalue but is", val3)
	}

	if len(lru.cache) != 2 {
		t.Error("lru.Cache size should be 2 but is", len(lru.cache))
	}
	if lru.queue.Len() != 2 {
		t.Error("lru.queue len should be 2 but is", lru.queue.Len())
	}
	if lru.queue.Front().Value.(string) != "c" {
		t.Error("Most recently used item should be 'c' but is", lru.queue.Front().Value.(string))
	}
	if lru.queue.Back().Value.(string) != "b" {
		t.Error("Most recently used item should be 'b' but is", lru.queue.Back().Value.(string))
	}
}
