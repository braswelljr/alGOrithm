package lookup

import (
	"fmt"
	"hash/fnv"
	"strconv"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// HashTableMaxLoad is the average chain length that makes a HashTable double
// its buckets.
const HashTableMaxLoad = 0.75

type hashEntry[K comparable, V any] struct {
	key   K
	value V
}

// HashTable maps keys to values with separate chaining: each key hashes to one
// bucket, and keys that collide share the bucket's chain. When the table gets
// too full it doubles its buckets and rehashes every key. Keys hash by their
// fmt.Sprint form (FNV-1a), which is simple to follow; Go's map is the fast
// version of the same idea. The zero HashTable is not ready; use NewHashTable.
type HashTable[K comparable, V any] struct {
	buckets [][]hashEntry[K, V]
	size    int
}

// NewHashTable returns an empty table with the given number of buckets (at least 1).
func NewHashTable[K comparable, V any](buckets int) *HashTable[K, V] {
	return &HashTable[K, V]{buckets: make([][]hashEntry[K, V], max(1, buckets))}
}

// Bucket returns the bucket key hashes to.
func (t *HashTable[K, V]) Bucket(key K) int {
	h := fnv.New64a()
	fmt.Fprint(h, key)
	return int(h.Sum64() % uint64(len(t.buckets)))
}

// Len is the number of keys. Time: O(1).
func (t *HashTable[K, V]) Len() int { return t.size }

// Buckets is the number of buckets. Time: O(1).
func (t *HashTable[K, V]) Buckets() int { return len(t.buckets) }

// Put sets key to value and reports whether key was new.
// Time: O(1) on average, O(n) when every key collides.
func (t *HashTable[K, V]) Put(key K, value V) bool {
	b := t.Bucket(key)
	for i := range t.buckets[b] {
		if t.buckets[b][i].key == key {
			t.buckets[b][i].value = value
			return false
		}
	}
	t.buckets[b] = append(t.buckets[b], hashEntry[K, V]{key, value})
	t.size++
	if float64(t.size) > HashTableMaxLoad*float64(len(t.buckets)) {
		t.resize(2 * len(t.buckets))
	}
	return true
}

// Get returns key's value; ok is false when key is absent. Time: O(1) average.
func (t *HashTable[K, V]) Get(key K) (value V, ok bool) {
	for _, e := range t.buckets[t.Bucket(key)] {
		if e.key == key {
			return e.value, true
		}
	}
	return value, false
}

// Remove deletes key and reports whether it was there. Time: O(1) average.
func (t *HashTable[K, V]) Remove(key K) bool {
	b := t.Bucket(key)
	for i, e := range t.buckets[b] {
		if e.key == key {
			t.buckets[b] = append(t.buckets[b][:i], t.buckets[b][i+1:]...)
			t.size--
			return true
		}
	}
	return false
}

func (t *HashTable[K, V]) resize(n int) {
	old := t.buckets
	t.buckets, t.size = make([][]hashEntry[K, V], n), 0
	for _, b := range old {
		for _, e := range b {
			t.Put(e.key, e.value)
		}
	}
}

// Snapshot describes the table as a tree: the table, its buckets ("#0", "#1",
// ...), and under each bucket its chain of "key: value" entries.
func (t *HashTable[K, V]) Snapshot() trace.Snapshot {
	s := trace.Snapshot{Layout: "tree", Nodes: []trace.Node{{ID: 0, Label: "table", Role: "root", Parent: -1}}}
	for b, chain := range t.buckets {
		bid := len(s.Nodes)
		s.Nodes = append(s.Nodes, trace.Node{ID: bid, Label: "#" + strconv.Itoa(b), Role: "bucket", Parent: 0})
		s.Edges = append(s.Edges, trace.Edge{From: 0, To: bid})
		parent := bid
		for _, e := range chain {
			id := len(s.Nodes)
			s.Nodes = append(s.Nodes, trace.Node{ID: id, Label: fmt.Sprintf("%v: %v", e.key, e.value), Parent: parent})
			s.Edges = append(s.Edges, trace.Edge{From: parent, To: id, Label: "next"})
			parent = id
		}
	}
	return s
}
