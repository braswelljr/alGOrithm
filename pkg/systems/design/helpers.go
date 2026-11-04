package design

import (
	"container/list"
	"hash/fnv"
	"math/rand"
)

// maxFrames is how many frames a case study records for the visualizer: the
// first requests in detail, then the totals in the last frame.
const maxFrames = 14

// recording reports whether the architecture still records frames.
func (a *Architecture) recording() bool { return len(a.Frames) < maxFrames-1 }

// popular draws keys 0 to n-1 with a few hot ones, as real traffic does
// (a Zipf distribution with exponent 1.1).
func popular(rng *rand.Rand, n int) func() int {
	if n <= 1 {
		return func() int { return 0 }
	}
	z := rand.NewZipf(rng, 1.1, 1, uint64(n-1))
	return func() int { return int(z.Uint64()) }
}

// lru is a least-recently-used cache of int keys.
type lru struct {
	cap   int
	order *list.List
	at    map[int]*list.Element
}

func newLRU(capacity int) *lru {
	return &lru{cap: capacity, order: list.New(), at: map[int]*list.Element{}}
}

// get reports whether k is cached, marking it most recently used.
func (c *lru) get(k int) bool {
	if e, ok := c.at[k]; ok {
		c.order.MoveToFront(e)
		return true
	}
	return false
}

// put caches k, evicting the least recently used key when full.
func (c *lru) put(k int) {
	if c.cap <= 0 || c.get(k) {
		return
	}
	if c.order.Len() == c.cap {
		old := c.order.Back()
		c.order.Remove(old)
		delete(c.at, old.Value.(int))
	}
	c.at[k] = c.order.PushFront(k)
}

// hash64 is FNV-1a of s.
func hash64(s string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}
