package scheduling

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Task is a periodic real-time task: every Period ticks a job is released that
// needs Burst ticks of CPU before the next release (its deadline).
type Task struct {
	Name   string `json:"name"`
	Period int    `json:"period"`
	Burst  int    `json:"burst"`
}

// Miss is a job that was not finished by its deadline.
type Miss struct {
	Task     string `json:"task"`
	Release  int    `json:"release"`
	Deadline int    `json:"deadline"`
	Left     int    `json:"left"` // ticks of work still to do at the deadline
}

// RealTime is the result of a real-time schedule over one hyperperiod (the
// least common multiple of the periods, at most MaxTime).
type RealTime struct {
	Timeline []Slice `json:"timeline"`
	Missed   []Miss  `json:"missed"`
	// Utilization is the sum of Burst/Period: above 1 no schedule can meet
	// every deadline.
	Utilization float64 `json:"utilization"`
	// Bound is the Liu-Layland bound n(2^(1/n) - 1): rate-monotonic meets
	// every deadline when Utilization is at most Bound (a sufficient test).
	Bound   float64 `json:"bound"`
	Horizon int     `json:"horizon"`
}

// ParseTasks reads periodic tasks, one a line: "name period burst". A header
// line and lines starting with # are skipped.
func ParseTasks(text string) ([]Task, error) {
	var ts []Task
	seen := map[string]bool{}
	for i, line := range strings.Split(text, "\n") {
		f := strings.FieldsFunc(line, func(r rune) bool { return r == ' ' || r == '\t' || r == ',' || r == '\r' })
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		if len(f) != 3 {
			return nil, fmt.Errorf("%w: line %d: want name period burst", ErrInput, i+1)
		}
		p, err1 := strconv.Atoi(f[1])
		b, err2 := strconv.Atoi(f[2])
		if err1 != nil && err2 != nil && len(ts) == 0 {
			continue // a header row
		}
		if err1 != nil || err2 != nil || p < 1 || b < 1 || b > p || seen[f[0]] {
			return nil, fmt.Errorf("%w: line %d: a unique name, period >= 1 and 1 <= burst <= period", ErrInput, i+1)
		}
		seen[f[0]] = true
		ts = append(ts, Task{Name: f[0], Period: p, Burst: b})
	}
	if len(ts) == 0 || len(ts) > MaxProcesses {
		return nil, fmt.Errorf("%w: give 1 to %d tasks", ErrInput, MaxProcesses)
	}
	return ts, nil
}

// EDF schedules periodic tasks by earliest deadline first: each tick, run the
// released job whose deadline is nearest. On one CPU it meets every deadline
// whenever Utilization is at most 1, so it is optimal; a missed deadline means
// the task set is overloaded.
//
// Time: O(H * n) for a hyperperiod of H ticks. Space: O(H).
func EDF(ts []Task) (RealTime, error) {
	return realTime(ts, func(j job) int { return j.deadline })
}

// RateMonotonic schedules periodic tasks by fixed priority: the shorter the
// period, the higher the priority (Liu and Layland, 1973). It is the optimal
// fixed-priority scheme, but can miss deadlines EDF would meet when
// Utilization is above Bound.
//
// Time: O(H * n). Space: O(H).
func RateMonotonic(ts []Task) (RealTime, error) {
	return realTime(ts, func(j job) int { return ts[j.task].Period })
}

type job struct {
	task, release, deadline, left int
}

func realTime(ts []Task, key func(job) int) (RealTime, error) {
	if len(ts) == 0 || len(ts) > MaxProcesses {
		return RealTime{}, fmt.Errorf("%w: give 1 to %d tasks", ErrInput, MaxProcesses)
	}
	h := 1
	for _, t := range ts {
		if t.Period < 1 || t.Burst < 1 || t.Burst > t.Period {
			return RealTime{}, fmt.Errorf("%w: %s needs period >= 1 and 1 <= burst <= period", ErrInput, t.Name)
		}
		h = lcm(h, t.Period)
		if h > MaxTime {
			h = MaxTime
			break
		}
	}
	rt := RealTime{Horizon: h}
	for _, t := range ts {
		rt.Utilization += float64(t.Burst) / float64(t.Period)
	}
	n := float64(len(ts))
	rt.Bound = n * (math.Pow(2, 1/n) - 1)

	var jobs []job
	ran := make([]int, 0, h)
	for t := 0; t < h; t++ {
		for i, task := range ts {
			if t%task.Period == 0 {
				jobs = append(jobs, job{task: i, release: t, deadline: t + task.Period, left: task.Burst})
			}
		}
		// Jobs whose deadline has come are missed, whatever is left of them.
		kept := jobs[:0]
		for _, j := range jobs {
			if j.left > 0 && j.deadline <= t {
				rt.Missed = append(rt.Missed, Miss{Task: ts[j.task].Name, Release: j.release, Deadline: j.deadline, Left: j.left})
				continue
			}
			if j.left > 0 {
				kept = append(kept, j)
			}
		}
		jobs = kept
		best := -1
		for k, j := range jobs {
			if best < 0 || key(j) < key(jobs[best]) || (key(j) == key(jobs[best]) && j.release < jobs[best].release) {
				best = k
			}
		}
		if best < 0 {
			ran = append(ran, -1)
			continue
		}
		jobs[best].left--
		ran = append(ran, jobs[best].task)
	}
	for _, j := range jobs { // jobs due exactly at the horizon
		if j.left > 0 && j.deadline <= h {
			rt.Missed = append(rt.Missed, Miss{Task: ts[j.task].Name, Release: j.release, Deadline: j.deadline, Left: j.left})
		}
	}
	for t, who := range ran {
		name := ""
		if who >= 0 {
			name = ts[who].Name
		}
		if n := len(rt.Timeline); n > 0 && rt.Timeline[n-1].Process == name {
			rt.Timeline[n-1].End = t + 1
			continue
		}
		rt.Timeline = append(rt.Timeline, Slice{Process: name, Start: t, End: t + 1})
	}
	return rt, nil
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func lcm(a, b int) int { return a / gcd(a, b) * b }
