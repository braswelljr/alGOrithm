package networking

import "fmt"

// LossKind is how the sender learns a segment was lost.
type LossKind int

// The ways a loss is detected.
const (
	// Timeout: no ack came back at all; the network may be badly congested.
	Timeout LossKind = iota + 1
	// TripleDupAck: three duplicate acks came back, so later segments got
	// through and only one was lost.
	TripleDupAck
)

// Round is the state of the congestion window during one round trip.
type Round struct {
	Round    int    `json:"round"`
	Cwnd     int    `json:"cwnd"` // segments sent this round
	Ssthresh int    `json:"ssthresh"`
	Phase    string `json:"phase"` // how cwnd got here: "slow start", "congestion avoidance", "after timeout", "fast recovery"
}

// Congestion runs TCP congestion control for rounds round trips, starting at
// one segment with the given slow-start threshold. losses maps a round to
// the loss detected at its end. In slow start the window doubles each round
// (capped at ssthresh); above it, it grows by one (additive increase). A
// timeout halves ssthresh and drops the window to one. Three duplicate acks
// halve ssthresh too; Reno (reno true) then continues from the new ssthresh
// (fast recovery), while Tahoe drops to one as for a timeout. The model
// counts whole segments per round trip and leaves out byte counting, the
// receiver's window and the inflation of the window during fast recovery.
//
// Time: O(rounds). Space: O(rounds).
func Congestion(rounds, ssthresh int, losses map[int]LossKind, reno bool) ([]Round, error) {
	if rounds < 1 || rounds > 1000 || ssthresh < 2 || ssthresh > 1<<20 {
		return nil, fmt.Errorf("%w: 1 to 1000 rounds and a threshold of at least 2", ErrInput)
	}
	out := make([]Round, 0, rounds)
	cwnd, phase := 1, "slow start"
	for r := 1; r <= rounds; r++ {
		out = append(out, Round{Round: r, Cwnd: cwnd, Ssthresh: ssthresh, Phase: phase})
		switch losses[r] {
		case Timeout:
			ssthresh, cwnd, phase = max(cwnd/2, 2), 1, "after timeout"
		case TripleDupAck:
			ssthresh = max(cwnd/2, 2)
			if reno {
				cwnd, phase = ssthresh, "fast recovery"
			} else {
				cwnd, phase = 1, "after timeout"
			}
		default:
			if cwnd < ssthresh {
				cwnd, phase = min(2*cwnd, ssthresh), "slow start"
			} else {
				cwnd, phase = cwnd+1, "congestion avoidance"
			}
		}
	}
	return out, nil
}
