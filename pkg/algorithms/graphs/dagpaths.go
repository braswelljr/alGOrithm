package graphs

import "github.com/braswelljr/alGOrithm/pkg/trace"

// DAGShortestPaths returns the shortest distance from source to every node of
// a directed acyclic graph (Infinity where unreachable) and each node's
// predecessor, by relaxing edges in topological order: each node is final
// when it is reached, so negative weights are fine and it runs in linear
// time. A graph with a cycle returns ErrCycle.
//
// Time: O(V + E). Space: O(V).
func DAGShortestPaths(g *Graph, source int) (dist, prev []int, err error) {
	return dagRelax(g, source, false, nop)
}

// TraceDAGShortestPaths is DAGShortestPaths, reporting each node taken in
// topological order (Visit) and each distance lowered (Set).
func TraceDAGShortestPaths(g *Graph, source int, r trace.Recorder) (dist, prev []int, err error) {
	return dagRelax(g, source, false, rec{r})
}

// DAGLongestPath returns the heaviest path anywhere in a directed acyclic
// graph and its weight: the critical path of a project whose edges are tasks
// with durations, which no amount of parallel work can shorten. It is the
// same topological relaxation, maximizing from every node at once.
//
// Time: O(V + E). Space: O(V).
func DAGLongestPath(g *Graph) (path []int, length int, err error) {
	return dagLongest(g, nop)
}

// TraceDAGLongestPath is DAGLongestPath, reporting what TraceDAGShortestPaths
// does and the path (Mark "path").
func TraceDAGLongestPath(g *Graph, r trace.Recorder) (path []int, length int, err error) {
	return dagLongest(g, rec{r})
}

func dagLongest(g *Graph, x rec) ([]int, int, error) {
	dist, prev, err := dagRelax(g, -1, true, x)
	if err != nil {
		return nil, 0, err
	}
	end := 0
	for v, d := range dist {
		if d > dist[end] {
			end = v
		}
	}
	var path []int
	for v := end; v >= 0; v = prev[v] {
		path = append(path, v)
	}
	path = reversed(path)
	x.mark("path", path)
	return path, dist[end], nil
}

// dagRelax relaxes every edge in topological order: shortest from source, or
// longest starting anywhere (every node at 0) when longest is set.
func dagRelax(g *Graph, source int, longest bool, x rec) ([]int, []int, error) {
	if !longest {
		if err := g.checkNode(source); err != nil {
			return nil, nil, err
		}
	}
	order, err := topological(g, nop)
	if err != nil {
		return nil, nil, err
	}
	adj := g.adjacency()
	dist, prev := filled(g.Len(), Infinity), filled(g.Len(), -1)
	if longest {
		dist = make([]int, g.Len())
	} else {
		dist[source] = 0
		x.set(source, 0)
	}
	for _, u := range order {
		x.visit(u)
		if dist[u] == Infinity {
			continue
		}
		for _, a := range adj[u] {
			d := dist[u] + a.weight
			if (longest && d > dist[a.to]) || (!longest && d < dist[a.to]) {
				dist[a.to], prev[a.to] = d, u
				x.set(a.to, d)
			}
		}
	}
	return dist, prev, nil
}
