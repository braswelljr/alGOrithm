package infrastructure

import (
	"fmt"
	"math"
	"math/rand"
)

// StampedeFix is how a cache protects a hot key's recomputation.
type StampedeFix string

// The fixes Stampede knows.
const (
	NoFix       StampedeFix = "none"  // every miss recomputes
	LockFix     StampedeFix = "lock"  // the first miss takes a lock and recomputes; the rest serve the stale value
	EarlyExpiry StampedeFix = "early" // recompute a little before expiry, chosen at random (XFetch)
)

// StampedeResult counts database queries at each tick.
type StampedeResult struct {
	DBPerTick []int    `json:"dbPerTick"`
	Peak      int      `json:"peak"`
	Total     int      `json:"total"`
	Log       []string `json:"log"`
}

// Stampede sends rate requests a tick for the given number of ticks for one hot key
// cached for ttl ticks; recomputing it takes recompute ticks of database
// work. When it expires with no protection, every request in the next
// recompute ticks misses and queries the database at once: a stampede (the
// thundering herd) that can take the database down. A lock lets one request
// recompute while the others serve the stale value; probabilistic early
// expiry refreshes it before it expires, at a random moment, so usually no
// one ever misses.
//
// Time: O(ticks * rate). Space: O(ticks).
func Stampede(rate, ticks, ttl, recompute int, fix StampedeFix, seed int64) (StampedeResult, error) {
	if fix != NoFix && fix != LockFix && fix != EarlyExpiry {
		return StampedeResult{}, inputErr("unknown fix %q", fix)
	}
	if rate < 1 || rate > 100 || ticks < 1 || ticks > 300 || ttl < 2 || ttl > 100 || recompute < 1 || recompute >= ttl {
		return StampedeResult{}, inputErr("rate 1-100, 1-300 ticks, ttl 2-100, recompute 1 to ttl-1")
	}
	rng := rand.New(rand.NewSource(seed))
	res := StampedeResult{DBPerTick: make([]int, ticks)}
	expires, readyAt, recomputing := ttl, -1, false
	for t := 0; t < ticks; t++ {
		if recomputing && t >= readyAt {
			recomputing, expires = false, t+ttl
		}
		for r := 0; r < rate; r++ {
			switch fix {
			case NoFix:
				if t >= expires {
					res.DBPerTick[t]++ // each miss queries the database
					if !recomputing {
						recomputing, readyAt = true, t+recompute
					}
				}
			case LockFix:
				if t >= expires && !recomputing {
					res.DBPerTick[t]++
					recomputing, readyAt = true, t+recompute
				}
			case EarlyExpiry:
				// XFetch: refresh early with probability rising toward expiry.
				early := float64(recompute) * -math.Log(1-rng.Float64()+1e-12)
				if !recomputing && float64(t)+early >= float64(expires) {
					res.DBPerTick[t]++
					recomputing, readyAt = true, t+recompute
				}
			}
		}
	}
	for _, n := range res.DBPerTick {
		res.Peak = max(res.Peak, n)
		res.Total += n
	}
	res.Log = append(res.Log, fmt.Sprintf("%s: %d database queries, at most %d in one tick", fix, res.Total, res.Peak))
	return res, nil
}
