package processes

import "strings"

// State is a process's state in the five-state model.
type State string

// The five states.
const (
	New        State = "new"
	Ready      State = "ready"
	Running    State = "running"
	Waiting    State = "waiting"
	Terminated State = "terminated"
)

// Transition is one move between states, caused by Event at Step.
type Transition struct {
	Step  int    `json:"step"`
	Event string `json:"event"`
	From  State  `json:"from"`
	To    State  `json:"to"`
	Why   string `json:"why"`
}

// lifecycle is the transition table: from a state, an event leads to a state.
var lifecycle = map[State]map[string]struct {
	to  State
	why string
}{
	New:     {"admit": {Ready, "the OS has built its PCB and address space, so it may now be scheduled"}},
	Ready:   {"dispatch": {Running, "the scheduler picked it and loaded its registers onto the CPU"}},
	Running: {"timeout": {Ready, "its time slice ran out; it can still run, it just has to wait its turn"}, "wait": {Waiting, "it asked for I/O or a lock and cannot continue until that happens"}, "exit": {Terminated, "it finished; the parent can now collect its exit status"}},
	Waiting: {"wake": {Ready, "the I/O finished, so it can run again, but it goes back in line"}},
}

// Lifecycle drives one process through events ("admit", "dispatch",
// "timeout", "wait", "wake", "exit"), separated by spaces or new lines, and
// returns every transition. An event the current state does not allow is an
// error naming the step: a waiting process cannot be dispatched, only woken.
//
// Time: O(n) for n events. Space: O(n).
func Lifecycle(events string) ([]Transition, error) {
	state := New
	var out []Transition
	for i, ev := range strings.Fields(events) {
		next, ok := lifecycle[state][ev]
		if !ok {
			return out, inputErr(0, "step %d: %q is not allowed in state %s", i+1, ev, state)
		}
		out = append(out, Transition{Step: i + 1, Event: ev, From: state, To: next.to, Why: next.why})
		state = next.to
	}
	return out, nil
}
