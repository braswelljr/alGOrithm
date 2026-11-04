package ipc

import (
	"fmt"
	"math/rand"
)

// LocalSocket simulates a server on a Unix-domain socket: socket, bind,
// listen with a backlog, then accept one connection every acceptEvery
// ticks, while the clients call connect at seeded random ticks in
// [0, 2*clients). A connection the server has not accepted yet waits in the
// backlog queue; when the queue is full, connect is refused (ECONNREFUSED
// on a Unix socket; TCP would drop the SYN and the client would retry).
// Counts holds "accepted" and "refused".
//
// Time: O(T + clients). Space: O(clients).
func LocalSocket(clients, backlog, acceptEvery int, seed int64) (Result, error) {
	if clients < 1 || clients > 30 || backlog < 0 || backlog > 30 || acceptEvery < 1 || acceptEvery > 10 {
		return Result{}, inputErr("1-30 clients, backlog 0-30, accept every 1-10 ticks")
	}
	rng := rand.New(rand.NewSource(seed))
	when := make([]int, clients)
	for i := range when {
		when[i] = rng.Intn(2 * clients)
	}
	res := Result{Counts: map[string]int{"accepted": 0, "refused": 0}}
	res.Log = append(res.Log, "server: socket(AF_UNIX), bind(\"/tmp/app.sock\"), listen(backlog "+fmt.Sprint(backlog)+")")
	var queue []int
	done := 0
	for tick := 0; tick < 1000 && done < clients; tick++ {
		for c, t := range when {
			if t != tick {
				continue
			}
			lane := fmt.Sprintf("client %d", c)
			if len(queue) < backlog {
				queue = append(queue, c)
				res.Timeline = append(res.Timeline, Event{tick, lane, "connect: queued in the backlog", "blocked"})
			} else {
				res.Counts["refused"]++
				done++
				res.Log = append(res.Log, fmt.Sprintf("tick %d: client %d refused: the backlog of %d is full", tick, c, backlog))
				res.Timeline = append(res.Timeline, Event{tick, lane, "connect: refused", "done"})
			}
		}
		if tick%acceptEvery == 0 && len(queue) > 0 {
			c := queue[0]
			queue = queue[1:]
			res.Counts["accepted"]++
			done++
			res.Log = append(res.Log, fmt.Sprintf("tick %d: server accepts client %d: a new socket just for it; send, recv, close", tick, c))
			res.Timeline = append(res.Timeline, Event{tick, "server", fmt.Sprintf("accept client %d", c), "running"},
				Event{tick, fmt.Sprintf("client %d", c), "connected", "running"})
		} else {
			res.Timeline = append(res.Timeline, Event{tick, "server", "busy with a connection", "idle"})
		}
	}
	return res, nil
}
