package distributed

import (
	"fmt"
	"math/rand"
)

// QuorumResult is what a run of quorum reads and writes saw.
type QuorumResult struct {
	Trials int `json:"trials"`
	// Stale counts reads that returned an older version than the last write.
	Stale int `json:"stale"`
	// Overlap is whether R + W > N, which makes every read quorum share a
	// replica with every write quorum.
	Overlap bool     `json:"overlap"`
	Steps   []string `json:"steps"` // the first few trials, in words
}

// Quorum writes a new version to w of n replicas, then reads r replicas and
// keeps the newest version seen, trials times, each time picking the
// replicas at random (seeded). When r + w > n every read quorum meets every
// write quorum, so no read is ever stale; otherwise some reads miss the
// latest write. Replicas never fail and the write finishes before the read;
// read repair and sloppy quorums are left out.
//
// Time: O(trials * n). Space: O(n).
func Quorum(n, r, w, trials int, seed int64) (QuorumResult, error) {
	if n < 1 || n > 50 || r < 1 || r > n || w < 1 || w > n || trials < 1 || trials > 100_000 {
		return QuorumResult{}, fmt.Errorf("%w: 1 to 50 replicas, R and W from 1 to N, 1 to 100000 trials", ErrInput)
	}
	rng := rand.New(rand.NewSource(seed))
	version := make([]int, n)
	res := QuorumResult{Trials: trials, Overlap: r+w > n}
	for t := 1; t <= trials; t++ {
		ws := rng.Perm(n)[:w]
		for _, i := range ws {
			version[i] = t
		}
		rs := rng.Perm(n)[:r]
		newest := 0
		for _, i := range rs {
			newest = max(newest, version[i])
		}
		if newest < t {
			res.Stale++
		}
		if len(res.Steps) < 5 {
			res.Steps = append(res.Steps, fmt.Sprintf("write v%d to %v, read %v: got v%d", t, ws, rs, newest))
		}
	}
	return res, nil
}
