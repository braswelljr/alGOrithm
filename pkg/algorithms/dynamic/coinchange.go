package dynamic

import (
	"errors"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// MaxAmount is the largest amount CoinChange accepts.
const MaxAmount = 100_000

// ErrAmount is returned by CoinChange for a negative amount, an amount above MaxAmount, or a
// coin that is not positive.
var ErrAmount = errors.New("dynamic: the amount must be 0 to 100000 and every coin positive")

// CoinChange returns the fewest coins that sum to amount, the coins used (largest
// first), and the table: table[a] is the fewest coins for a, or -1 when a
// cannot be made. It returns -1 and no coins when amount cannot be made.
//
// Time: O(amount * len(coins)).
// Space: O(amount).
func CoinChange(coins []int, amount int) (int, []int, []int, error) {
	return coinChange(coins, amount, trace.Nop{})
}

// TraceCoinChange is CoinChange, reporting each table cell read (Compare) and written
// (Set, Indices[0] is the amount), and the amounts on the way to the answer
// (Mark "path").
func TraceCoinChange(coins []int, amount int, r trace.Recorder) (int, []int, []int, error) {
	return coinChange(coins, amount, r)
}

func coinChange(coins []int, amount int, r trace.Recorder) (int, []int, []int, error) {
	if amount < 0 || amount > MaxAmount {
		return 0, nil, nil, ErrAmount
	}
	for _, c := range coins {
		if c <= 0 {
			return 0, nil, nil, ErrAmount
		}
	}
	const none = -1
	t := make([]int, amount+1)
	last := make([]int, amount+1) // the coin that made each amount
	r.Record(trace.Event{Kind: trace.Set, Indices: []int{0}, Values: []int{0}})
	for a := 1; a <= amount; a++ {
		t[a], last[a] = none, 0
		for _, c := range coins {
			if c <= a && t[a-c] != none {
				r.Record(trace.Event{Kind: trace.Compare, Indices: []int{a - c, a}})
				if t[a] == none || t[a-c]+1 < t[a] {
					t[a], last[a] = t[a-c]+1, c
				}
			}
		}
		r.Record(trace.Event{Kind: trace.Set, Indices: []int{a}, Values: []int{t[a]}})
	}
	if t[amount] == none {
		return -1, nil, t, nil
	}
	var used, path []int
	for a := amount; a > 0; a -= last[a] {
		path = append(path, a)
		used = append(used, last[a])
	}
	path = append(path, 0)
	r.Record(trace.Event{Kind: trace.Mark, Indices: path, Note: "path"})
	return t[amount], used, t, nil
}
