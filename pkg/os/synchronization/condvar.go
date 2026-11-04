package synchronization

import "fmt"

// monitor is a lock, a condition variable and the buffer they guard.
type monitor struct {
	owner     int // -1 when free
	buffer    int
	sleepers  []int
	signaled  map[int]bool
	underflow int
}

type cvConsumer struct {
	id, pc    int
	useWhile  bool
	m         *monitor
	remaining int
}

func (t *cvConsumer) step() (string, string, bool) {
	m := t.m
	switch t.pc {
	case 0: // lock
		if m.owner != -1 {
			return "waits for the lock", "blocked", false
		}
		m.owner, t.pc = t.id, 1
		return "locks", "running", false
	case 1: // if / while (buffer empty) wait
		if m.buffer == 0 {
			m.sleepers = append(m.sleepers, t.id)
			m.owner, t.pc = -1, 2
			return "buffer empty: waits on the condition (unlocks)", "blocked", false
		}
		t.pc = 3
		return "buffer has an item", "running", false
	case 2: // asleep until signaled, then reacquire the lock
		if !m.signaled[t.id] {
			return "asleep on the condition", "blocked", false
		}
		if m.owner != -1 {
			return "signaled, waits to relock", "blocked", false
		}
		delete(m.signaled, t.id)
		m.owner = t.id
		if t.useWhile {
			t.pc = 1
			return "relocks, checks again (while)", "running", false
		}
		t.pc = 3
		return "relocks, assumes an item is there (if)", "running", false
	case 3: // take
		if m.buffer == 0 {
			m.underflow++
			t.pc = 4
			return "takes from an EMPTY buffer: another consumer got there first", "running", false
		}
		m.buffer--
		t.pc = 4
		return fmt.Sprintf("takes an item, %d left", m.buffer), "running", false
	default:
		m.owner = -1
		t.remaining--
		t.pc = 0
		return "unlocks", "running", t.remaining == 0
	}
}

type cvProducer struct {
	id, pc, remaining int
	m                 *monitor
}

func (t *cvProducer) step() (string, string, bool) {
	m := t.m
	switch t.pc {
	case 0:
		if m.owner != -1 {
			return "waits for the lock", "blocked", false
		}
		m.owner, t.pc = t.id, 1
		return "locks", "running", false
	case 1:
		m.buffer++
		t.pc = 2
		if len(m.sleepers) > 0 {
			w := m.sleepers[0]
			m.sleepers = m.sleepers[1:]
			m.signaled[w] = true
			return fmt.Sprintf("adds an item, signals consumer %d", w), "running", false
		}
		return "adds an item, nobody to signal", "running", false
	default:
		m.owner = -1
		t.remaining--
		t.pc = 0
		return "unlocks", "running", t.remaining == 0
	}
}

// ConditionVariable runs one producer adding the given items and the
// consumers sharing them out, around a lock and a condition variable with
// Mesa semantics (as in Go's sync.Cond and POSIX): signal only wakes a
// sleeper, which must then relock, so another consumer can take the item in
// between. If the wait sits in an if, the woken consumer takes from an empty
// buffer (Counts["underflows"]); in a while loop, it checks again and goes
// back to sleep. That is why the rule is "always wait in a loop".
//
// Time: O(items * consumers) ticks. Space: O(consumers).
func ConditionVariable(useWhile bool, consumers, items int, seed int64) (Result, error) {
	if consumers < 2 || consumers > 6 || items < 1 || items > 20 || items%consumers != 0 {
		return Result{}, inputErr("2-6 consumers, 1-20 items, items a multiple of consumers")
	}
	m := &monitor{owner: -1, signaled: map[int]bool{}}
	ts := []thread{&cvProducer{id: consumers, remaining: items, m: m}}
	lanes := []string{"producer"}
	for i := 0; i < consumers; i++ {
		ts = append(ts, &cvConsumer{id: i, useWhile: useWhile, m: m, remaining: items / consumers})
		lanes = append(lanes, fmt.Sprintf("consumer %d", i))
	}
	res := schedule(ts, lanes, seed, 5000, func() int { return 0 }, 1)
	res.Counts = map[string]int{"underflows": m.underflow}
	res.Violated = m.underflow > 0
	return res, nil
}
