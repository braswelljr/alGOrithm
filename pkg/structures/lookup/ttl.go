package lookup

import (
	"container/list"
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

type ttlEntry[K comparable, V any] struct {
	key     K
	value   V
	expires int64
}

// TTLCache is an LRU cache whose entries also expire: each one lives ttl
// time units after it was last written, like a DNS or session cache. Time is
// passed in (now), not read from a clock, so behavior is the same on every run
// and easy to test; pass time.Now().Unix() or a request counter.
//
// An expired entry is gone for Get, and removed when it is found; when the
// cache is full, Put first drops expired entries and then the least recently
// used one.
//
// Time: O(1) Get and Put, O(n) when Put sweeps expired entries.
// Space: O(capacity).
type TTLCache[K comparable, V any] struct {
	capacity int
	ttl      int64
	order    *list.List // front is the most recently used
	items    map[K]*list.Element
}

// NewTTLCache returns a cache of at most capacity entries (at least 1) that
// each live ttl time units (at least 1).
func NewTTLCache[K comparable, V any](capacity int, ttl int64) *TTLCache[K, V] {
	return &TTLCache[K, V]{capacity: max(capacity, 1), ttl: max(ttl, 1), order: list.New(), items: map[K]*list.Element{}}
}

// Get returns the value for key if it is there and has not expired at now,
// and marks it most recently used.
func (c *TTLCache[K, V]) Get(key K, now int64) (V, bool) {
	var zero V
	el, ok := c.items[key]
	if !ok {
		return zero, false
	}
	e := el.Value.(*ttlEntry[K, V])
	if now >= e.expires {
		c.order.Remove(el)
		delete(c.items, key)
		return zero, false
	}
	c.order.MoveToFront(el)
	return e.value, true
}

// Put stores value for key at now, so it expires at now+ttl, and returns the
// keys it evicted to make room: expired ones first, else the least recently used.
func (c *TTLCache[K, V]) Put(key K, value V, now int64) []K {
	if el, ok := c.items[key]; ok {
		e := el.Value.(*ttlEntry[K, V])
		e.value, e.expires = value, now+c.ttl
		c.order.MoveToFront(el)
		return nil
	}
	var evicted []K
	if len(c.items) >= c.capacity {
		evicted = c.Expire(now)
	}
	if len(c.items) >= c.capacity {
		el := c.order.Back()
		k := el.Value.(*ttlEntry[K, V]).key
		c.order.Remove(el)
		delete(c.items, k)
		evicted = append(evicted, k)
	}
	c.items[key] = c.order.PushFront(&ttlEntry[K, V]{key: key, value: value, expires: now + c.ttl})
	return evicted
}

// Expire removes every entry expired at now and returns their keys, oldest
// use first.
func (c *TTLCache[K, V]) Expire(now int64) []K {
	var out []K
	for el := c.order.Back(); el != nil; {
		prev := el.Prev()
		if e := el.Value.(*ttlEntry[K, V]); now >= e.expires {
			c.order.Remove(el)
			delete(c.items, e.key)
			out = append(out, e.key)
		}
		el = prev
	}
	return out
}

// Len returns the number of entries held, expired ones not yet removed included.
func (c *TTLCache[K, V]) Len() int { return len(c.items) }

// Keys returns the keys from most to least recently used.
func (c *TTLCache[K, V]) Keys() []K {
	out := make([]K, 0, len(c.items))
	for el := c.order.Front(); el != nil; el = el.Next() {
		out = append(out, el.Value.(*ttlEntry[K, V]).key)
	}
	return out
}

// Snapshot describes the entries from most to least recently used, each with
// the time it expires.
func (c *TTLCache[K, V]) Snapshot() trace.Snapshot {
	labels := make([]string, 0, len(c.items))
	for el := c.order.Front(); el != nil; el = el.Next() {
		e := el.Value.(*ttlEntry[K, V])
		labels = append(labels, fmt.Sprintf("%v=%v until %d", e.key, e.value, e.expires))
	}
	return trace.Row(labels, "next", "newest", "oldest")
}

// TTL returns how long an entry lives after it is written.
func (c *TTLCache[K, V]) TTL() int64 { return c.ttl }
