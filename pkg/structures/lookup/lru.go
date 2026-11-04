package lookup

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

type lruEntry[K comparable, V any] struct {
	key        K
	value      V
	prev, next *lruEntry[K, V]
}

// LRU is a least-recently-used cache: at most Capacity entries, and when a new
// one does not fit the one used longest ago is evicted. A map finds an entry
// and a doubly linked list keeps the order of use, so every operation is O(1).
// The zero LRU is not ready; use NewLRU.
type LRU[K comparable, V any] struct {
	capacity   int
	items      map[K]*lruEntry[K, V]
	head, tail *lruEntry[K, V] // head is the most recently used
}

// NewLRU returns an empty cache holding at most capacity entries (at least 1).
func NewLRU[K comparable, V any](capacity int) *LRU[K, V] {
	return &LRU[K, V]{capacity: max(1, capacity), items: map[K]*lruEntry[K, V]{}}
}

// Len is the number of entries; Capacity the most it holds. Time: O(1).
func (c *LRU[K, V]) Len() int { return len(c.items) }

// Capacity is the most entries the cache holds.
func (c *LRU[K, V]) Capacity() int { return c.capacity }

func (c *LRU[K, V]) unlink(e *lruEntry[K, V]) {
	if e.prev != nil {
		e.prev.next = e.next
	} else {
		c.head = e.next
	}
	if e.next != nil {
		e.next.prev = e.prev
	} else {
		c.tail = e.prev
	}
	e.prev, e.next = nil, nil
}

func (c *LRU[K, V]) front(e *lruEntry[K, V]) {
	e.next = c.head
	if c.head != nil {
		c.head.prev = e
	}
	c.head = e
	if c.tail == nil {
		c.tail = e
	}
}

// Get returns key's value and makes it the most recently used; ok is false
// when key is not cached. Time: O(1).
func (c *LRU[K, V]) Get(key K) (value V, ok bool) {
	e, ok := c.items[key]
	if !ok {
		return value, false
	}
	c.unlink(e)
	c.front(e)
	return e.value, true
}

// Put sets key to value as the most recently used entry. When the cache was
// full it evicts the least recently used entry and returns its key with
// evicted true. Time: O(1).
func (c *LRU[K, V]) Put(key K, value V) (evictedKey K, evicted bool) {
	if e, ok := c.items[key]; ok {
		e.value = value
		c.unlink(e)
		c.front(e)
		return evictedKey, false
	}
	if len(c.items) == c.capacity {
		old := c.tail
		c.unlink(old)
		delete(c.items, old.key)
		evictedKey, evicted = old.key, true
	}
	e := &lruEntry[K, V]{key: key, value: value}
	c.items[key] = e
	c.front(e)
	return evictedKey, evicted
}

// Keys returns the keys from most to least recently used. Time: O(n).
func (c *LRU[K, V]) Keys() []K {
	out := make([]K, 0, len(c.items))
	for e := c.head; e != nil; e = e.next {
		out = append(out, e.key)
	}
	return out
}

// Snapshot describes the cache newest to oldest as "key: value" nodes.
func (c *LRU[K, V]) Snapshot() trace.Snapshot {
	var labels []string
	for e := c.head; e != nil; e = e.next {
		labels = append(labels, fmt.Sprintf("%v: %v", e.key, e.value))
	}
	return trace.Row(labels, "next", "newest", "oldest")
}
