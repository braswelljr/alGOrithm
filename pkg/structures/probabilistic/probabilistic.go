// Package probabilistic holds structures that answer approximately in a
// fraction of the memory an exact answer needs, and say how wrong they can be:
// BloomFilter and CuckooFilter (membership, false positives only),
// CountMinSketch (counts, overestimates only) and HyperLogLog (distinct
// counts, within a few percent).
package probabilistic
