package distributed

import (
	"fmt"
	"math/rand"
)

// Gossip spreads a rumor from node 0 through n nodes in rounds, seeded. In
// push gossip every node that knows it tells fanout others picked at random;
// with pull true, every node that does not know it also asks one random node
// each round, which finishes the last few much faster (the last uninformed
// nodes are hard to hit by pushing, easy to find by asking). It returns how
// many nodes know the rumor after each round, starting with 1; the number of
// rounds grows like log n. Messages are never lost and every node is up.
//
// Time: O(rounds * n * fanout). Space: O(n).
func Gossip(n, fanout int, pull bool, seed int64) ([]int, error) {
	if n < 2 || n > 100_000 || fanout < 1 || fanout >= n {
		return nil, fmt.Errorf("%w: 2 to 100000 nodes and a fanout of 1 to n-1", ErrInput)
	}
	r := rand.New(rand.NewSource(seed))
	knows := make([]bool, n)
	knows[0] = true
	count := 1
	out := []int{1}
	other := func(i int) int { // a random node other than i
		j := r.Intn(n - 1)
		if j >= i {
			j++
		}
		return j
	}
	for round := 0; count < n && round < 1000; round++ {
		next := append([]bool(nil), knows...)
		for i := 0; i < n; i++ {
			if knows[i] {
				for f := 0; f < fanout; f++ {
					next[other(i)] = true
				}
			} else if pull && knows[other(i)] {
				next[i] = true
			}
		}
		knows, count = next, 0
		for _, k := range knows {
			if k {
				count++
			}
		}
		out = append(out, count)
	}
	return out, nil
}
