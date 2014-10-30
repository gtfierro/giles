package main

import (
	"container/list"
)

/*
 * LRU cache for UUIDs and Metadata hashes
 *
 * Get(key string) -> interface{}, bool (value, found)
 * Set(key string, value interface{})
 * Contains(key string) -> bool
 * TODO: add locks
**/

type LRU struct {
	size     uint32
	cache    map[string]interface{}   //map key:value
	elements map[string]*list.Element //map key:element pointer
	queue    *list.List               //doubly-linked list of elements
}

func NewLRU(size uint32) *LRU {
	log.Notice("Creating new LRU with size %v", size)
	if size < 1 {
		return nil
	}
	return &LRU{size: size,
		cache:    make(map[string]interface{}),
		elements: make(map[string]*list.Element),
		queue:    list.New(),
	}
}

func (lru *LRU) Get(key string) (interface{}, bool) {
	/*
	 * We want to retrieve the value associated with the key
	 * Check the cache.
	 * If the key is in the cache, move the associated *list.Element
	 * to the front of the queue and return the value
	 * If the key is NOT in the cache, if we are at capacity, first delete the least-recently-used
	 * key/value pair. The key will be the .Back() of the queue. Remove it from cache, elements and queue
	 * Then, fetch the value using the refresh() fxn. Add it to the cache, then create a new element and move
	 * it to the front of the queue, remembering to add it to elements
	 */
	var (
		val    interface{}
		hasval bool
	)
	if val, hasval = lru.cache[key]; !hasval {
		return nil, false
	}
	// we used an item, so move to front
	lru.queue.MoveToFront(lru.elements[key])
	return val, true
}

func (lru *LRU) Set(key string, value interface{}) {
	if uint32(lru.queue.Len()) == lru.size {
		remkey := lru.queue.Remove(lru.queue.Back())
		delete(lru.cache, remkey.(string))
		delete(lru.elements, remkey.(string))
	}
	lru.cache[key] = value
	e := lru.queue.PushFront(key)
	lru.elements[key] = e
}

func (lru *LRU) Len() int {
	return lru.queue.Len()
}
