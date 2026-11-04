package greedy

import (
	"fmt"
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Interval is an activity that occupies [Start, End): it can share a resource
// with another only when one ends before (or as) the other starts.
type Interval struct {
	Name       string
	Start, End int
}

// ParseIntervals reads "name start end" lines (see parseLines); an interval
// must end after it starts.
func ParseIntervals(text string) ([]Interval, error) {
	names, nums, err := parseLines(text, 2, "name start end")
	if err != nil {
		return nil, err
	}
	out := make([]Interval, len(names))
	for i, n := range names {
		if nums[i][1] <= nums[i][0] {
			return nil, fmt.Errorf("%w: %s ends at %d, not after its start %d", ErrInput, n, nums[i][1], nums[i][0])
		}
		out[i] = Interval{Name: n, Start: nums[i][0], End: nums[i][1]}
	}
	return out, nil
}

// byEnd returns the indexes of acts ordered by end time, ties by start, then
// by input order.
func byEnd(acts []Interval) []int {
	order := make([]int, len(acts))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool {
		x, y := acts[order[a]], acts[order[b]]
		if x.End != y.End {
			return x.End < y.End
		}
		return x.Start < y.Start
	})
	return order
}

// ActivitySelection returns the indexes, in order of end time, of a largest
// set of activities that do not overlap: take the activity that ends first,
// drop everything overlapping it, repeat. Greedy is optimal here because the
// earliest-ending activity leaves the most room for the rest (an exchange
// argument: any optimal set can swap its first activity for it).
//
// Time: O(n log n) for the sort. Space: O(n).
func ActivitySelection(acts []Interval) []int { return activities(acts, nop) }

// TraceActivitySelection is ActivitySelection, reporting each candidate
// compared with the last one taken (Compare, indexes into acts) and each
// activity taken (Mark "chosen").
func TraceActivitySelection(acts []Interval, r trace.Recorder) []int {
	return activities(acts, rec{r})
}

func activities(acts []Interval, x rec) []int {
	var chosen []int
	last := -1
	for _, i := range byEnd(acts) {
		if last >= 0 {
			x.compare(i, last)
			if acts[i].Start < acts[last].End {
				continue
			}
		}
		chosen = append(chosen, i)
		x.chosen(i)
		last = i
	}
	return chosen
}
