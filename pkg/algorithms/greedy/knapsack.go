package greedy

import (
	"fmt"
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Item is something to carry: its weight and its value.
type Item struct {
	Name          string
	Weight, Value int
}

// ParseItems reads "name weight value" lines; weights must be positive and
// values not negative.
func ParseItems(text string) ([]Item, error) {
	names, nums, err := parseLines(text, 2, "name weight value")
	if err != nil {
		return nil, err
	}
	out := make([]Item, len(names))
	for i, n := range names {
		if nums[i][0] <= 0 || nums[i][1] < 0 {
			return nil, fmt.Errorf("%w: %s needs a positive weight and a value of at least 0", ErrInput, n)
		}
		out[i] = Item{Name: n, Weight: nums[i][0], Value: nums[i][1]}
	}
	return out, nil
}

// FractionalKnapsack fills a knapsack of capacity with the items, any
// fraction of each allowed, for the most value: take items by value per unit
// of weight, best first, and a fraction of the first one that does not fit.
// Greedy is optimal because any weight spent on a worse ratio could carry
// more value spent on a better one. (For the 0/1 knapsack, where items cannot
// be split, greedy is not optimal; that needs dynamic programming.) It returns
// the value carried and the fraction taken of each item, in input order.
//
// Time: O(n log n). Space: O(n).
func FractionalKnapsack(items []Item, capacity int) (float64, []float64) {
	return knapsack(items, capacity, nop)
}

// TraceFractionalKnapsack is FractionalKnapsack, reporting each item weighed
// (Compare) and each taken (Mark "chosen", Values[0] the percent taken).
func TraceFractionalKnapsack(items []Item, capacity int, r trace.Recorder) (float64, []float64) {
	return knapsack(items, capacity, rec{r})
}

func knapsack(items []Item, capacity int, x rec) (float64, []float64) {
	order := make([]int, len(items))
	for i := range order {
		order[i] = i
	}
	// Compare ratios without division: a.V/a.W > b.V/b.W iff a.V*b.W > b.V*a.W.
	sort.SliceStable(order, func(a, b int) bool {
		p, q := items[order[a]], items[order[b]]
		return int64(p.Value)*int64(q.Weight) > int64(q.Value)*int64(p.Weight)
	})
	taken := make([]float64, len(items))
	left, total := float64(max(0, capacity)), 0.0
	for _, i := range order {
		x.compare(i)
		if left <= 0 {
			break
		}
		w := float64(items[i].Weight)
		f := 1.0
		if w > left {
			f = left / w
		}
		taken[i] = f
		left -= f * w
		total += f * float64(items[i].Value)
		x.chosen(i, int(f*100+0.5))
	}
	return total, taken
}
