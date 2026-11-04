package greedy

import (
	"fmt"
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Job is a task taking one unit of time that earns Profit if it is finished
// by Deadline (in units of time, from 1).
type Job struct {
	Name             string
	Deadline, Profit int
}

// ParseJobs reads "name deadline profit" lines; deadlines must be at least 1.
func ParseJobs(text string) ([]Job, error) {
	names, nums, err := parseLines(text, 2, "name deadline profit")
	if err != nil {
		return nil, err
	}
	out := make([]Job, len(names))
	for i, n := range names {
		if nums[i][0] < 1 || nums[i][0] > MaxItems {
			return nil, fmt.Errorf("%w: %s needs a deadline from 1 to %d", ErrInput, n, MaxItems)
		}
		out[i] = Job{Name: n, Deadline: nums[i][0], Profit: nums[i][1]}
	}
	return out, nil
}

// JobSequencing schedules unit-time jobs for the most profit: take jobs by
// profit, highest first, each in the latest free slot before its deadline,
// skipping a job when every such slot is taken. Greedy is optimal: placing a
// job as late as possible leaves the earlier slots for jobs with earlier
// deadlines (the jobs that fit form a matroid). It returns, for each time slot
// 1..maximum deadline, the index of the job run in it or -1, and the profit.
//
// Time: O(n log n + n * d) for the largest deadline d. Space: O(d).
func JobSequencing(jobs []Job) ([]int, int) { return sequence(jobs, nop) }

// TraceJobSequencing is JobSequencing, reporting each job weighed (Compare)
// and each placed (Mark "chosen", Values[0] the slot, from 0).
func TraceJobSequencing(jobs []Job, r trace.Recorder) ([]int, int) { return sequence(jobs, rec{r}) }

func sequence(jobs []Job, x rec) ([]int, int) {
	order := make([]int, len(jobs))
	maxD := 0
	for i := range order {
		order[i] = i
		maxD = max(maxD, jobs[i].Deadline)
	}
	sort.SliceStable(order, func(a, b int) bool { return jobs[order[a]].Profit > jobs[order[b]].Profit })
	slots := make([]int, maxD)
	for i := range slots {
		slots[i] = -1
	}
	profit := 0
	for _, i := range order {
		x.compare(i)
		for s := jobs[i].Deadline - 1; s >= 0; s-- {
			if slots[s] < 0 {
				slots[s] = i
				profit += jobs[i].Profit
				x.chosen(i, s)
				break
			}
		}
	}
	return slots, profit
}
