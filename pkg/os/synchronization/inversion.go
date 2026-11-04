package synchronization

import "fmt"

// task is one task in PriorityInversion: a priority and a program of steps,
// "cpu", "lock", "cs" (inside the lock) or "unlock".
type task struct {
	name     string
	priority int
	arrival  int
	program  []string
	pc       int
	finish   int
}

// PriorityInversion runs three tasks on one CPU with a preemptive priority
// scheduler: Low (priority 1) arrives at 0 and takes a lock for three ticks;
// High (3) arrives at 2 and needs the same lock; Medium (2) arrives at 3 and
// only computes, for medium ticks. Without inheritance, High waits for the
// lock, Low cannot run because Medium outranks it, so High (the most urgent)
// waits for Medium (which does not even use the lock): priority inversion,
// as on the Mars Pathfinder in 1997. With inheritance, Low runs at High's
// priority while it holds the lock High wants, finishes the critical
// section, and High goes next. Counts["high finish"] is when High finished.
//
// Time: O(total ticks * 3). Space: O(total ticks).
func PriorityInversion(medium int, inherit bool) (Result, error) {
	if medium < 1 || medium > 20 {
		return Result{}, inputErr("medium work 1-20 ticks")
	}
	m := make([]string, medium)
	for i := range m {
		m[i] = "cpu"
	}
	tasks := []*task{
		{name: "low", priority: 1, arrival: 0, program: []string{"cpu", "lock", "cs", "cs", "cs", "unlock", "cpu"}},
		{name: "medium", priority: 2, arrival: 3, program: m},
		{name: "high", priority: 3, arrival: 2, program: []string{"lock", "cs", "unlock", "cpu"}},
	}
	holder := (*task)(nil)
	var waiter *task
	res := Result{Counts: map[string]int{}}
	for tick := 0; tick < 200; tick++ {
		effective := func(t *task) int {
			if inherit && t == holder && waiter != nil && waiter.priority > t.priority {
				return waiter.priority
			}
			return t.priority
		}
		var run *task
		for _, t := range tasks {
			if t.arrival > tick || t.pc == len(t.program) || (t.program[t.pc] == "lock" && holder != nil && holder != t) {
				continue
			}
			if run == nil || effective(t) > effective(run) {
				run = t
			}
		}
		if run == nil {
			done := true
			for _, t := range tasks {
				if t.pc < len(t.program) {
					done = false
				}
			}
			if done {
				break
			}
		}
		for _, t := range tasks {
			if t.arrival > tick || t.pc == len(t.program) {
				continue
			}
			switch {
			case t == run:
				what := t.program[t.pc]
				switch what {
				case "lock":
					holder = t
					if waiter == t {
						waiter = nil
					}
					what = "takes the lock"
				case "unlock":
					holder = nil
					what = "releases the lock"
				case "cs":
					what = "in the critical section"
					if effective(t) != t.priority {
						what += fmt.Sprintf(" (inherited priority %d)", effective(t))
					}
				default:
					what = "computing"
				}
				t.pc++
				if t.pc == len(t.program) {
					t.finish = tick + 1
				}
				res.Timeline = append(res.Timeline, Event{Tick: tick, Lane: t.name, What: what, State: "running"})
			case t.program[t.pc] == "lock" && holder != nil && holder != t:
				waiter = t
				res.Timeline = append(res.Timeline, Event{Tick: tick, Lane: t.name, What: "blocked on the lock held by " + holder.name, State: "blocked"})
			default:
				res.Timeline = append(res.Timeline, Event{Tick: tick, Lane: t.name, What: "ready, outranked", State: "idle"})
			}
		}
		res.Ticks = tick + 1
	}
	for _, t := range tasks {
		res.Counts[t.name+" finish"] = t.finish
		res.Log = append(res.Log, fmt.Sprintf("%s (priority %d) arrived at %d and finished at %d", t.name, t.priority, t.arrival, t.finish))
	}
	return res, nil
}
