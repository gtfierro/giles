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
		if elem == c.tail {
			c.tail = elem.next
			if elem.next != nil {
				c.tail.prev = nil
			}
		} else if elem != c.head {
			elem.next = nil
			elem.prev = c.head
			c.head.next = elem
			c.head = elem
		}
		return elem.value, true
	}
	return nil, false
}

/* Sends the value identified by key
 * to the head of the list and adjusts
 * the pointers appropriately
 */
func (c *Cache) sendToHead(element *Element) {
	if element == c.head {
		// if already head, do nothing
		return
	} else if element == c.tail {
		// if element is the tail, set the new tail
		c.tail = c.tail.next
		c.tail.prev = nil
	}
	c.head.next = element
	element.prev = c.head
	c.head = element
}

func (c *Cache) Set(key string, value interface{}) {
	c.Lock()
	defer c.Unlock()
	if elem, found := c.cache[key]; found && value == elem.value {
		c.sendToHead(elem)
	}
	e := &Element{key: key, value: value}
	switch len(c.cache) {
	case 0: // empty! add head
		c.head = e
	case 1: // change element to head, move prev element to tail
		c.tail = c.head
		c.head = e
		c.tail.next = c.head
		c.head.prev = c.tail
		c.head.next = nil
		c.tail.prev = nil
	case int(c.size): // have max, so evict tail and add to head
		e.prev = c.head
		c.head.next = e
		c.head = e
		if c.tail != nil {
			delete(c.cache, c.tail.key)
			c.tail = c.tail.next
			c.tail.prev = nil
		}
	default:
		c.sendToHead(e)
	}
	c.cache[key] = e
}
