package distributed

import (
	"fmt"
	"math/rand"
	"slices"
	"strconv"
)

// SnapshotResult is a recorded global state and whether it is consistent.
type SnapshotResult struct {
	// Balances is each process's recorded balance, Channels the money
	// recorded in flight on each channel ("P0->P1": amount).
	Balances []int          `json:"balances"`
	Channels map[string]int `json:"channels"`
	// Total is the recorded balances plus the recorded money in flight;
	// it equals the starting total exactly when the snapshot is consistent.
	Total    int      `json:"total"`
	Expected int      `json:"expected"`
	Timeline Timeline `json:"timeline"`
	Steps    []string `json:"steps"`
}

type transfer struct {
	at, seq, from, to, amount int // seq orders sends, so a channel stays FIFO
	marker                    bool
}

// ChandyLamport runs the Chandy-Lamport snapshot (1985) on n processes that
// start with 100 each and send each other random amounts (seeded) over
// FIFO channels with random delays, transfers times in all. At tick snapAt,
// process 0 records its balance and sends a MARKER on every outgoing
// channel. A process that receives its first marker records its balance,
// records that channel as empty, sends markers on, and records every other
// incoming channel until a marker arrives on it. The recorded state never
// happened all at once, yet it is consistent: no money is created or lost,
// which the run checks. Channels never lose or reorder messages.
//
// Time: O(transfers + n^2) messages. Space: O(n^2).
func ChandyLamport(n, transfers, snapAt int, seed int64) (SnapshotResult, error) {
	if n < 2 || n > 10 || transfers < 0 || transfers > 1000 || snapAt < 0 {
		return SnapshotResult{}, fmt.Errorf("%w: 2 to 10 processes, 0 to 1000 transfers", ErrInput)
	}
	r := rand.New(rand.NewSource(seed))
	p := func(i int) string { return "P" + strconv.Itoa(i) }
	balance := make([]int, n)
	for i := range balance {
		balance[i] = 100
	}
	res := SnapshotResult{Expected: 100 * n, Balances: make([]int, n), Channels: map[string]int{}}
	recorded := make([]bool, n)
	recording := map[[2]int]bool{} // channel (from, to) being recorded by to
	markersIn := make([]int, n)
	lastAt := map[[2]int]int{} // FIFO: a channel's latest delivery tick
	var flight []transfer
	seq := 0
	post := func(t int, m transfer) {
		seq++
		m.seq = seq
		ch := [2]int{m.from, m.to}
		m.at = max(t+1+r.Intn(3), lastAt[ch])
		lastAt[ch] = m.at
		flight = append(flight, m)
	}
	record := func(t, i int) {
		recorded[i] = true
		res.Balances[i] = balance[i]
		res.Timeline.add(t, p(i), "records "+strconv.Itoa(balance[i]), "done")
		res.Steps = append(res.Steps, fmt.Sprintf("t%d: %s records balance %d", t, p(i), balance[i]))
		for j := 0; j < n; j++ {
			if j != i {
				recording[[2]int{j, i}] = true
				post(t, transfer{from: i, to: j, marker: true})
			}
		}
	}
	sent := 0
	for t := 0; t < MaxTicks; t++ {
		// deliveries, in channel order
		var due []transfer
		rest := flight[:0:0]
		for _, m := range flight {
			if m.at == t {
				due = append(due, m)
			} else {
				rest = append(rest, m)
			}
		}
		flight = rest // markers sent while delivering join the later deliveries
		slices.SortFunc(due, func(a, b transfer) int { return a.seq - b.seq })
		for _, m := range due {
			ch := [2]int{m.from, m.to}
			if m.marker {
				markersIn[m.to]++
				if !recorded[m.to] {
					record(t, m.to)
				}
				recording[ch] = false // this channel's state is complete
				continue
			}
			balance[m.to] += m.amount
			res.Timeline.add(t, p(m.to), fmt.Sprintf("+%d from %s", m.amount, p(m.from)), "running")
			if recorded[m.to] && recording[ch] {
				res.Channels[p(m.from)+"->"+p(m.to)] += m.amount
			}
		}
		if t == snapAt && !recorded[0] {
			record(t, 0)
		}
		if sent < transfers {
			from := r.Intn(n)
			to := (from + 1 + r.Intn(n-1)) % n
			amount := 0
			if balance[from] > 0 {
				amount = 1 + r.Intn(min(balance[from], 30))
			}
			if amount > 0 {
				balance[from] -= amount
				post(t, transfer{from: from, to: to, amount: amount})
				res.Timeline.add(t, p(from), fmt.Sprintf("-%d to %s", amount, p(to)), "running")
			}
			sent++
		}
		done := len(flight) == 0 && sent >= transfers
		for i := range markersIn {
			done = done && recorded[i] && markersIn[i] == n-1
		}
		if done {
			break
		}
		if t > snapAt && len(flight) == 0 && sent >= transfers && !recorded[0] {
			break
		}
	}
	if !recorded[0] {
		return res, fmt.Errorf("%w: the transfers ended before the snapshot tick", ErrInput)
	}
	for _, b := range res.Balances {
		res.Total += b
	}
	for _, v := range res.Channels {
		res.Total += v
	}
	return res, nil
}
