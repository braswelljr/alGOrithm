package graphs

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Limits of the exact, exponential searches.
const (
	MaxHamiltonian = 20        // nodes HamiltonianPath accepts
	MaxHeldKarp    = 15        // nodes HeldKarp accepts
	hamiltonSteps  = 2_000_000 // nodes tried before HamiltonianPath gives up
)

// HamiltonianPath returns a path that visits every node exactly once (with
// cycle, a cycle back to its start, which is then repeated at the end), by
// backtracking: extend the path with each unvisited neighbor in turn and undo
// the step when it leads nowhere. The problem is NP-complete, so the search
// is limited to MaxHamiltonian nodes and gives up with ErrLimit after a
// budget of steps. It returns nil when there is none.
//
// Time: O(n!) in the worst case. Space: O(n).
func HamiltonianPath(g *Graph, cycle bool) ([]int, error) { return hamiltonian(g, cycle, nop) }

// TraceHamiltonianPath is HamiltonianPath, reporting each node added to the
// path (Push) and each one taken back (Pop), and the path found (Mark
// "path").
func TraceHamiltonianPath(g *Graph, cycle bool, r trace.Recorder) ([]int, error) {
	return hamiltonian(g, cycle, rec{r})
}

func hamiltonian(g *Graph, cycle bool, x rec) ([]int, error) {
	n := g.Len()
	if n > MaxHamiltonian {
		return nil, fmt.Errorf("%w: at most %d", ErrTooManyNodes, MaxHamiltonian)
	}
	link := make([][]bool, n)
	for i := range link {
		link[i] = make([]bool, n)
	}
	for _, a := range g.Edges {
		link[a.From][a.To] = true
		if !g.Directed {
			link[a.To][a.From] = true
		}
	}
	inPath := make([]bool, n)
	path := make([]int, 0, n+1)
	steps := 0
	var extend func(u int) (bool, error)
	extend = func(u int) (bool, error) {
		if steps++; steps > hamiltonSteps {
			return false, ErrLimit
		}
		path = append(path, u)
		inPath[u] = true
		x.push(u)
		if len(path) == n && (!cycle || link[u][path[0]]) {
			return true, nil
		}
		for v := 0; v < n; v++ {
			if link[u][v] && !inPath[v] {
				if ok, err := extend(v); ok || err != nil {
					return ok, err
				}
			}
		}
		inPath[u] = false
		path = path[:len(path)-1]
		x.r.Record(trace.Event{Kind: trace.Pop, Indices: []int{u}})
		return false, nil
	}
	starts := n
	if cycle {
		starts = min(n, 1) // a cycle can start anywhere
	}
	for s := 0; s < starts; s++ {
		ok, err := extend(s)
		if err != nil {
			return nil, err
		}
		if ok {
			if cycle {
				path = append(path, path[0])
			}
			x.mark("path", path)
			return path, nil
		}
	}
	return nil, nil
}
