// Package infrastructure simulates the patterns that keep services up
// under load and failure, one file per idea, deterministic with a seed:
//
//	loadbalance.go  round robin, weighted, least connections, two choices
//	breaker.go      the circuit breaker's closed, open and half-open states
//	backoff.go      retries with exponential backoff, with and without jitter
//	caching.go      cache-aside, write-through and write-behind
//	stampede.go     a cache stampede on a hot key, and two fixes
//	cdn.go          a CDN's edge caches by region
//	delivery.go     at-most-once, at-least-once and exactly-once delivery
//	autoscale.go    autoscaling on a load curve
//	rollout.go      blue/green, canary and rolling deployments of a bad build
//	cgroups.go      container CPU shares and memory limits, with the OOM killer
//
// Rate limiting (token and leaky buckets) is in the networking package, and
// consistent hashing in the distributed and hashing packages.
package infrastructure

import (
	"errors"
	"fmt"
)

// ErrInput is wrapped by every error about invalid input.
var ErrInput = errors.New("infrastructure: invalid input")

// Event is what one lane did at one tick.
type Event struct {
	Tick  int    `json:"tick"`
	Lane  string `json:"lane"`
	What  string `json:"what"`
	State string `json:"state"` // "running", "blocked", "idle" or "done"
}

func inputErr(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInput, fmt.Sprintf(format, args...))
}
