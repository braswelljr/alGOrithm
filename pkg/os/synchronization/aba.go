package synchronization

import (
	"fmt"
	"strings"
)

// stackState is a lock-free stack of nodes named by letter; next maps a
// node to the one below it ("" for none).
type stackState struct {
	top   string
	tag   int
	next  map[string]string
	freed map[string]bool
}

func (s *stackState) list() string {
	var out []string
	seen := map[string]bool{}
	for n := s.top; n != "" && !seen[n]; n = s.next[n] {
		seen[n] = true
		label := n
		if s.freed[n] {
			label += " (freed!)"
		}
		out = append(out, label)
	}
	if len(out) == 0 {
		return "empty"
	}
	return strings.Join(out, " -> ")
}

// ABAResult is the run: the log and whether the stack ended corrupted.
type ABAResult struct {
	Log       []string `json:"log"`
	Final     string   `json:"final"`
	Corrupted bool     `json:"corrupted"`
}

// ABA replays the ABA problem in a lock-free stack (top A -> B -> C).
// Thread 1 starts a pop: it reads top = A and next = B, and is preempted
// before its compare-and-swap. Thread 2 pops A, pops B (freeing it), and
// pushes A back. Thread 1 resumes: top is A again, so CAS(top, A, B)
// succeeds, and the top now points at the freed node B. The value is the
// same, but the stack is not. With tagged pointers (tagged), every change
// bumps a counter stored with the pointer, so thread 1's CAS compares
// (A, tag 0) with (A, tag 3), fails, and it retries correctly.
//
// Time: O(1). Space: O(1).
func ABA(tagged bool) ABAResult {
	s := &stackState{top: "A", next: map[string]string{"A": "B", "B": "C", "C": ""}, freed: map[string]bool{}}
	var log []string
	say := func(f string, args ...any) { log = append(log, fmt.Sprintf(f, args...)) }
	say("start: %s", s.list())
	seenTop, seenNext, seenTag := s.top, s.next[s.top], s.tag
	say("thread 1: pop reads top = %s, next = %s (tag %d), then is preempted before its CAS", seenTop, seenNext, seenTag)
	pop := func() string {
		n := s.top
		s.top = s.next[n]
		s.tag++
		return n
	}
	a := pop()
	say("thread 2: pops %s; stack: %s", a, s.list())
	b := pop()
	s.freed[b] = true
	say("thread 2: pops %s and frees it; stack: %s", b, s.list())
	s.next[a] = s.top
	s.top = a
	s.tag++
	say("thread 2: pushes %s back (its memory is reused); stack: %s", a, s.list())
	if s.top == seenTop && (!tagged || s.tag == seenTag) {
		s.top = seenNext
		s.tag++
		say("thread 1: CAS(top, %s, %s) succeeds, because top is %s again", seenTop, seenNext, seenTop)
	} else {
		say("thread 1: CAS((%s, tag %d), (%s, tag %d)) fails: the tag is now %d, so the stack changed under it", seenTop, seenTag, seenNext, seenTag+1, s.tag)
		n := pop()
		say("thread 1: retries and pops %s correctly", n)
	}
	final := s.list()
	say("end: %s", final)
	return ABAResult{Log: log, Final: final, Corrupted: s.freed[s.top] || strings.Contains(final, "freed")}
}
