package archiver

import (
	"sync"
)

type Element struct {
	next  *Element
	prev  *Element
	key   string
	value interface{}
}

type Cache struct {
	sync.Mutex
	size  uint32
	cache map[string]*Element
	head  *Element
	tail  *Element
}

func NewCache(size uint32) *Cache {
	return &Cache{size: size,
		cache: make(map[string]*Element, size)}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.Lock()
	defer c.Unlock()
	if elem, found := c.cache[key]; found { // if we found key in map
		elem.next = nil
		elem.prev = c.head
		c.head.next = elem
		c.head = elem
		return elem.value, true
	}
	return nil, false
}

func (c *Cache) Set(key string, value interface{}) {
	c.Lock()
	defer c.Unlock()
	e := &Element{key: key, value: value}
	switch len(c.cache) {
	case 0: // empty! add head
		c.head = e
	case 1: // have head, add tail
		e.next = c.head
		c.tail = e
		c.head.prev = e
	case int(c.size): // have max, so evict tail
		c.tail.next = e.next
		delete(c.cache, c.tail.key)
		c.tail = e
	default:
		e.next = c.tail
		c.tail.prev = e
		c.tail = e
	}
	c.cache[key] = e
}
