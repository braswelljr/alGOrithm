package infrastructure

import (
	"fmt"
	"strings"
)

// BreakerState is a circuit breaker's state.
type BreakerState string

// The three states.
const (
	Closed   BreakerState = "closed"    // calls go through; failures are counted
	Open     BreakerState = "open"      // calls fail fast, without reaching the service
	HalfOpen BreakerState = "half-open" // one trial call decides
)

// BreakerResult is a CircuitBreaker run: the state before each call, the
// timeline, the log and the counts.
type BreakerResult struct {
	States   []BreakerState `json:"states"`
	Timeline []Event        `json:"timeline"`
	Log      []string       `json:"log"`
	Rejected int            `json:"rejected"` // calls that failed fast
	Failed   int            `json:"failed"`   // calls that reached the service and failed
}

// CircuitBreaker makes one call a tick to a service whose answers are
// outcomes ("ok" or "fail", separated by spaces: what the service would
// answer at that tick). After threshold failures in a row the breaker
// opens: calls fail fast for cooldown ticks, which gives the service room
// to recover and keeps callers from waiting on timeouts. Then it goes
// half-open and lets one call through: success closes it, failure opens it
// again.
//
// Time: O(n). Space: O(n).
func CircuitBreaker(outcomes string, threshold, cooldown int) (BreakerResult, error) {
	calls := strings.Fields(outcomes)
	if len(calls) == 0 || len(calls) > 200 || threshold < 1 || threshold > 20 || cooldown < 1 || cooldown > 50 {
		return BreakerResult{}, inputErr("1-200 outcomes, threshold 1-20, cooldown 1-50")
	}
	state, fails, openedAt := Closed, 0, 0
	var res BreakerResult
	set := func(tick int, to BreakerState, why string) {
		res.Log = append(res.Log, fmt.Sprintf("tick %d: %s -> %s: %s", tick, state, to, why))
		state = to
	}
	for tick, c := range calls {
		if c != "ok" && c != "fail" {
			return res, inputErr("outcome %d is %q: want ok or fail", tick+1, c)
		}
		if state == Open && tick-openedAt >= cooldown {
			set(tick, HalfOpen, fmt.Sprintf("the cooldown of %d ticks is over: try one call", cooldown))
		}
		res.States = append(res.States, state)
		res.Timeline = append(res.Timeline, Event{tick, "breaker", string(state), map[BreakerState]string{Closed: "running", Open: "blocked", HalfOpen: "idle"}[state]})
		if state == Open {
			res.Rejected++
			res.Timeline = append(res.Timeline, Event{tick, "call", "rejected fast", "blocked"})
			continue
		}
		res.Timeline = append(res.Timeline, Event{tick, "call", c, map[string]string{"ok": "running", "fail": "blocked"}[c]})
		if c == "ok" {
			fails = 0
			if state == HalfOpen {
				set(tick, Closed, "the trial call succeeded")
			}
			continue
		}
		res.Failed++
		fails++
		if state == HalfOpen || fails >= threshold {
			why := fmt.Sprintf("%d failures in a row", fails)
			if state == HalfOpen {
				why = "the trial call failed"
			}
			set(tick, Open, why)
			openedAt, fails = tick, 0
		}
	}
	return res, nil
}
