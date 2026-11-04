package probabilistic

import "hash/fnv"

// hash64 is FNV-1a of item with seed mixed in, then the SplitMix64 finalizer,
// so that every bit of the result depends on every bit of the input (FNV-1a
// alone leaves the high bits weak for short keys).
func hash64(item string, seed uint64) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(item)) // a hash.Hash never returns an error
	x := h.Sum64() ^ (seed * 0x9e3779b97f4a7c15)
	x ^= x >> 30
	x *= 0xbf58476d1ce4e5b9
	x ^= x >> 27
	x *= 0x94d049bb133111eb
	return x ^ x>>31
}
