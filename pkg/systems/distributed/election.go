package distributed

import (
	"fmt"
	"slices"
	"strconv"
)

// ElectionResult is the outcome of a leader election.
type ElectionResult struct {
	Leader   int      `json:"leader"` // the winner's id, -1 if none
	Messages int      `json:"messages"`
	Ticks    int      `json:"ticks"`
	Timeline Timeline `json:"timeline"`
	Steps    []string `json:"steps"`
}

type note struct {
	at, from, to int
	kind         string
}

// Bully runs the bully election (Hector Garcia-Molina, 1982) among
// processes 0..n-1, where those listed in down have crashed and starter
// notices the leader is gone. A process holding an election sends ELECTION
// to every higher id; a live higher process answers OK and holds its own;
// one that hears no OK within timeout ticks wins and sends COORDINATOR to
// everyone lower. The highest live id always wins. Messages take one tick;
// a crashed process answers nothing.
//
// Time: O(n^2) messages in the worst case (the lowest id starts). Space: O(n^2).
func Bully(n int, down []int, starter int) (ElectionResult, error) {
	if n < 2 || n > 26 || starter < 0 || starter >= n || slices.Contains(down, starter) {
		return ElectionResult{}, fmt.Errorf("%w: 2 to 26 processes, and the starter is alive", ErrInput)
	}
	alive := func(p int) bool { return !slices.Contains(down, p) }
	const timeout = 3
	res := ElectionResult{Leader: -1}
	lane := func(p int) string { return "P" + strconv.Itoa(p) }
	var queue []note
	electing := make([]int, n) // tick an election began, -1 if not holding one
	gotOK := make([]bool, n)
	for i := range electing {
		electing[i] = -1
	}
	send := func(at, from, to int, kind string) {
		res.Messages++
		queue = append(queue, note{at + 1, from, to, kind})
		res.Steps = append(res.Steps, fmt.Sprintf("t%d: P%d -> P%d %s", at, from, to, kind))
		res.Timeline.add(at, lane(from), kind+" to P"+strconv.Itoa(to), "running")
	}
	start := func(t, p int) {
		electing[p], gotOK[p] = t, false
		for q := p + 1; q < n; q++ {
			send(t, p, q, "ELECTION")
		}
	}
	start(0, starter)
	for t := 1; t < MaxTicks; t++ {
		due := queue
		queue = nil // what is sent while delivering goes out again
		for _, m := range due {
			if m.at != t {
				queue = append(queue, m)
				continue
			}
			if !alive(m.to) {
				res.Timeline.add(t, lane(m.to), "down: ignores "+m.kind, "blocked")
				continue
			}
			switch m.kind {
			case "ELECTION":
				send(t, m.to, m.from, "OK")
				if electing[m.to] == -1 && res.Leader < 0 {
					start(t, m.to)
				}
			case "OK":
				gotOK[m.to] = true
				res.Timeline.add(t, lane(m.to), "waits for a coordinator", "idle")
			case "COORDINATOR":
				res.Timeline.add(t, lane(m.to), "accepts P"+strconv.Itoa(m.from), "done")
			}
		}
		for p := 0; p < n; p++ {
			if electing[p] >= 0 && !gotOK[p] && t-electing[p] >= timeout && res.Leader < 0 {
				res.Leader = p
				res.Timeline.add(t, lane(p), "wins", "done")
				for q := 0; q < p; q++ {
					send(t, p, q, "COORDINATOR")
				}
			}
			if electing[p] >= 0 && (gotOK[p] || res.Leader >= 0) && t-electing[p] >= timeout {
				electing[p] = -2 // finished: lost, or someone won
			}
		}
		if res.Leader >= 0 && len(queue) == 0 {
			res.Ticks = t
			return res, nil
		}
	}
	return res, fmt.Errorf("%w: the election did not end", ErrInput)
}

// Ring runs the Chang-Roberts election (1979) on processes in a ring, ids
// listing their distinct ids clockwise. Every process sends its id to its
// clockwise neighbor; a process forwards an id larger than its own and drops
// a smaller one, so only the largest goes all the way round, and the
// process that gets its own id back is the leader and sends an ELECTED
// message round once. Every message moves one hop per tick.
//
// Time: O(n log n) messages on average, O(n^2) when the ids decrease
// clockwise, 2n-1 when they increase (plus n for ELECTED). Space: O(n).
func Ring(ids []int) (ElectionResult, error) {
	n := len(ids)
	if n < 2 || n > 64 {
		return ElectionResult{}, fmt.Errorf("%w: 2 to 64 processes in the ring", ErrInput)
	}
	seen := map[int]bool{}
	for _, id := range ids {
		if seen[id] || id < 0 {
			return ElectionResult{}, fmt.Errorf("%w: ids are distinct and not negative", ErrInput)
		}
		seen[id] = true
	}
	res := ElectionResult{Leader: -1}
	lane := func(i int) string { return "id " + strconv.Itoa(ids[i]) }
	type msg struct{ at, id int } // the message is at process at
	var flight []msg
	for i, id := range ids {
		res.Messages++
		flight = append(flight, msg{(i + 1) % n, id})
		res.Timeline.add(0, lane(i), "send "+strconv.Itoa(id), "running")
	}
	for t := 1; t <= n*n+n; t++ {
		var next []msg
		for _, m := range flight {
			me := ids[m.at]
			switch {
			case m.id > me:
				res.Messages++
				next = append(next, msg{(m.at + 1) % n, m.id})
				res.Timeline.add(t, lane(m.at), "forward "+strconv.Itoa(m.id), "running")
			case m.id < me:
				res.Timeline.add(t, lane(m.at), "drop "+strconv.Itoa(m.id), "idle")
			default:
				res.Leader = me
				res.Timeline.add(t, lane(m.at), "leader", "done")
				res.Steps = append(res.Steps, fmt.Sprintf("t%d: %d got its own id back and is leader", t, me))
				res.Messages += n // ELECTED goes round once
				res.Ticks = t + n
				return res, nil
			}
		}
		flight = next
	}
	return res, fmt.Errorf("%w: no leader", ErrInput)
}
