package graphs

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Errors of the searches, tours and paths in this file's neighbors.
var (
	ErrNoEuler      = errors.New("graphs: no Euler path")
	ErrNoTour       = errors.New("graphs: no tour visits every node")
	ErrNotBipartite = errors.New("graphs: the graph is not bipartite")
	ErrLimit        = errors.New("graphs: the search gave up after too many steps")
	ErrTooManyNodes = errors.New("graphs: too many nodes for an exact search")
	ErrDisconnected = errors.New("graphs: the graph is not connected")
	errNeedTwoNodes = fmt.Errorf("%w: at least two nodes", ErrInput)
)

// Positions places nodes in the plane, for the heuristic of AStarGraph and
// BeamSearch: Positions[node] is its x and y.
type Positions map[int][2]int

// ParsePositions reads node positions from comment lines of a graph's text,
// "# pos NAME X Y", so a graph and its layout travel together and Parse,
// which skips comments, still reads the graph. Other lines are ignored.
func ParsePositions(g *Graph, text string) (Positions, error) {
	pos := Positions{}
	for i, line := range strings.Split(text, "\n") {
		f := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "#"))
		if !strings.HasPrefix(strings.TrimSpace(line), "#") || len(f) == 0 || !strings.EqualFold(f[0], "pos") {
			continue
		}
		if len(f) != 4 {
			return nil, fmt.Errorf("%w: line %d: want \"# pos NAME X Y\"", ErrInput, i+1)
		}
		n, ok := g.Lookup(f[1])
		if !ok {
			return nil, fmt.Errorf("%w: line %d: %q", ErrNode, i+1, f[1])
		}
		x, err1 := strconv.Atoi(f[2])
		y, err2 := strconv.Atoi(f[3])
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("%w: line %d: the position must be two whole numbers", ErrInput, i+1)
		}
		pos[n] = [2]int{x, y}
	}
	return pos, nil
}

// estimate is the straight-line distance from a to b, rounded down, or 0 when
// either has no position. It never overestimates when every edge weighs at
// least the straight-line length between its ends.
func (p Positions) estimate(a, b int) int {
	pa, okA := p[a]
	pb, okB := p[b]
	if !okA || !okB {
		return 0
	}
	dx, dy := float64(pa[0]-pb[0]), float64(pa[1]-pb[1])
	return int(math.Sqrt(dx*dx + dy*dy))
}

// reverse returns every node's arcs followed backward: arcs into it on a
// directed graph, the same arcs on an undirected one.
func (g *Graph) reverseAdjacency() [][]arc {
	if !g.Directed {
		return g.adjacency()
	}
	adj := make([][]arc, g.Len())
	for i, e := range g.Edges {
		adj[e.To] = append(adj[e.To], arc{e.From, e.Weight, i})
	}
	return adj
}

func reversed(s []int) []int {
	out := make([]int, len(s))
	for i, v := range s {
		out[len(s)-1-i] = v
	}
	return out
}
