package hashing

// DistributionResult is how a hash spread keys over buckets.
type DistributionResult struct {
	Counts    []int   // keys per bucket
	Empty     int     // buckets with no key
	Max       int     // the fullest bucket
	ChiSquare float64 // sum((count - expected)^2 / expected); about Buckets-1 for a uniform hash
}

// Distribution hashes every key with hash and counts how many land in each of
// the buckets (hash mod buckets). A good hash gives counts near
// len(keys)/buckets and a chi-square near buckets-1.
//
// Time: O(total key length + buckets). Space: O(buckets).
func Distribution(keys []string, buckets int, hash func([]byte) uint32) DistributionResult {
	if buckets < 1 {
		buckets = 1
	}
	r := DistributionResult{Counts: make([]int, buckets)}
	for _, k := range keys {
		r.Counts[hash([]byte(k))%uint32(buckets)]++
	}
	expected := float64(len(keys)) / float64(buckets)
	for _, c := range r.Counts {
		if c == 0 {
			r.Empty++
		}
		r.Max = max(r.Max, c)
		if expected > 0 {
			d := float64(c) - expected
			r.ChiSquare += d * d / expected
		}
	}
	return r
}
