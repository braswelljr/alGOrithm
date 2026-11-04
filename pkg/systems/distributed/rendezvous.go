package distributed

import (
	"fmt"
	"hash/fnv"
	"strconv"
)

// Owner returns the node that owns key under rendezvous hashing (highest
// random weight, David Thaler and Chinya Ravishankar, 1996): every node
// scores the key with a hash of the pair, and the highest score wins. Every
// client computes the same owner with no shared ring, and when a node joins
// it takes only the keys where it scores highest: about 1/(n+1) of them.
//
// Time: O(nodes) per key. Space: O(1).
func Owner(key string, nodes []string) string {
	best, owner := uint64(0), ""
	for _, node := range nodes {
		h := fnv.New64a()
		_, _ = h.Write([]byte(node))
		_, _ = h.Write([]byte{0})
		_, _ = h.Write([]byte(key))
		score := mix(h.Sum64())
		if owner == "" || score > best {
			best, owner = score, node
		}
	}
	return owner
}

// mix spreads FNV's bits (the finalizer of SplitMix64), so similar keys get
// unrelated scores.
func mix(x uint64) uint64 {
	x ^= x >> 30
	x *= 0xbf58476d1ce4e5b9
	x ^= x >> 27
	x *= 0x94d049bb133111eb
	return x ^ x>>31
}

// Rebalance is how many keys change owner when a node joins.
type Rebalance struct {
	Keys       int `json:"keys"`
	Rendezvous int `json:"rendezvous"` // moved under rendezvous hashing
	Modulo     int `json:"modulo"`     // moved under hash(key) mod nodes
	// Load is the keys each node owns under rendezvous hashing after the
	// join, in node order (the new node last).
	Load []int `json:"load"`
}

// Join counts the keys key0..key(keys-1) that change owner when one node
// joins n nodes, under rendezvous hashing and under the naive hash mod n,
// which reshuffles almost every key (all but about 1/(n+1)).
//
// Time: O(keys * nodes). Space: O(nodes).
func Join(n, keys int) (Rebalance, error) {
	if n < 1 || n > 100 || keys < 1 || keys > 100_000 {
		return Rebalance{}, fmt.Errorf("%w: 1 to 100 nodes and 1 to 100000 keys", ErrInput)
	}
	before := make([]string, n)
	for i := range before {
		before[i] = "node-" + strconv.Itoa(i)
	}
	after := append(append([]string(nil), before...), "node-"+strconv.Itoa(n))
	res := Rebalance{Keys: keys, Load: make([]int, n+1)}
	for k := 0; k < keys; k++ {
		key := "key" + strconv.Itoa(k)
		o := Owner(key, after)
		if Owner(key, before) != o {
			res.Rendezvous++
		}
		idx, _ := strconv.Atoi(o[len("node-"):])
		res.Load[idx]++
		h := fnv.New64a()
		_, _ = h.Write([]byte(key))
		x := mix(h.Sum64())
		if x%uint64(n) != x%uint64(n+1) {
			res.Modulo++
		}
	}
	return res, nil
}
