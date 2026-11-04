// Package cache simulates a CPU cache on a trace of memory addresses: how an
// address splits into tag, set index and block offset, whether each access
// hits or misses, which line is evicted, and why each miss happened. One
// simulator covers the three organizations, set by Ways: direct-mapped (1
// way), n-way set-associative, and fully associative (one set holding every
// line), with LRU or FIFO replacement. AMAT computes the average memory access
// time from hit time, miss rate and miss penalty.
//
//	r, _ := cache.Simulate(cache.Config{Lines: 4, BlockSize: 4, Ways: 1}, []int{0, 4, 16, 0})
//	fmt.Println(r.Hits, r.Misses)
package cache

import (
	"errors"
	"fmt"
	"math/bits"
)

// Policy is how a full set chooses the line to evict.
type Policy string

// The replacement policies.
const (
	LRU  Policy = "lru"  // the line used longest ago
	FIFO Policy = "fifo" // the line loaded longest ago
)

// Config describes a cache: Lines lines of BlockSize bytes, grouped into sets
// of Ways lines. Lines, BlockSize and Lines/Ways must be powers of two, and
// Ways must divide Lines. Ways 1 is direct-mapped; Ways == Lines is fully
// associative.
type Config struct {
	Lines     int
	BlockSize int
	Ways      int
	Policy    Policy
}

// Sets is the number of sets.
func (c Config) Sets() int { return c.Lines / c.Ways }

// ErrConfig is returned for a configuration that is not a valid cache, or a
// trace that is empty, too long or negative.
var ErrConfig = errors.New("cache: invalid configuration")

// MaxAccesses is the longest trace Simulate takes.
const MaxAccesses = 10_000

func pow2(n int) bool { return n > 0 && n&(n-1) == 0 }

func (c Config) check() error {
	switch {
	case !pow2(c.Lines) || c.Lines > 1<<16, !pow2(c.BlockSize) || c.BlockSize > 1<<16:
		return fmt.Errorf("%w: lines and block size must be powers of two up to 65536", ErrConfig)
	case c.Ways < 1 || c.Ways > c.Lines || c.Lines%c.Ways != 0 || !pow2(c.Lines/c.Ways):
		return fmt.Errorf("%w: ways must divide the lines into a power-of-two number of sets", ErrConfig)
	case c.Policy != "" && c.Policy != LRU && c.Policy != FIFO:
		return fmt.Errorf("%w: policy %q (lru or fifo)", ErrConfig, c.Policy)
	}
	return nil
}

// Split is an address cut into its parts.
type Split struct {
	Tag, Set, Offset int
}

// Split cuts addr into tag, set index and block offset: the low log2(block
// size) bits are the offset, the next log2(sets) bits the set, the rest the
// tag.
func (c Config) Split(addr int) Split {
	ob := bits.TrailingZeros(uint(c.BlockSize))
	sb := bits.TrailingZeros(uint(c.Sets()))
	return Split{Tag: addr >> (ob + sb), Set: (addr >> ob) & (c.Sets() - 1), Offset: addr & (c.BlockSize - 1)}
}

// Miss kinds, the "three Cs".
const (
	Hit        = "hit"
	Compulsory = "compulsory" // the block was never in the cache
	Capacity   = "capacity"   // a fully associative cache of the same size would miss too
	Conflict   = "conflict"   // only the mapping caused it: a fully associative cache would hit
)

// Access is the result of one address.
type Access struct {
	Addr int `json:"addr"`
	Split
	Result string `json:"result"` // Hit, Compulsory, Capacity or Conflict
	Way    int    `json:"way"`    // the line within the set it is now in
	// Evicted is the tag replaced, or -1.
	Evicted int `json:"evicted"`
}

// Result is a whole run.
type Result struct {
	Accesses []Access `json:"accesses"`
	Hits     int      `json:"hits"`
	Misses   int      `json:"misses"`
	// ByKind counts the misses by kind.
	ByKind map[string]int `json:"byKind"`
	// Final is each set's tags after the last access, way by way; -1 is empty.
	Final [][]int `json:"final"`
}

// HitRate is Hits / accesses.
func (r Result) HitRate() float64 {
	if len(r.Accesses) == 0 {
		return 0
	}
	return float64(r.Hits) / float64(len(r.Accesses))
}

type line struct {
	tag, used, loaded int
	valid             bool
}

// Simulate runs the cache on trace (byte addresses), starting empty.
//
// Time: O(n * ways). Space: O(lines + n).
func Simulate(c Config, trace []int) (Result, error) {
	if c.Policy == "" {
		c.Policy = LRU
	}
	if err := c.check(); err != nil {
		return Result{}, err
	}
	if len(trace) == 0 || len(trace) > MaxAccesses {
		return Result{}, fmt.Errorf("%w: give 1 to %d addresses", ErrConfig, MaxAccesses)
	}
	sets := make([][]line, c.Sets())
	for i := range sets {
		sets[i] = make([]line, c.Ways)
	}
	// A fully associative LRU cache of the same size, to tell capacity from
	// conflict misses.
	full := map[int]int{} // block -> last use
	seen := map[int]bool{}
	r := Result{ByKind: map[string]int{}}
	for t, addr := range trace {
		if addr < 0 {
			return Result{}, fmt.Errorf("%w: address %d is negative", ErrConfig, addr)
		}
		sp := c.Split(addr)
		block := addr / c.BlockSize
		a := Access{Addr: addr, Split: sp, Evicted: -1, Way: -1}
		set := sets[sp.Set]
		for w := range set {
			if set[w].valid && set[w].tag == sp.Tag {
				a.Way = w
			}
		}
		_, inFull := full[block]
		switch {
		case a.Way >= 0:
			a.Result = Hit
			set[a.Way].used = t
			r.Hits++
		default:
			r.Misses++
			switch {
			case !seen[block]:
				a.Result = Compulsory
			case inFull:
				a.Result = Conflict
			default:
				a.Result = Capacity
			}
			r.ByKind[a.Result]++
			a.Way = victim(set, c.Policy)
			if set[a.Way].valid {
				a.Evicted = set[a.Way].tag
			}
			set[a.Way] = line{tag: sp.Tag, used: t, loaded: t, valid: true}
		}
		seen[block] = true
		full[block] = t
		if len(full) > c.Lines { // evict the LRU block of the fully associative shadow
			oldest, ob := -1, 0
			for b, u := range full {
				if oldest < 0 || u < oldest {
					oldest, ob = u, b
				}
			}
			delete(full, ob)
		}
		r.Accesses = append(r.Accesses, a)
	}
	for _, set := range sets {
		row := make([]int, c.Ways)
		for w, l := range set {
			row[w] = -1
			if l.valid {
				row[w] = l.tag
			}
		}
		r.Final = append(r.Final, row)
	}
	return r, nil
}

// victim returns the way to fill: an empty one, else the policy's choice.
func victim(set []line, p Policy) int {
	best := 0
	for w, l := range set {
		if !l.valid {
			return w
		}
		key, bestKey := l.used, set[best].used
		if p == FIFO {
			key, bestKey = l.loaded, set[best].loaded
		}
		if key < bestKey {
			best = w
		}
	}
	return best
}

// AMAT is the average memory access time: hitTime + missRate * missPenalty,
// in whatever unit the times are given (cycles, nanoseconds). For more levels,
// pass the next level's AMAT as the miss penalty.
func AMAT(hitTime, missRate, missPenalty float64) float64 {
	return hitTime + missRate*missPenalty
}
