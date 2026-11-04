package lookup

import (
	"container/list"
	"fmt"
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

type lfuEntry[K comparable, V any] struct {
	key   K
	value V
	freq  int
	elem  *list.Element
}

// LFUCache is a least-frequently-used cache: at most Capacity entries, and
// when a new one does not fit, the one used the fewest times is evicted (the
// least recently used of those, on a tie). One list per use count and a
// pointer to the smallest count make every operation O(1). The zero LFUCache
// is not ready; use NewLFUCache.
type LFUCache[K comparable, V any] struct {
	capacity int
	items    map[K]*lfuEntry[K, V]
	freqs    map[int]*list.List // count -> entries, most recently used at the front
	minFreq  int
}

// NewLFUCache returns an empty cache of at most capacity entries (at least 1).
func NewLFUCache[K comparable, V any](capacity int) *LFUCache[K, V] {
	return &LFUCache[K, V]{capacity: max(1, capacity), items: map[K]*lfuEntry[K, V]{}, freqs: map[int]*list.List{}}
}

// Len is the number of entries. Time: O(1).
func (c *LFUCache[K, V]) Len() int { return len(c.items) }

// Capacity is the most entries the cache holds.
func (c *LFUCache[K, V]) Capacity() int { return c.capacity }

func (c *LFUCache[K, V]) push(e *lfuEntry[K, V]) {
	l, ok := c.freqs[e.freq]
	if !ok {
		l = list.New()
		c.freqs[e.freq] = l
	}
	e.elem = l.PushFront(e)
}

// touch moves e up one use count.
func (c *LFUCache[K, V]) touch(e *lfuEntry[K, V]) {
	l := c.freqs[e.freq]
	l.Remove(e.elem)
	if l.Len() == 0 {
		delete(c.freqs, e.freq)
		if c.minFreq == e.freq {
			c.minFreq++
		}
	}
	e.freq++
	c.push(e)
}

// Get returns key's value and counts a use; ok is false when key is not
// cached. Time: O(1).
func (c *LFUCache[K, V]) Get(key K) (value V, ok bool) {
	e, ok := c.items[key]
	if !ok {
		return value, false
	}
	c.touch(e)
	return e.value, true
}

// Put sets key to value and counts a use. When the cache was full it evicts
// the least frequently used entry and returns its key with evicted true.
// Time: O(1).
func (c *LFUCache[K, V]) Put(key K, value V) (evictedKey K, evicted bool) {
	if e, ok := c.items[key]; ok {
		e.value = value
		c.touch(e)
		return evictedKey, false
	}
	if len(c.items) == c.capacity {
		l := c.freqs[c.minFreq]
		old := l.Remove(l.Back()).(*lfuEntry[K, V])
		if l.Len() == 0 {
			delete(c.freqs, c.minFreq)
		}
		delete(c.items, old.key)
		evictedKey, evicted = old.key, true
	}
	e := &lfuEntry[K, V]{key: key, value: value, freq: 1}
	c.items[key] = e
	c.push(e)
	c.minFreq = 1
	return evictedKey, evicted
}

// Frequency returns how many times key has been used; ok is false when it is
// not cached.
func (c *LFUCache[K, V]) Frequency(key K) (int, bool) {
	e, ok := c.items[key]
	if !ok {
		return 0, false
	}
	return e.freq, true
}

// entries returns the entries most used first, most recent first on a tie.
func (c *LFUCache[K, V]) entries() []*lfuEntry[K, V] {
	counts := make([]int, 0, len(c.freqs))
	for f := range c.freqs {
		counts = append(counts, f)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(counts)))
	var out []*lfuEntry[K, V]
	for _, f := range counts {
		for el := c.freqs[f].Front(); el != nil; el = el.Next() {
			out = append(out, el.Value.(*lfuEntry[K, V]))
		}
	}
	return out
}

// Keys returns the keys most used first. Time: O(n log n).
func (c *LFUCache[K, V]) Keys() []K {
	var out []K
	for _, e := range c.entries() {
		out = append(out, e.key)
	}
	return out
}

// Snapshot describes the entries most used first, labeled "key: value (freq
// f)"; the last is the next to be evicted.
func (c *LFUCache[K, V]) Snapshot() trace.Snapshot {
	var labels []string
	for _, e := range c.entries() {
		labels = append(labels, fmt.Sprintf("%v: %v (freq %d)", e.key, e.value, e.freq))
	}
	return trace.Row(labels, "", "most used", "next out")
}
