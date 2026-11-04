package graphs

import (
	"container/heap"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Dijkstra returns the shortest distance from source to every node
// (Infinity where unreachable) and prev, each node's predecessor on its
// shortest path (-1 for the source and unreached nodes), with Dijkstra's
// algorithm (1956): always settle the closest unsettled node. It needs
// non-negative weights and returns ErrNegative otherwise.
//
// Time: O((V + E) log V) with a binary heap. Space: O(V).
func Dijkstra(g *Graph, source int) (dist, prev []int, err error) { return dijkstra(g, source, nop) }

// TraceDijkstra is Dijkstra, reporting each node queued (Push), settled
// (Visit), and each distance lowered (Set).
func TraceDijkstra(g *Graph, source int, r trace.Recorder) (dist, prev []int, err error) {
	return dijkstra(g, source, rec{r})
}

type item struct{ node, dist, order int }

type minQueue []item

func (q minQueue) Len() int { return len(q) }
func (q minQueue) Less(i, j int) bool {
	if q[i].dist != q[j].dist {
		return q[i].dist < q[j].dist
	}
	return q[i].order < q[j].order
}
func (q minQueue) Swap(i, j int) { q[i], q[j] = q[j], q[i] }
func (q *minQueue) Push(x any)   { *q = append(*q, x.(item)) }
func (q *minQueue) Pop() any {
	old := *q
	it := old[len(old)-1]
	*q = old[:len(old)-1]
	return it
}

func dijkstra(g *Graph, source int, x rec) ([]int, []int, error) {
	if err := g.checkNode(source); err != nil {
		return nil, nil, err
	}
	for _, e := range g.Edges {
		if e.Weight < 0 {
			return nil, nil, ErrNegative
		}
	}
	adj := g.adjacency()
	dist, prev := filled(g.Len(), Infinity), filled(g.Len(), -1)
	done := make([]bool, g.Len())
	dist[source] = 0
	x.set(source, 0)
	q := &minQueue{{source, 0, 0}}
	x.push(source)
	pushed := 1
	for q.Len() > 0 {
		it := heap.Pop(q).(item)
		u := it.node
		if done[u] {
			continue
		}
		done[u] = true
		x.visit(u)
		for _, a := range adj[u] {
			if d := dist[u] + a.weight; !done[a.to] && d < dist[a.to] {
				dist[a.to], prev[a.to] = d, u
				x.set(a.to, d)
				heap.Push(q, item{a.to, d, pushed})
				pushed++
				x.push(a.to)
			}
		}
	}
	return dist, prev, nil
}
