package probabilistic

import (
	"fmt"
	"math/rand"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// bucketSize is how many fingerprints a cuckoo filter bucket holds.
const bucketSize = 4

// maxKicks is how many fingerprints Add moves before it declares the filter full.
const maxKicks = 500

// CuckooFilter answers "have I seen this?" like a BloomFilter, false
// positives only, but also supports Remove (Bin Fan et al., 2014). It stores
// an 8-bit fingerprint of each item in one of two buckets: i1 = hash(item) and
// i2 = i1 xor hash(fingerprint), so either bucket can be found from the other
// and the fingerprint alone. When both are full, Add evicts a fingerprint to
// its other bucket, cuckoo style, and so on, up to maxKicks moves.
//
// The false-positive rate is about 2 * bucketSize / 256 when full. Removing
// an item that was never added can remove another item's fingerprint, so only
// remove what you added.
//
// Time: O(1) Contains and Remove, O(1) amortized Add.
// Space: O(buckets) bytes.
type CuckooFilter struct {
	buckets [][bucketSize]uint8 // 0 is an empty slot
	n       int
	rng     *rand.Rand
}

// NewCuckooFilter returns a filter of at least buckets buckets, rounded up to
// a power of two so the xor stays in range; seed picks the evictions.
func NewCuckooFilter(buckets int, seed int64) *CuckooFilter {
	size := 1
	for size < buckets {
		size <<= 1
	}
	return &CuckooFilter{buckets: make([][bucketSize]uint8, size), rng: rand.New(rand.NewSource(seed))}
}

func (f *CuckooFilter) mask() uint64 { return uint64(len(f.buckets) - 1) }

// Locate returns item's fingerprint and its two buckets.
func (f *CuckooFilter) Locate(item string) (fp uint8, i1, i2 int) {
	h := hash64(item, 7)
	fp = uint8(h>>56) | 1 // never 0, which marks an empty slot
	i1 = int(h & f.mask())
	return fp, i1, f.alt(i1, fp)
}

func (f *CuckooFilter) alt(i int, fp uint8) int {
	return int((uint64(i) ^ hash64(string([]byte{fp}), 9)) & f.mask())
}

func (f *CuckooFilter) put(i int, fp uint8) bool {
	for s := range f.buckets[i] {
		if f.buckets[i][s] == 0 {
			f.buckets[i][s] = fp
			return true
		}
	}
	return false
}

// Add stores item and reports whether there was room; on false the filter
// is too full and item is not stored (one evicted fingerprint may have been
// moved, never lost).
func (f *CuckooFilter) Add(item string) bool {
	fp, i1, i2 := f.Locate(item)
	if f.put(i1, fp) || f.put(i2, fp) {
		f.n++
		return true
	}
	i := i1
	if f.rng.Intn(2) == 1 {
		i = i2
	}
	type move struct {
		bucket, slot int
		fp           uint8
	}
	var moves []move
	for k := 0; k < maxKicks; k++ {
		s := f.rng.Intn(bucketSize)
		moves = append(moves, move{i, s, f.buckets[i][s]})
		fp, f.buckets[i][s] = f.buckets[i][s], fp
		i = f.alt(i, fp)
		if f.put(i, fp) {
			f.n++
			return true
		}
	}
	for k := len(moves) - 1; k >= 0; k-- { // undo, so no stored item is lost
		m := moves[k]
		f.buckets[m.bucket][m.slot] = m.fp
	}
	return false
}

// Contains reports whether item might have been added; false is certain.
func (f *CuckooFilter) Contains(item string) bool {
	fp, i1, i2 := f.Locate(item)
	for s := 0; s < bucketSize; s++ {
		if f.buckets[i1][s] == fp || f.buckets[i2][s] == fp {
			return true
		}
	}
	return false
}

// Remove deletes one copy of item's fingerprint and reports whether it found one.
func (f *CuckooFilter) Remove(item string) bool {
	fp, i1, i2 := f.Locate(item)
	for _, i := range []int{i1, i2} {
		for s := range f.buckets[i] {
			if f.buckets[i][s] == fp {
				f.buckets[i][s] = 0
				f.n--
				return true
			}
		}
	}
	return false
}

// Len returns the number of fingerprints stored.
func (f *CuckooFilter) Len() int { return f.n }

// LoadFactor returns the fraction of slots in use.
func (f *CuckooFilter) LoadFactor() float64 {
	return float64(f.n) / float64(len(f.buckets)*bucketSize)
}

// Snapshot describes the buckets, each with its fingerprints in hex (.. empty).
func (f *CuckooFilter) Snapshot() trace.Snapshot {
	labels := make([]string, len(f.buckets))
	for i, b := range f.buckets {
		l := fmt.Sprintf("%d:", i)
		for _, fp := range b {
			if fp == 0 {
				l += " .."
			} else {
				l += fmt.Sprintf(" %02x", fp)
			}
		}
		labels[i] = l
	}
	return trace.Row(labels, "", "", "")
}
