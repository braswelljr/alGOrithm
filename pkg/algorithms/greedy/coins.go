package greedy

import (
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// CoinChange makes amount from coins (any number of each) the greedy way:
// the largest coin that fits, again and again. It returns the coins used,
// largest first, and false when what is left cannot be paid. Greedy is optimal
// for canonical systems like 1, 5, 10, 25, but not in general: with coins 1, 3
// and 4, it pays 6 as 4+1+1 where 3+3 needs one coin fewer (see
// dynamic.CoinChange for the always-optimal answer).
//
// Time: O(k log k + amount / smallest coin) for k coins. Space: O(k).
func CoinChange(coins []int, amount int) ([]int, bool) { return coinChange(coins, amount, nop) }

// TraceCoinChange is CoinChange, reporting each coin tried (Compare, Indices[0]
// its index in coins) and each taken (Mark "chosen").
func TraceCoinChange(coins []int, amount int, r trace.Recorder) ([]int, bool) {
	return coinChange(coins, amount, rec{r})
}

func coinChange(coins []int, amount int, x rec) ([]int, bool) {
	order := make([]int, 0, len(coins))
	for i, c := range coins {
		if c > 0 {
			order = append(order, i)
		}
	}
	sort.SliceStable(order, func(a, b int) bool { return coins[order[a]] > coins[order[b]] })
	var used []int
	for _, i := range order {
		x.compare(i)
		for amount >= coins[i] {
			amount -= coins[i]
			used = append(used, coins[i])
			x.chosen(i)
		}
	}
	return used, amount == 0
}
