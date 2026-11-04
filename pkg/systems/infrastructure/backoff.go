package infrastructure

import (
	"fmt"
	"math/rand"
)

// Jitter is how retry delays are randomized.
type Jitter string

// The jitter kinds of the AWS architecture blog's "Exponential Backoff And
// Jitter".
const (
	NoJitter    Jitter = "none"  // delay = min(cap, base * 2^attempt)
	FullJitter  Jitter = "full"  // a random delay in [0, that]
	EqualJitter Jitter = "equal" // half of it, plus a random half
)

// BackoffResult is a Backoff run: how many retries hit the server at each
// tick, and the peak.
type BackoffResult struct {
	PerTick []int    `json:"perTick"`
	Peak    int      `json:"peak"`
	Total   int      `json:"total"`
	Log     []string `json:"log"`
}

// Backoff has every client fail at tick 0 (the server blipped) and
// retry attempts times, waiting min(capDelay, base * 2^k) ticks before retry
// k, randomized by jitter with a seed. Without jitter every client retries
// at the same ticks, so the retries arrive in synchronized spikes that can
// knock the server over again; with full jitter they spread out over the
// whole window.
//
// Time: O(clients * attempts + T). Space: O(T).
func Backoff(clients, attempts, base, capDelay int, jitter Jitter, seed int64) (BackoffResult, error) {
	if jitter != NoJitter && jitter != FullJitter && jitter != EqualJitter {
		return BackoffResult{}, inputErr("unknown jitter %q", jitter)
	}
	if clients < 1 || clients > 500 || attempts < 1 || attempts > 10 || base < 1 || base > 100 || capDelay < base || capDelay > 1000 {
		return BackoffResult{}, inputErr("1-500 clients, 1-10 attempts, base 1-100, cap from base to 1000")
	}
	rng := rand.New(rand.NewSource(seed))
	var res BackoffResult
	hits := map[int]int{}
	last := 0
	for c := 0; c < clients; c++ {
		t := 0
		for k := 0; k < attempts; k++ {
			d := min(capDelay, base<<k)
			switch jitter {
			case FullJitter:
				d = rng.Intn(d + 1)
			case EqualJitter:
				d = d/2 + rng.Intn(d/2+1)
			case NoJitter:
			}
			t += max(d, 1)
			hits[t]++
			last = max(last, t)
		}
	}
	res.PerTick = make([]int, last+1)
	for t, n := range hits {
		res.PerTick[t] = n
		res.Peak = max(res.Peak, n)
		res.Total += n
	}
	res.Log = append(res.Log, fmt.Sprintf("%d retries over %d ticks; at most %d in one tick (%s jitter)", res.Total, last, res.Peak, jitter))
	return res, nil
}
