package graphs

import (
	"container/heap"
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// disjoint is a union-find over node indexes, for Kruskal's algorithm.
type disjoint struct{ parent, rank []int }

func newDisjoint(n int) *disjoint {
	d := &disjoint{parent: make([]int, n), rank: make([]int, n)}
	for i := range d.parent {
		d.parent[i] = i
	}
	return d
}

func (d *disjoint) find(x int) int {
	for d.parent[x] != x {
		d.parent[x] = d.parent[d.parent[x]] // path halving
		x = d.parent[x]
	}
	return x
}

func (d *disjoint) union(a, b int) bool {
	ra, rb := d.find(a), d.find(b)
	if ra == rb {
		return false
	}
	if d.rank[ra] < d.rank[rb] {
		ra, rb = rb, ra
	}
	d.parent[rb] = ra
	if d.rank[ra] == d.rank[rb] {
		d.rank[ra]++
	}
	return true
}

// Kruskal returns a minimum spanning forest of an undirected graph (one tree
// per connected component) as edge indexes, and its total weight, with
// Kruskal's algorithm (1956): take the edges lightest first, skipping any that
// would close a cycle. Ties go to the edge listed first.
//
// Time: O(E log E). Space: O(V + E).
func Kruskal(g *Graph) (edges []int, total int, err error) { return kruskal(g, nop) }

// TraceKruskal is Kruskal, reporting each edge considered (Mark "edge",
// Indices[0] the edge index) and each edge taken (Mark "tree").
func TraceKruskal(g *Graph, r trace.Recorder) (edges []int, total int, err error) {
	return kruskal(g, rec{r})
}

func kruskal(g *Graph, x rec) ([]int, int, error) {
	if g.Directed {
		return nil, 0, ErrDirected
	}
	order := make([]int, len(g.Edges))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool { return g.Edges[order[a]].Weight < g.Edges[order[b]].Weight })
	d := newDisjoint(g.Len())
	var tree []int
	total := 0
	for _, i := range order {
		e := g.Edges[i]
		x.mark("edge", []int{i})
		if d.union(e.From, e.To) {
			tree = append(tree, i)
			total += e.Weight
			x.mark("tree", []int{i})
		}
	}
	return tree, total, nil
}

// Prim returns a minimum spanning forest of an undirected graph as edge
// indexes, and its total weight, with Prim's algorithm (Jarnik 1930, Prim
// 1957): grow a tree from a node, always adding the lightest edge to a node
// not yet in it; start a new tree from the next node outside when one runs
// out.
//
// Time: O(E log V) with a binary heap. Space: O(V + E).
func Prim(g *Graph) (edges []int, total int, err error) { return prim(g, nop) }

// TracePrim is Prim, reporting each node added (Visit) and each edge taken
// (Mark "tree").
func TracePrim(g *Graph, r trace.Recorder) (edges []int, total int, err error) {
	return prim(g, rec{r})
}

type edgeItem struct{ weight, edge, to, order int }

type edgeQueue []edgeItem

func (q edgeQueue) Len() int { return len(q) }
func (q edgeQueue) Less(i, j int) bool {
	if q[i].weight != q[j].weight {
		return q[i].weight < q[j].weight
	}
	return q[i].order < q[j].order
}
func (q edgeQueue) Swap(i, j int) { q[i], q[j] = q[j], q[i] }
func (q *edgeQueue) Push(x any)   { *q = append(*q, x.(edgeItem)) }
func (q *edgeQueue) Pop() any {
	old := *q
	it := old[len(old)-1]
	*q = old[:len(old)-1]
	return it
}

func prim(g *Graph, x rec) ([]int, int, error) {
	if g.Directed {
		return nil, 0, ErrDirected
	}
	adj := g.adjacency()
	in := make([]bool, g.Len())
	var tree []int
	total, pushed := 0, 0
	q := &edgeQueue{}
	add := func(u int) {
		in[u] = true
		x.visit(u)
		for _, a := range adj[u] {
			if !in[a.to] {
				heap.Push(q, edgeItem{a.weight, a.edge, a.to, pushed})
				pushed++
			}
		}
	}
	for start := 0; start < g.Len(); start++ {
		if in[start] {
			continue
		}
		add(start)
		for q.Len() > 0 {
			it := heap.Pop(q).(edgeItem)
			if in[it.to] {
				continue
			}
			tree = append(tree, it.edge)
			total += it.weight
			x.mark("tree", []int{it.edge})
			add(it.to)
		}
	}
	return tree, total, nil
}
