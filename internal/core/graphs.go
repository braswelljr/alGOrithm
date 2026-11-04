package core

import (
	"errors"
	"fmt"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/graphs"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Example graphs for the graph entries.
const (
	graphWeighted = "A B 4\nA C 2\nB C 5\nB D 10\nC E 3\nE D 4\nD F 11\nE F 9"
	graphMST      = "a b 4\na h 8\nb c 8\nb h 11\nc d 7\nc f 4\nc i 2\nd e 9\nd f 14\ne f 10\nf g 2\ng h 1\ng i 6\nh i 7"
	graphNegative = "directed\ns t 6\ns y 7\nt x 5\nt y 8\nt z -4\nx t -2\ny x -3\ny z 9\nz x 7\nz s 2"
	graphDressing = "directed\nundershorts pants\nundershorts shoes\npants belt\npants shoes\nbelt jacket\nshirt belt\nshirt tie\ntie jacket\nsocks shoes\nwatch"
	graphSCC      = "directed\na b\nb c\nb e\nb f\nc d\nc g\nd c\nd h\ne a\ne f\nf g\ng f\ng h\nh h"
	graphFlow     = "directed\ns v1 16\ns v2 13\nv1 v3 12\nv2 v1 4\nv2 v4 14\nv3 v2 9\nv3 t 20\nv4 v3 7\nv4 t 4"
	graphCycle    = "directed\nA B\nB C\nC D\nD B\nC E"
	graphTwoSides = "A B\nA D\nB C\nC D\nC E\nE F"
)

// graphRun is what a graph entry does once its graph is parsed and its start
// and end nodes are found.
type graphRun func(g *graphs.Graph, start, end int, r trace.Recorder) (Output, error)

// graphEntry builds a graph entry: Input.Text is the graph, Input.Other names
// the start node (default the first) and Input.Key the end node (default the
// last). call writes the Go expression for snippets, using g, start and end.
func graphEntry(id, name, file, fn, summary, time, space, example string, ends bool, call string, run graphRun) Entry {
	e := entry(id, name, "algorithms/graphs/"+file, fn)
	e.Summary, e.Time, e.Space, e.Input, e.Traced = summary, time, space, GraphInput, true
	e.Limits = fmt.Sprintf("at most %d nodes and %d edges", graphs.MaxNodes, graphs.MaxEdges)
	e.Example = &Input{Text: example}
	if ends {
		e.Params = []string{"start", "end"}
		e.Summary += " Name the start node (other) and the end node (key); they default to the first and last nodes."
	}
	e.Run = func(in Input, r trace.Recorder) (Output, error) {
		g, start, end, err := graphEnds(in)
		if err != nil {
			return Output{}, err
		}
		out, err := run(g, start, end, r)
		if err != nil {
			return Output{}, wrap(err)
		}
		out.Graph = graphOutput(g)
		return out, nil
	}
	e.Call = func(in Input) (string, string) {
		setup := fmt.Sprintf("g, err := graphs.Parse(%s)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", "`"+in.Text+"`")
		if ends {
			if g, start, end, err := graphEnds(in); err == nil {
				setup += fmt.Sprintf("start, _ := g.Lookup(%q)\n\tend, _ := g.Lookup(%q)\n\t_ = end\n\t", g.Names[start], g.Names[end])
			}
		}
		return setup, call
	}
	return e
}

// graphEnds parses in.Text and finds the start (in.Other) and end (in.Key)
// nodes.
func graphEnds(in Input) (*graphs.Graph, int, int, error) {
	g, err := graphs.Parse(in.Text)
	if err != nil {
		return nil, 0, 0, wrap(err)
	}
	find := func(name string, def int) (int, error) {
		if strings.TrimSpace(name) == "" {
			return def, nil
		}
		i, ok := g.Lookup(strings.TrimSpace(name))
		if !ok {
			return 0, inputError(fmt.Sprintf("the graph has no node %q", name))
		}
		return i, nil
	}
	start, err := find(in.Other, 0)
	if err != nil {
		return nil, 0, 0, err
	}
	end, err := find(in.Key, g.Len()-1)
	if err != nil {
		return nil, 0, 0, err
	}
	return g, start, end, nil
}

func graphOutput(g *graphs.Graph) *GraphOutput {
	out := &GraphOutput{Directed: g.Directed, Nodes: append([]string(nil), g.Names...), Edges: make([]GraphEdge, len(g.Edges))}
	for i, e := range g.Edges {
		out.Edges[i] = GraphEdge{From: e.From, To: e.To, Weight: e.Weight}
	}
	return out
}

func nodeNames(g *graphs.Graph, ids []int) string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = g.Names[id]
	}
	return strings.Join(out, " ")
}

func edgeName(g *graphs.Graph, i int) string {
	e := g.Edges[i]
	link := "-"
	if g.Directed {
		link = "->"
	}
	return fmt.Sprintf("%s %s %s (%d)", g.Names[e.From], link, g.Names[e.To], e.Weight)
}

func distance(d int) string {
	if d == graphs.Infinity {
		return "unreachable"
	}
	return fmt.Sprint(d)
}

// distances lists every node's distance and path from start.
func distances(g *graphs.Graph, start int, dist, prev []int) []string {
	lines := make([]string, g.Len())
	for v := range lines {
		lines[v] = fmt.Sprintf("%s: %s", g.Names[v], distance(dist[v]))
		if p := graphs.PathTo(prev, start, v); p != nil && v != start {
			lines[v] += " via " + nodeNames(g, p)
		}
	}
	return lines
}

func graphEntries() []Entry {
	order := func(f func(*graphs.Graph, int, trace.Recorder) ([]int, error)) graphRun {
		return func(g *graphs.Graph, start, _ int, r trace.Recorder) (Output, error) {
			o, err := f(g, start, r)
			if err != nil {
				return Output{}, err
			}
			return Output{Values: o, Number: ptr(len(o)), Lines: []string{"order: " + nodeNames(g, o)}}, nil
		}
	}
	spanning := func(f func(*graphs.Graph, trace.Recorder) ([]int, int, error)) graphRun {
		return func(g *graphs.Graph, _, _ int, r trace.Recorder) (Output, error) {
			edges, total, err := f(g, r)
			if err != nil {
				return Output{}, err
			}
			var lines []string
			for _, i := range edges {
				lines = append(lines, edgeName(g, i))
			}
			lines = append(lines, fmt.Sprintf("total weight %d", total))
			return Output{Values: edges, Number: &total, Lines: lines}, nil
		}
	}
	return []Entry{
		graphEntry("graph-bfs", "Breadth-First Traversal", "traversal.go", "BreadthFirst",
			"Visit every node reachable from the start, nearest first, level by level.", "O(V + E)", "O(V)", graphWeighted, true,
			"graphs.BreadthFirst(g, start)", order(graphs.TraceBreadthFirst)),
		graphEntry("graph-dfs", "Depth-First Traversal", "traversal.go", "DepthFirst",
			"Visit every node reachable from the start, going as deep as possible before backing up.", "O(V + E)", "O(V)", graphWeighted, true,
			"graphs.DepthFirst(g, start)", order(graphs.TraceDepthFirst)),
		graphEntry("graph-dijkstra", "Dijkstra's Shortest Paths", "dijkstra.go", "Dijkstra",
			"Shortest distances from the start with non-negative weights: always settle the closest unsettled node (1956).", "O((V + E) log V)", "O(V)", graphWeighted, true,
			"graphs.Dijkstra(g, start)",
			func(g *graphs.Graph, start, end int, r trace.Recorder) (Output, error) {
				dist, prev, err := graphs.TraceDijkstra(g, start, r)
				if err != nil {
					return Output{}, err
				}
				path := graphs.PathTo(prev, start, end)
				r.Record(trace.Event{Kind: trace.Mark, Indices: path, Note: "path"})
				return Output{Path: path, Number: ptr(dist[end]), Lines: distances(g, start, dist, prev)}, nil
			}),
		graphEntry("bellman-ford", "Bellman-Ford", "bellmanford.go", "BellmanFord",
			"Shortest distances that allow negative weights, relaxing every edge V-1 times; finds negative cycles.", "O(V * E)", "O(V)", graphNegative, true,
			"graphs.BellmanFord(g, start)",
			func(g *graphs.Graph, start, end int, r trace.Recorder) (Output, error) {
				dist, prev, cycle, err := graphs.TraceBellmanFord(g, start, r)
				if err != nil && !errors.Is(err, graphs.ErrNegativeCycle) {
					return Output{}, err
				}
				if err != nil { // a negative cycle is an answer, shown to the reader
					return Output{Bool: ptr(false), Path: cycle, Lines: []string{"negative cycle: " + nodeNames(g, cycle), "distances on or past it are unbounded"}}, nil //nolint:nilerr // a negative cycle is the answer to show, not a failure
				}
				path := graphs.PathTo(prev, start, end)
				r.Record(trace.Event{Kind: trace.Mark, Indices: path, Note: "path"})
				return Output{Bool: ptr(true), Path: path, Number: ptr(dist[end]), Lines: distances(g, start, dist, prev)}, nil
			}),
		graphEntry("floyd-warshall", "Floyd-Warshall", "floydwarshall.go", "FloydWarshall",
			"The shortest distance between every pair of nodes: for each node, see whether going through it shortens any path. Unreachable pairs are blank.", "O(V^3)", "O(V^2)", graphWeighted, false,
			"graphs.FloydWarshall(g)",
			func(g *graphs.Graph, _, _ int, r trace.Recorder) (Output, error) {
				d, err := graphs.TraceFloydWarshall(g, r)
				if err != nil {
					return Output{}, err
				}
				n := g.Len()
				t := &TableOutput{Rows: n, Cols: n, Cells: make([]int, 0, n*n), RowLabels: g.Names, ColLabels: g.Names, Blank: ptr(unreachable)}
				var unreachablePairs []string
				for i := range d {
					for j, v := range d[i] {
						if v == graphs.Infinity {
							v = unreachable // graphs.Infinity is beyond what JavaScript holds exactly
							unreachablePairs = append(unreachablePairs, g.Names[i]+" to "+g.Names[j])
						}
						t.Cells = append(t.Cells, v)
					}
				}
				var lines []string
				if len(unreachablePairs) > 0 {
					lines = []string{"unreachable: " + strings.Join(unreachablePairs, ", ")}
				}
				return Output{Table: t, Lines: lines}, nil
			}),
		graphEntry("topological-sort", "Topological Sort", "topological.go", "TopologicalSort",
			"Order a directed acyclic graph so every edge points forward: repeatedly take a node nothing points to (Kahn, 1962).", "O(V + E)", "O(V)", graphDressing, false,
			"graphs.TopologicalSort(g)",
			func(g *graphs.Graph, _, _ int, r trace.Recorder) (Output, error) {
				o, err := graphs.TraceTopologicalSort(g, r)
				if err != nil && !errors.Is(err, graphs.ErrCycle) {
					return Output{}, err
				}
				lines := []string{"order: " + nodeNames(g, o)}
				if err != nil {
					lines = append(lines, fmt.Sprintf("stuck after %d of %d nodes: the rest are on a cycle", len(o), g.Len()))
				}
				return Output{Values: o, Bool: ptr(err == nil), Lines: lines}, nil
			}),
		graphEntry("kruskal", "Kruskal's Minimum Spanning Tree", "mst.go", "Kruskal",
			"Take the edges lightest first, skipping any that would close a cycle (1956).", "O(E log E)", "O(V + E)", graphMST, false,
			"graphs.Kruskal(g)", spanning(graphs.TraceKruskal)),
		graphEntry("prim", "Prim's Minimum Spanning Tree", "mst.go", "Prim",
			"Grow a tree from one node, always adding the lightest edge to a node outside it (Jarnik 1930, Prim 1957).", "O(E log V)", "O(V + E)", graphMST, false,
			"graphs.Prim(g)", spanning(graphs.TracePrim)),
		graphEntry("tarjan-scc", "Tarjan's Strongly Connected Components", "tarjan.go", "TarjanSCC",
			"The groups of nodes that can all reach each other, from one depth-first search (1972).", "O(V + E)", "O(V)", graphSCC, false,
			"graphs.TarjanSCC(g)",
			func(g *graphs.Graph, _, _ int, r trace.Recorder) (Output, error) {
				comps := graphs.TraceTarjanSCC(g, r)
				lines := make([]string, len(comps))
				for i, c := range comps {
					lines[i] = fmt.Sprintf("component %d: %s", i, nodeNames(g, c))
				}
				return Output{Number: ptr(len(comps)), Lines: lines}, nil
			}),
		graphEntry("bipartite-check", "Bipartite Check", "bipartite.go", "Bipartite",
			"Whether the nodes split into two sides with every edge between them, by two-coloring; otherwise an odd cycle proves it cannot.", "O(V + E)", "O(V)", graphTwoSides, false,
			"graphs.Bipartite(g)",
			func(g *graphs.Graph, _, _ int, r trace.Recorder) (Output, error) {
				ok, side, odd := graphs.TraceBipartite(g, r)
				if !ok {
					return Output{Bool: &ok, Path: odd, Lines: []string{"odd cycle: " + nodeNames(g, odd)}}, nil
				}
				var a, b []int
				for v, s := range side {
					if s == 0 {
						a = append(a, v)
					} else {
						b = append(b, v)
					}
				}
				return Output{Bool: &ok, Values: side, Lines: []string{"side 0: " + nodeNames(g, a), "side 1: " + nodeNames(g, b)}}, nil
			}),
		graphEntry("max-flow", "Maximum Flow (Edmonds-Karp)", "maxflow.go", "MaxFlow",
			"The most that can flow from the start to the end with edge weights as capacities, pushing along shortest augmenting paths (1972).", "O(V * E^2)", "O(V + E)", graphFlow, true,
			"graphs.MaxFlow(g, start, end)",
			func(g *graphs.Graph, start, end int, r trace.Recorder) (Output, error) {
				total, flow, err := graphs.TraceMaxFlow(g, start, end, r)
				if err != nil {
					return Output{}, err
				}
				var lines []string
				for i, f := range flow {
					if f != 0 {
						lines = append(lines, fmt.Sprintf("%s carries %d", edgeName(g, i), f))
					}
				}
				return Output{Number: &total, Values: flow, Lines: append(lines, fmt.Sprintf("maximum flow %d", total))}, nil
			}),
		graphEntry("cycle-detection", "Cycle Detection", "cycle.go", "CycleDetect",
			"Find a cycle by depth-first search: an edge back to a node still on the search path.", "O(V + E)", "O(V)", graphCycle, false,
			"graphs.CycleDetect(g)",
			func(g *graphs.Graph, _, _ int, r trace.Recorder) (Output, error) {
				c := graphs.TraceCycleDetect(g, r)
				if c == nil {
					return Output{Bool: ptr(false), Lines: []string{"no cycle"}}, nil
				}
				return Output{Bool: ptr(true), Path: c, Lines: []string{"cycle: " + nodeNames(g, c) + " " + g.Names[c[0]]}}, nil
			}),
	}
}
