package scheduling

import "slices"

// FCFS schedules first come, first served: run processes to completion in
// order of arrival. Simple and fair in order, but one long job makes everyone
// behind it wait (the convoy effect).
//
// Time: O(T * n) for T ticks of work. Space: O(T).
func FCFS(ps []Process) (Schedule, error) {
	if err := validate(ps); err != nil {
		return Schedule{}, err
	}
	return simulate(ps, func(ready []int, last int, remaining []int) int {
		if slices.Contains(ready, last) {
			return last // never preempt
		}
		return ready[0] // the earliest arrival
	}), nil
}

// SJF schedules shortest job first, without preemption: when the CPU frees up,
// run the ready process with the shortest burst to completion. It minimizes
// average waiting time among non-preemptive schedules, but long jobs can
// starve.
//
// Time: O(T * n). Space: O(T).
func SJF(ps []Process) (Schedule, error) {
	if err := validate(ps); err != nil {
		return Schedule{}, err
	}
	return simulate(ps, func(ready []int, last int, remaining []int) int {
		if slices.Contains(ready, last) {
			return last
		}
		return argmin(ready, func(i int) int { return ps[i].Burst })
	}), nil
}

// SRTF schedules shortest remaining time first: SJF with preemption, so a new
// arrival with less work left than the running process takes the CPU. It
// minimizes average waiting time of all schedules.
//
// Time: O(T * n). Space: O(T).
func SRTF(ps []Process) (Schedule, error) {
	if err := validate(ps); err != nil {
		return Schedule{}, err
	}
	return simulate(ps, func(ready []int, last int, remaining []int) int {
		best := argmin(ready, func(i int) int { return remaining[i] })
		if slices.Contains(ready, last) && remaining[last] == remaining[best] {
			return last // on a tie keep running: no needless switch
		}
		return best
	}), nil
}

// Priority schedules the ready process with the lowest Priority number. With
// preempt, a more urgent arrival takes the CPU at once; without, it waits
// until the running process finishes. Low-priority processes can starve.
//
// Time: O(T * n). Space: O(T).
func Priority(ps []Process, preempt bool) (Schedule, error) {
	if err := validate(ps); err != nil {
		return Schedule{}, err
	}
	return simulate(ps, func(ready []int, last int, remaining []int) int {
		best := argmin(ready, func(i int) int { return ps[i].Priority })
		if slices.Contains(ready, last) && (!preempt || ps[last].Priority <= ps[best].Priority) {
			return last
		}
		return best
	}), nil
}

// RoundRobin gives each ready process the CPU for at most quantum ticks in
// turn, sending it to the back of the queue if it is not done. Response time
// is good and nothing starves; a quantum too small spends the CPU switching, a
// quantum too large is FCFS. A new arrival joins the queue before the process
// whose quantum just ended, as most textbooks do.
//
// Time: O(T * n). Space: O(T + n).
func RoundRobin(ps []Process, quantum int) (Schedule, error) {
	if err := validate(ps); err != nil {
		return Schedule{}, err
	}
	if quantum < 1 {
		quantum = 1
	}
	remaining := make([]int, len(ps))
	for i, p := range ps {
		remaining[i] = p.Burst
	}
	arrived := make([]bool, len(ps))
	var queue, ran []int
	admit := func(t int) {
		for _, i := range byArrival(ps) {
			if !arrived[i] && ps[i].Arrival <= t {
				arrived[i] = true
				queue = append(queue, i)
			}
		}
	}
	left := len(ps)
	for t := 0; left > 0; {
		admit(t)
		if len(queue) == 0 {
			ran = append(ran, -1)
			t++
			continue
		}
		cur := queue[0]
		queue = queue[1:]
		for q := 0; q < quantum && remaining[cur] > 0; q++ {
			ran = append(ran, cur)
			remaining[cur]--
			t++
			admit(t)
		}
		if remaining[cur] > 0 {
			queue = append(queue, cur)
		} else {
			left--
		}
	}
	return build(ps, ran, nil), nil
}

// MLFQ schedules with a multilevel feedback queue: one round-robin queue per
// entry of quanta, most urgent first. A process starts in queue 0; using its
// whole quantum moves it down a level, so CPU-bound work sinks and short jobs
// finish quickly. The last queue keeps its quantum. Every boost ticks, every
// process returns to queue 0, so nothing starves (0 turns boosting off).
//
// Time: O(T * n). Space: O(T + n).
func MLFQ(ps []Process, quanta []int, boost int) (Schedule, error) {
	if err := validate(ps); err != nil {
		return Schedule{}, err
	}
	if len(quanta) == 0 {
		quanta = []int{2, 4, 8}
	}
	remaining := make([]int, len(ps))
	level := make([]int, len(ps))
	for i, p := range ps {
		remaining[i] = p.Burst
	}
	arrived := make([]bool, len(ps))
	queues := make([][]int, len(quanta))
	var ran, ranQueue []int
	admit := func(t int) {
		for _, i := range byArrival(ps) {
			if !arrived[i] && ps[i].Arrival <= t {
				arrived[i] = true
				queues[0] = append(queues[0], i)
			}
		}
	}
	left := len(ps)
	for t := 0; left > 0; {
		if boost > 0 && t > 0 && t%boost == 0 {
			for l := 1; l < len(queues); l++ {
				for _, i := range queues[l] {
					level[i] = 0
				}
				queues[0] = append(queues[0], queues[l]...)
				queues[l] = nil
			}
		}
		admit(t)
		l := 0
		for l < len(queues) && len(queues[l]) == 0 {
			l++
		}
		if l == len(queues) {
			ran, ranQueue = append(ran, -1), append(ranQueue, 0)
			t++
			continue
		}
		cur := queues[l][0]
		queues[l] = queues[l][1:]
		used := 0
		for used < quanta[l] && remaining[cur] > 0 {
			ran, ranQueue = append(ran, cur), append(ranQueue, l)
			remaining[cur]--
			used++
			t++
			admit(t)
		}
		switch {
		case remaining[cur] == 0:
			left--
		case used == quanta[l] && l+1 < len(queues):
			level[cur] = l + 1
			queues[l+1] = append(queues[l+1], cur)
		default:
			queues[l] = append(queues[l], cur)
		}
	}
	return build(ps, ran, ranQueue), nil
}

func byArrival(ps []Process) []int {
	order := make([]int, len(ps))
	for i := range order {
		order[i] = i
	}
	sortStable(order, func(a, b int) bool { return ps[a].Arrival < ps[b].Arrival })
	return order
}

// argmin returns the element of s with the least key, the first on a tie.
func argmin(s []int, key func(int) int) int {
	best := s[0]
	for _, i := range s[1:] {
		if key(i) < key(best) {
			best = i
		}
	}
	return best
}
