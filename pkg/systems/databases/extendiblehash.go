package databases

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// maxDepth bounds extendible hashing's directory at 2^maxDepth entries, so
// keys that share all their low bits fail instead of doubling forever.
const maxDepth = 10

// lowBits is the hash both hash indexes use: the key's own low bits, so a
// reader can follow every step by hand.
func lowBits(key, bits int) int { return int(uint(key) & (1<<uint(bits) - 1)) }

type xBucket struct {
	depth int
	keys  []int
}

// ExtendibleHash is extendible hashing (Fagin and others, 1979): a directory
// of 2^g pointers, g the global depth, indexed by a key's low g bits, to
// buckets of at most c keys. Several entries can share a bucket, which then
// has a local depth below g. A full bucket splits in two on one more bit;
// only when its local depth already equals g does the directory double. A
// lookup is one directory read and one bucket read, whatever the size.
//
// Simplified: the hash is the key itself. The zero ExtendibleHash is not
// ready; use NewExtendibleHash.
type ExtendibleHash struct {
	capacity int
	global   int
	dir      []*xBucket
}

// NewExtendibleHash returns an empty index with buckets of capacity keys.
func NewExtendibleHash(capacity int) *ExtendibleHash {
	return &ExtendibleHash{capacity: max(1, capacity), dir: []*xBucket{{}}}
}

// GlobalDepth is the number of key bits the directory reads.
func (h *ExtendibleHash) GlobalDepth() int { return h.global }

// Search reports whether key is in the index, reporting the directory entry
// read to r as a Visit. Time: O(c).
func (h *ExtendibleHash) Search(key int, r trace.Recorder) bool {
	i := lowBits(key, h.global)
	r.Record(trace.Event{Kind: trace.Visit, Indices: []int{i}})
	return slices.Contains(h.dir[i].keys, key)
}

// Insert adds key and reports whether it is new, splitting the bucket (and
// doubling the directory) until there is room. Time: O(c) usually; a split
// moves O(c) keys and a doubling copies O(2^g) pointers.
func (h *ExtendibleHash) Insert(key int) (bool, error) {
	for {
		b := h.dir[lowBits(key, h.global)]
		if slices.Contains(b.keys, key) {
			return false, nil
		}
		if len(b.keys) < h.capacity {
			b.keys = append(b.keys, key)
			return true, nil
		}
		if b.depth == h.global {
			if h.global == maxDepth {
				return false, fmt.Errorf("%w: more than %d keys share their low %d bits", ErrInput, h.capacity, maxDepth)
			}
			h.dir = append(h.dir, h.dir...)
			h.global++
		}
		h.split(b)
	}
}

func (h *ExtendibleHash) split(b *xBucket) {
	bit := 1 << uint(b.depth)
	b.depth++
	nb := &xBucket{depth: b.depth}
	keys := b.keys
	b.keys = nil
	for _, k := range keys {
		if int(uint(k))&bit != 0 {
			nb.keys = append(nb.keys, k)
		} else {
			b.keys = append(b.keys, k)
		}
	}
	for i := range h.dir {
		if h.dir[i] == b && i&bit != 0 {
			h.dir[i] = nb
		}
	}
}

// Delete removes key and reports whether it was there. Simplified: buckets
// never merge and the directory never shrinks. Time: O(c).
func (h *ExtendibleHash) Delete(key int) bool {
	b := h.dir[lowBits(key, h.global)]
	i := slices.Index(b.keys, key)
	if i < 0 {
		return false
	}
	b.keys = slices.Delete(b.keys, i, i+1)
	return true
}

// Snapshot draws the directory entries, each with its bits, and the buckets
// they point to, each with its local depth and keys.
func (h *ExtendibleHash) Snapshot() trace.Snapshot {
	s := trace.Snapshot{Layout: "row", Nodes: []trace.Node{}}
	bucketID := map[*xBucket]int{}
	var order []*xBucket
	for _, b := range h.dir {
		if _, ok := bucketID[b]; !ok {
			bucketID[b] = -1
			order = append(order, b)
		}
	}
	for i := range h.dir {
		label := "dir"
		if h.global > 0 {
			label = fmt.Sprintf("dir %0*b", h.global, i)
		}
		nd := trace.Node{ID: i, Label: label, Parent: -1}
		if i == 0 {
			nd.Role = fmt.Sprintf("global depth %d", h.global)
		}
		s.Nodes = append(s.Nodes, nd)
	}
	for _, b := range order {
		id := len(s.Nodes)
		bucketID[b] = id
		s.Nodes = append(s.Nodes, trace.Node{ID: id, Label: fmt.Sprintf("d=%d: %s", b.depth, joinKeys(b.keys)), Parent: -1})
	}
	for i, b := range h.dir {
		s.Edges = append(s.Edges, trace.Edge{From: i, To: bucketID[b]})
	}
	return s
}

func joinKeys(keys []int) string {
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = strconv.Itoa(k)
	}
	if len(parts) == 0 {
		return "(empty)"
	}
	return strings.Join(parts, " ")
}
