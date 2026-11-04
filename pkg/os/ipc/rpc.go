package ipc

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

// RPC calls a remote procedure, "add A B" or "mul A B", over a network that
// drops each message with probability loss percent (seeded). The client stub
// marshals the call into bytes, sends it and waits; on a timeout it sends
// again, up to retries times. The server stub unmarshals, runs the
// procedure and marshals the reply. When the reply, not the request, is
// lost, the retry makes the server run the procedure twice: RPC is "at
// least once" unless the server remembers request ids (dedupe), which makes
// a non-idempotent call (like "transfer money") safe to retry.
//
// Time: O(retries). Space: O(retries).
func RPC(call string, loss, retries int, dedupe bool, seed int64) (Result, error) {
	f := strings.Fields(call)
	if len(f) != 3 || (f[0] != "add" && f[0] != "mul") {
		return Result{}, inputErr("call is \"add A B\" or \"mul A B\"")
	}
	a, err1 := strconv.Atoi(f[1])
	b, err2 := strconv.Atoi(f[2])
	if err1 != nil || err2 != nil || loss < 0 || loss > 90 || retries < 0 || retries > 10 {
		return Result{}, inputErr("whole-number arguments, loss 0-90 percent, retries 0-10")
	}
	rng := rand.New(rand.NewSource(seed))
	lost := func() bool { return rng.Intn(100) < loss }
	res := Result{Counts: map[string]int{"executions": 0, "attempts": 0}}
	wire := fmt.Sprintf("{id:7 proc:%q args:[%d %d]}", f[0], a, b)
	res.Log = append(res.Log, "client stub marshals "+call+" into "+wire)
	seen := map[int]int{}
	tick := 0
	for attempt := 0; attempt <= retries; attempt++ {
		res.Counts["attempts"]++
		res.Timeline = append(res.Timeline, Event{tick, "client", fmt.Sprintf("send request (try %d)", attempt+1), "running"})
		tick++
		if lost() {
			res.Log = append(res.Log, fmt.Sprintf("try %d: the request is lost; the client times out", attempt+1))
			res.Timeline = append(res.Timeline, Event{tick, "network", "request lost", "blocked"}, Event{tick, "client", "waiting, then timeout", "blocked"})
			tick += 2
			continue
		}
		var result int
		if r, ok := seen[7]; ok && dedupe {
			result = r
			res.Log = append(res.Log, fmt.Sprintf("try %d: server sees request id 7 again and resends the saved reply without running it", attempt+1))
			res.Timeline = append(res.Timeline, Event{tick, "server", "duplicate: resend saved reply", "running"})
		} else {
			if f[0] == "add" {
				result = a + b
			} else {
				result = a * b
			}
			seen[7] = result
			res.Counts["executions"]++
			res.Log = append(res.Log, fmt.Sprintf("try %d: server stub unmarshals, runs %s = %d, marshals the reply", attempt+1, call, result))
			res.Timeline = append(res.Timeline, Event{tick, "server", "unmarshal, run, marshal", "running"})
		}
		tick++
		if lost() {
			res.Log = append(res.Log, fmt.Sprintf("try %d: the reply is lost; the client times out and will retry", attempt+1))
			res.Timeline = append(res.Timeline, Event{tick, "network", "reply lost", "blocked"}, Event{tick, "client", "waiting, then timeout", "blocked"})
			tick += 2
			continue
		}
		res.Counts["result"] = result
		res.Log = append(res.Log, fmt.Sprintf("client stub unmarshals the reply: %d", result))
		res.Timeline = append(res.Timeline, Event{tick, "client", fmt.Sprintf("got %d", result), "done"})
		return res, nil
	}
	res.Log = append(res.Log, "the client gives up: the call may or may not have run")
	return res, nil
}
