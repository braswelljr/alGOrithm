package randomized

import "math/rand"

// KargerMinCut estimates the minimum cut of an undirected multigraph with
// nodes 0..n-1 (the fewest edges whose removal splits it in two) with David
// Karger's contraction algorithm (1993): while more than two groups remain,
// pick a random edge and merge its two ends' groups; the edges left between
// the last two groups are a cut. One trial finds a minimum cut with
// probability at least 2 / (n(n-1)), so it runs trials independent trials
// and keeps the smallest: about n^2 ln n trials make failure unlikely. A
// Monte Carlo algorithm: the answer may be too large, never too small. It
// returns the cut size and which nodes are on the first node's side. A
// disconnected graph has cut 0. ErrInput for n < 2 or an edge naming a node
// outside 0..n-1; self-loops are ignored.
//
// Time: O(trials * m * alpha(n)) with union-find. Space: O(n + m).
func KargerMinCut(n int, edges [][2]int, trials int, seed int64) (cut int, side []bool, err error) {
	if n < 2 {
		return 0, nil, ErrInput
	}
	for _, e := range edges {
		if e[0] < 0 || e[0] >= n || e[1] < 0 || e[1] >= n {
			return 0, nil, ErrInput
		}
	}
	rng := rand.New(rand.NewSource(seed))
	best := -1
	for t := 0; t < max(trials, 1); t++ {
		parent := make([]int, n)
		for i := range parent {
			parent[i] = i
		}
		find := func(x int) int {
			for parent[x] != x {
				parent[x] = parent[parent[x]]
				x = parent[x]
			}
			return x
		}
		groups := n
		order := rng.Perm(len(edges)) // contracting edges in a random order is the same as picking random edges
		for _, k := range order {
			if groups == 2 {
				break
			}
			a, b := find(edges[k][0]), find(edges[k][1])
			if a != b {
				parent[a] = b
				groups--
			}
		}
		if groups > 2 { // disconnected: one component against the rest crosses no edge
			root, other := find(0), -1
			for i := range parent {
				switch r := find(i); {
				case r == root:
				case other < 0:
					other = r
				case r != other:
					parent[r] = other
				}
			}
		}
		c := 0
		for _, e := range edges {
			if find(e[0]) != find(e[1]) {
				c++
			}
		}
		if best < 0 || c < best {
			best = c
			root := find(0)
			side = make([]bool, n)
			for i := range side {
				side[i] = find(i) == root
			}
		}
	}
	return best, side, nil
}
