package ipc

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// MessageQueue runs a script against a message queue of capacity messages,
// one call a line: "send PRIORITY TEXT" or "receive". Messages keep their
// boundaries (unlike a pipe's bytes), and receive takes the highest
// priority first, oldest first among equals, as POSIX message queues do. A
// send to a full queue fails with EAGAIN (a non-blocking queue), and a
// receive from an empty one too.
//
// Time: O(n log c) for n calls and capacity c. Space: O(c).
func MessageQueue(script string, capacity int) (Result, error) {
	if capacity < 1 || capacity > 32 {
		return Result{}, inputErr("capacity 1-32 messages")
	}
	type msg struct {
		prio, seq int
		text      string
	}
	var q []msg
	res := Result{Counts: map[string]int{}}
	seq := 0
	for n, line := range strings.Split(script, "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		switch f[0] {
		case "send":
			if len(f) < 3 {
				return res, inputErr("line %d: send PRIORITY TEXT", n+1)
			}
			p, err := strconv.Atoi(f[1])
			if err != nil || p < 0 || p > 31 {
				return res, inputErr("line %d: priority 0-31", n+1)
			}
			text := strings.Join(f[2:], " ")
			if len(q) == capacity {
				res.Counts["full"]++
				res.Log = append(res.Log, fmt.Sprintf("send %q (priority %d): EAGAIN, the queue is full", text, p))
				continue
			}
			q = append(q, msg{p, seq, text})
			seq++
			sort.SliceStable(q, func(i, j int) bool { return q[i].prio > q[j].prio })
			res.Counts["sent"]++
			res.Log = append(res.Log, fmt.Sprintf("send %q (priority %d): queued, %d waiting", text, p, len(q)))
		case "receive":
			if len(q) == 0 {
				res.Counts["empty"]++
				res.Log = append(res.Log, "receive: EAGAIN, the queue is empty")
				continue
			}
			m := q[0]
			q = q[1:]
			res.Counts["received"]++
			res.Log = append(res.Log, fmt.Sprintf("receive: %q (priority %d, sent %s)", m.text, m.prio, ordinal(m.seq+1)))
		default:
			return res, inputErr("line %d: send or receive", n+1)
		}
	}
	return res, nil
}

func ordinal(n int) string {
	suffix := "th"
	switch {
	case n%100 >= 11 && n%100 <= 13:
	case n%10 == 1:
		suffix = "st"
	case n%10 == 2:
		suffix = "nd"
	case n%10 == 3:
		suffix = "rd"
	}
	return strconv.Itoa(n) + suffix
}
