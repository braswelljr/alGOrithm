package graphs

import (
	"slices"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// KCore returns each node's core number: the largest k such that the node
// belongs to a k-core, a part of the graph where every node has at least k
// neighbors inside it. It peels the graph (Matula and Beck, 1983): repeatedly
// remove a node of least remaining degree; the degree when it goes, never
// lower than earlier removals', is its core number. Edge directions,
// self-loops and repeated edges are ignored.
//
// Time: O(V + E). Space: O(V + E).
func KCore(g *Graph) []int { return kCore(g, nop) }

// TraceKCore is KCore, reporting each node removed (Visit) with its core
// number (Mark "group", Values[0] the core).
func TraceKCore(g *Graph, r trace.Recorder) []int { return kCore(g, rec{r}) }

func kCore(g *Graph, x rec) []int {
	n := g.Len()
	nbrs := make([]map[int]bool, n)
	for i := range nbrs {
		nbrs[i] = map[int]bool{}
	}
	for _, e := range g.Edges {
		if e.From != e.To {
			nbrs[e.From][e.To], nbrs[e.To][e.From] = true, true
		}
	}
	deg := make([]int, n)
	maxDeg := 0
	for v := range nbrs {
		deg[v] = len(nbrs[v])
		maxDeg = max(maxDeg, deg[v])
	}
	buckets := make([][]int, maxDeg+1)
	for v, d := range deg {
		buckets[d] = append(buckets[d], v)
	}
	core, removed := make([]int, n), make([]bool, n)
	k := 0
	for left := n; left > 0; {
		d := 0
		for d < len(buckets) && len(buckets[d]) == 0 {
			d++
		}
		v := buckets[d][len(buckets[d])-1]
		buckets[d] = buckets[d][:len(buckets[d])-1]
		if removed[v] || deg[v] != d {
			continue // a stale entry: v moved to a lower bucket since
		}
		removed[v] = true
		left--
		k = max(k, d)
		core[v] = k
		x.visit(v)
		x.mark("group", []int{v}, k)
		// In order, so the trace is the same on every run (map order is not).
		for _, u := range sortedKeys(nbrs[v]) {
			if !removed[u] {
				deg[u]--
				buckets[deg[u]] = append(buckets[deg[u]], u)
			}
		}
	}
	return core
}

// sortedKeys returns the keys of m in increasing order.
func sortedKeys(m map[int]bool) []int {
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}
