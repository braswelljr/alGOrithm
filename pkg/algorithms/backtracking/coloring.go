package backtracking

import (
	"fmt"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Graph is an undirected graph: node names and, for each node, its
// neighbors.
type Graph struct {
	Nodes []string
	Adj   [][]int
}

// MaxGraphNodes bounds the graphs ParseGraph reads.
const MaxGraphNodes = 200

// ParseGraph reads an undirected graph, one edge a line, "A B" (a line with
// one name adds a lone node). Nodes are numbered in order of appearance.
func ParseGraph(text string) (Graph, error) {
	var g Graph
	index := map[string]int{}
	node := func(name string) int {
		if i, ok := index[name]; ok {
			return i
		}
		index[name] = len(g.Nodes)
		g.Nodes = append(g.Nodes, name)
		g.Adj = append(g.Adj, nil)
		return len(g.Nodes) - 1
	}
	for i, line := range strings.Split(text, "\n") {
		f := strings.FieldsFunc(line, func(r rune) bool { return r == ' ' || r == '\t' || r == ',' || r == '-' || r == '\r' })
		switch {
		case len(f) == 0 || strings.HasPrefix(f[0], "#"):
		case len(f) == 1:
			node(f[0])
		case len(f) == 2:
			a, b := node(f[0]), node(f[1])
			if a == b {
				return Graph{}, fmt.Errorf("%w: line %d: %s cannot be its own neighbor", ErrInput, i+1, f[0])
			}
			g.Adj[a], g.Adj[b] = append(g.Adj[a], b), append(g.Adj[b], a)
		default:
			return Graph{}, fmt.Errorf("%w: line %d: want one edge, \"A B\"", ErrInput, i+1)
		}
		if len(g.Nodes) > MaxGraphNodes {
			return Graph{}, fmt.Errorf("%w: at most %d nodes", ErrInput, MaxGraphNodes)
		}
	}
	if len(g.Nodes) == 0 {
		return Graph{}, fmt.Errorf("%w: no nodes", ErrInput)
	}
	return g, nil
}

// GraphColoring gives each node one of m colors (0 to m-1) so no two
// neighbors share one: color the nodes in order, each with the first color
// its colored neighbors do not use, and back up when there is none. It
// returns the colors and true, or nil and false when m colors are not enough.
//
// Time: O(m^n) in the worst case. Space: O(n).
func GraphColoring(g Graph, m int) ([]int, bool) { return coloring(g, m, nop) }

// TraceGraphColoring is GraphColoring, reporting each color given (Mark
// "group", Indices[0] the node, Values[0] the color; -1 when it is taken
// back) and each clash with a neighbor (Compare: the node and the neighbor).
func TraceGraphColoring(g Graph, m int, r trace.Recorder) ([]int, bool) {
	return coloring(g, m, rec{r})
}

func coloring(g Graph, m int, x rec) ([]int, bool) {
	color := make([]int, len(g.Nodes))
	for i := range color {
		color[i] = -1
	}
	group := func(n, c int) {
		x.r.Record(trace.Event{Kind: trace.Mark, Indices: []int{n}, Values: []int{c}, Note: "group"})
	}
	var paint func(n int) bool
	paint = func(n int) bool {
		if n == len(g.Nodes) {
			return true
		}
	colors:
		for c := 0; c < m; c++ {
			for _, nb := range g.Adj[n] {
				if color[nb] == c {
					x.compare(n, nb)
					continue colors
				}
			}
			color[n] = c
			group(n, c)
			if paint(n + 1) {
				return true
			}
			color[n] = -1
			group(n, -1)
		}
		return false
	}
	if m < 1 || !paint(0) {
		return nil, false
	}
	return color, true
}
