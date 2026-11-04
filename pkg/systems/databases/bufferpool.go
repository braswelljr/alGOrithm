package databases

import (
	"fmt"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Replacement policies for a BufferPool.
const (
	Clock = "clock"
	LRUK  = "lru-k"
)

type frame struct {
	page       int
	used       bool
	pins       int
	dirty, ref bool
}

// PoolStats counts a buffer pool's work.
type PoolStats struct {
	Hits, Misses, Evictions, WriteBacks int
}

// BufferPool is the part of a database that keeps disk pages in a fixed
// number of memory frames. A page is fetched into a frame on its first use;
// while a query pins it, it cannot be evicted; a dirty page is written back
// before its frame is reused. When every frame is full, the policy picks the
// frame to reuse:
//
//   - Clock (second chance): a hand sweeps the frames, clearing reference
//     bits, and takes the first unpinned frame whose bit is already clear.
//   - LRU-K (O'Neil, O'Neil and Weikum, 1993): take the unpinned page whose
//     K-th most recent use is oldest; pages used fewer than K times go first,
//     oldest last use first. With K = 2 a single scan cannot flush the pages
//     used again and again, which plain LRU lets it do.
//
// Simplified: pages are numbers and disk writes are counted, not made. The
// zero BufferPool is not ready; use NewBufferPool.
type BufferPool struct {
	policy  string
	k       int
	frames  []frame
	hand    int
	clock   int
	history map[int][]int // each page's last k use times, newest last
	Stats   PoolStats
}

// NewBufferPool returns a pool of n frames using policy (Clock or LRUK, with
// k the number of uses LRU-K remembers).
func NewBufferPool(n int, policy string, k int) (*BufferPool, error) {
	if policy != Clock && policy != LRUK {
		return nil, fmt.Errorf("%w: unknown policy %q (use %s or %s)", ErrInput, policy, Clock, LRUK)
	}
	return &BufferPool{policy: policy, k: max(1, k), frames: make([]frame, max(1, n)), history: map[int][]int{}}, nil
}

// Fetch brings page into a frame, if it is not already in one, and pins it
// when pin is set; dirty marks it changed. It reports whether the page was
// already in memory, and fails when every frame is pinned. Each frame looked
// at is reported to r as a Visit; an eviction as a Mark "evict". Time: O(n)
// for n frames.
func (b *BufferPool) Fetch(page int, pin, dirty bool, r trace.Recorder) (bool, error) {
	b.clock++
	b.history[page] = append(b.history[page], b.clock)
	if h := b.history[page]; len(h) > b.k {
		b.history[page] = h[len(h)-b.k:]
	}
	i := b.find(page)
	hit := i >= 0
	if hit {
		b.Stats.Hits++
	} else {
		b.Stats.Misses++
		var err error
		if i, err = b.victim(r); err != nil {
			return false, err
		}
		if f := b.frames[i]; f.used {
			b.Stats.Evictions++
			if f.dirty {
				b.Stats.WriteBacks++
			}
			r.Record(trace.Event{Kind: trace.Mark, Indices: []int{i}, Values: []int{f.page}, Note: "evict"})
		}
		b.frames[i] = frame{page: page, used: true}
	}
	f := &b.frames[i]
	f.ref = true
	if pin {
		f.pins++
	}
	if dirty {
		f.dirty = true
	}
	return hit, nil
}

// Unpin releases one pin on page and reports whether it was pinned.
func (b *BufferPool) Unpin(page int) bool {
	i := b.find(page)
	if i < 0 || b.frames[i].pins == 0 {
		return false
	}
	b.frames[i].pins--
	return true
}

func (b *BufferPool) find(page int) int {
	for i, f := range b.frames {
		if f.used && f.page == page {
			return i
		}
	}
	return -1
}

func (b *BufferPool) victim(r trace.Recorder) (int, error) {
	for i, f := range b.frames {
		if !f.used {
			return i, nil
		}
	}
	if b.policy == Clock {
		for n := 0; n <= 2*len(b.frames); n++ {
			i := b.hand
			b.hand = (b.hand + 1) % len(b.frames)
			r.Record(trace.Event{Kind: trace.Visit, Indices: []int{i}})
			f := &b.frames[i]
			if f.pins > 0 {
				continue
			}
			if f.ref {
				f.ref = false
				continue
			}
			return i, nil
		}
		return 0, fmt.Errorf("%w: every frame is pinned", ErrInput)
	}
	best := -1
	for i, f := range b.frames {
		r.Record(trace.Event{Kind: trace.Visit, Indices: []int{i}})
		if f.pins > 0 {
			continue
		}
		if best < 0 || b.older(f.page, b.frames[best].page) {
			best = i
		}
	}
	if best < 0 {
		return 0, fmt.Errorf("%w: every frame is pinned", ErrInput)
	}
	return best, nil
}

// older reports whether page a is a better LRU-K victim than page b: fewer
// than K uses beats K uses, then the older K-th (or, below K, last) use.
func (b *BufferPool) older(a, c int) bool {
	ha, hc := b.history[a], b.history[c]
	fa, fc := len(ha) < b.k, len(hc) < b.k
	if fa != fc {
		return fa
	}
	if fa {
		return ha[len(ha)-1] < hc[len(hc)-1]
	}
	return ha[0] < hc[0]
}

// Snapshot draws the frames in order, with each page's pins, reference bit
// and dirty flag, and the clock hand.
func (b *BufferPool) Snapshot() trace.Snapshot {
	s := trace.Snapshot{Layout: "row", Nodes: []trace.Node{}}
	for i, f := range b.frames {
		label := "empty"
		if f.used {
			var flags []string
			if f.pins > 0 {
				flags = append(flags, fmt.Sprintf("pinned %d", f.pins))
			}
			if b.policy == Clock && f.ref {
				flags = append(flags, "ref")
			}
			if f.dirty {
				flags = append(flags, "dirty")
			}
			label = fmt.Sprintf("page %d", f.page)
			if len(flags) > 0 {
				label += " (" + strings.Join(flags, ", ") + ")"
			}
		}
		nd := trace.Node{ID: i, Label: label, Parent: -1}
		if b.policy == Clock && i == b.hand {
			nd.Role = "hand"
		}
		s.Nodes = append(s.Nodes, nd)
	}
	return s
}
