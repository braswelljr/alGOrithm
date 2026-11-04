// Package scheduling simulates CPU scheduling, the operating system's choice
// of which ready process runs next: first come first served, shortest job
// first, shortest remaining time first, round robin, priority (with and
// without preemption) and a multilevel feedback queue, one file each.
//
// Each takes the processes (name, arrival time, CPU burst, priority) and
// returns a Schedule: the timeline of who ran when (a Gantt chart), and each
// process's start, finish, turnaround, waiting and response times. Time is in
// abstract ticks, and switching processes costs nothing, as in the textbook
// examples:
//
//	ps, _ := scheduling.Parse("P1 0 5\nP2 1 3\nP3 2 8")
//	s := scheduling.RoundRobin(ps, 2)
//	fmt.Println(s.AvgWaiting)
package scheduling

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Process is one job: it arrives at Arrival, needs Burst ticks of CPU, and has
// a Priority (a lower number is more urgent).
type Process struct {
	Name     string `json:"name"`
	Arrival  int    `json:"arrival"`
	Burst    int    `json:"burst"`
	Priority int    `json:"priority"`
}

// Slice is a stretch of time [Start, End) during which Process ran; an empty
// Process means the CPU was idle.
type Slice struct {
	Process string `json:"process"`
	Start   int    `json:"start"`
	End     int    `json:"end"`
	// Queue is the level of a multilevel feedback queue it ran at, else 0.
	Queue int `json:"queue,omitempty"`
}

// Stats are one process's times. Turnaround is Finish - Arrival, Waiting is
// Turnaround - Burst (time ready but not running), and Response is Start -
// Arrival (time until it first ran).
type Stats struct {
	Process
	Start      int `json:"start"`
	Finish     int `json:"finish"`
	Turnaround int `json:"turnaround"`
	Waiting    int `json:"waiting"`
	Response   int `json:"response"`
}

// Schedule is the result of a scheduling algorithm.
type Schedule struct {
	Timeline        []Slice `json:"timeline"`
	Stats           []Stats `json:"stats"` // in the order the processes were given
	AvgWaiting      float64 `json:"avgWaiting"`
	AvgTurnaround   float64 `json:"avgTurnaround"`
	AvgResponse     float64 `json:"avgResponse"`
	ContextSwitches int     `json:"contextSwitches"` // changes from one process to another
	Makespan        int     `json:"makespan"`        // when the last process finished
}

// MaxProcesses and MaxTime bound the input the simulations accept.
const (
	MaxProcesses = 100
	MaxTime      = 100_000
)

// ErrInput is returned for an empty list, a burst below 1, a negative
// arrival, repeated names, or more than MaxProcesses processes or MaxTime
// ticks of work.
var ErrInput = errors.New("scheduling: invalid processes")

func validate(ps []Process) error {
	if len(ps) == 0 || len(ps) > MaxProcesses {
		return fmt.Errorf("%w: give 1 to %d processes", ErrInput, MaxProcesses)
	}
	seen := map[string]bool{}
	total := 0
	for _, p := range ps {
		if p.Burst < 1 || p.Arrival < 0 || p.Name == "" || seen[p.Name] {
			return fmt.Errorf("%w: %q needs a unique name, arrival >= 0 and burst >= 1", ErrInput, p.Name)
		}
		seen[p.Name] = true
		total += p.Burst
		if p.Arrival > MaxTime {
			return fmt.Errorf("%w: arrival above %d", ErrInput, MaxTime)
		}
	}
	if total > MaxTime {
		return fmt.Errorf("%w: more than %d ticks of work", ErrInput, MaxTime)
	}
	return nil
}

// Parse reads processes, one per line: "name arrival burst [priority]",
// separated by spaces, tabs or commas; a header line starting with a letter
// in the arrival column is skipped. Lines starting with # are comments.
func Parse(text string) ([]Process, error) {
	var ps []Process
	for i, line := range strings.Split(text, "\n") {
		f := strings.FieldsFunc(line, func(r rune) bool { return r == ' ' || r == '\t' || r == ',' || r == '\r' })
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		if len(f) < 3 || len(f) > 4 {
			return nil, fmt.Errorf("%w: line %d: want name arrival burst [priority]", ErrInput, i+1)
		}
		nums := make([]int, len(f)-1)
		bad := 0
		for j, s := range f[1:] {
			n, err := strconv.Atoi(s)
			if err != nil {
				bad++
			}
			nums[j] = n
		}
		if bad == len(nums) && len(ps) == 0 {
			continue // a header row: no number where numbers go
		}
		if bad > 0 {
			return nil, fmt.Errorf("%w: line %d: arrival, burst and priority must be whole numbers", ErrInput, i+1)
		}
		p := Process{Name: f[0], Arrival: nums[0], Burst: nums[1]}
		if len(nums) == 3 {
			p.Priority = nums[2]
		}
		ps = append(ps, p)
	}
	return ps, validate(ps)
}

// build turns a per-tick record of who ran (-1 idle) into a Schedule.
func build(ps []Process, ran []int, queue []int) Schedule {
	var s Schedule
	start := make([]int, len(ps))
	finish := make([]int, len(ps))
	for i := range start {
		start[i] = -1
	}
	for t, who := range ran {
		name, q := "", 0
		if who >= 0 {
			name = ps[who].Name
			if start[who] < 0 {
				start[who] = t
			}
			finish[who] = t + 1
			if queue != nil {
				q = queue[t]
			}
		}
		if n := len(s.Timeline); n > 0 && s.Timeline[n-1].Process == name && s.Timeline[n-1].Queue == q {
			s.Timeline[n-1].End = t + 1
			continue
		}
		if n := len(s.Timeline); n > 0 && name != "" {
			if prev := s.Timeline[n-1].Process; prev != "" && prev != name {
				s.ContextSwitches++
			}
		}
		s.Timeline = append(s.Timeline, Slice{Process: name, Start: t, End: t + 1, Queue: q})
	}
	s.Makespan = len(ran)
	for i, p := range ps {
		st := Stats{Process: p, Start: start[i], Finish: finish[i]}
		st.Turnaround = st.Finish - p.Arrival
		st.Waiting = st.Turnaround - p.Burst
		st.Response = st.Start - p.Arrival
		s.Stats = append(s.Stats, st)
		s.AvgWaiting += float64(st.Waiting)
		s.AvgTurnaround += float64(st.Turnaround)
		s.AvgResponse += float64(st.Response)
	}
	n := float64(len(ps))
	s.AvgWaiting /= n
	s.AvgTurnaround /= n
	s.AvgResponse /= n
	return s
}

// simulate runs a scheduler one tick at a time. pick chooses among the ready
// processes (indexes into ps, in arrival order) given which one ran last tick
// (-1 none) and the remaining bursts; it returns the index to run.
func simulate(ps []Process, pick func(ready []int, last int, remaining []int) int) Schedule {
	remaining := make([]int, len(ps))
	order := make([]int, len(ps)) // indexes sorted by arrival, ties by input order
	for i, p := range ps {
		remaining[i], order[i] = p.Burst, i
	}
	sort.SliceStable(order, func(a, b int) bool { return ps[order[a]].Arrival < ps[order[b]].Arrival })
	var ran []int
	left, last := len(ps), -1
	for t := 0; left > 0; t++ {
		var ready []int
		for _, i := range order {
			if ps[i].Arrival <= t && remaining[i] > 0 {
				ready = append(ready, i)
			}
		}
		if len(ready) == 0 {
			ran, last = append(ran, -1), -1
			continue
		}
		who := pick(ready, last, remaining)
		remaining[who]--
		if remaining[who] == 0 {
			left--
		}
		ran, last = append(ran, who), who
	}
	return build(ps, ran, nil)
}

func sortStable(s []int, less func(a, b int) bool) {
	sort.SliceStable(s, func(i, j int) bool { return less(s[i], s[j]) })
}
