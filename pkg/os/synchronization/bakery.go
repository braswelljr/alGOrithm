package synchronization

import "fmt"

// bakeryThread runs Lamport's bakery algorithm, one shared read or write a
// step.
type bakeryThread struct {
	id, rounds, csLen  int
	choosing           []bool
	number             []int
	inside             *int
	pc, j, max, csLeft int
}

func (t *bakeryThread) step() (string, string, bool) {
	n := len(t.number)
	switch t.pc {
	case 0: // take a ticket: say we are choosing
		t.choosing[t.id], t.j, t.max = true, 0, 0
		t.pc = 1
		return "choosing a ticket", "running", false
	case 1: // read every ticket, one at a time
		t.max = max(t.max, t.number[t.j])
		t.j++
		if t.j == n {
			t.pc = 2
		}
		return fmt.Sprintf("reads ticket %d", t.j-1), "running", false
	case 2:
		t.number[t.id] = t.max + 1
		t.choosing[t.id] = false
		t.pc, t.j = 3, 0
		return fmt.Sprintf("takes ticket %d", t.max+1), "running", false
	case 3: // wait for everyone with a smaller (ticket, id)
		if t.j == n {
			t.pc, t.csLeft = 4, t.csLen
			*t.inside++
			return "enters the critical section", "running", false
		}
		j := t.j
		if j != t.id && t.choosing[j] {
			return fmt.Sprintf("waits: %d is still choosing", j), "blocked", false
		}
		if j != t.id && t.number[j] != 0 && (t.number[j] < t.number[t.id] || (t.number[j] == t.number[t.id] && j < t.id)) {
			return fmt.Sprintf("waits for %d (ticket %d)", j, t.number[j]), "blocked", false
		}
		t.j++
		return fmt.Sprintf("checked %d", j), "running", false
	case 4:
		t.csLeft--
		if t.csLeft > 0 {
			return "in the critical section", "running", false
		}
		*t.inside--
		t.number[t.id] = 0
		t.rounds--
		t.pc = 0
		return "leaves, ticket back to 0", "running", t.rounds == 0
	}
	return "", "done", true
}

// Bakery runs the given number of threads through Lamport's bakery algorithm (1974),
// each entering the critical section rounds times for csLen ticks. A thread
// takes a ticket one higher than every ticket it reads, then waits for every
// thread holding a smaller (ticket, id); the choosing flags stop a thread
// from being overtaken while it is still reading. It needs only reads and
// writes, no atomic instructions, and works for any number of threads.
//
// Time: O(threads^2 * rounds) steps. Space: O(threads).
func Bakery(threads, rounds, csLen int, seed int64) (Result, error) {
	if threads < 2 || threads > 6 || rounds < 1 || rounds > 5 || csLen < 1 || csLen > 5 {
		return Result{}, inputErr("2-6 threads, 1-5 rounds, critical section 1-5 ticks")
	}
	choosing, number, inside := make([]bool, threads), make([]int, threads), 0
	ts := make([]thread, threads)
	for i := range ts {
		ts[i] = &bakeryThread{id: i, rounds: rounds, csLen: csLen, choosing: choosing, number: number, inside: &inside}
	}
	return schedule(ts, names("thread", threads), seed, 5000, func() int { return inside }, 1), nil
}
