package core

import (
	"errors"
	"fmt"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/graphs"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// graphsEntries is every entry of the graphs category.
func graphsEntries() []Entry {
	return []Entry{
		aStarGraphEntry(),
		articulationPointsEntry(),
		beamSearchEntry(),
		bellmanFordEntry(),
		bidirectionalBfsEntry(),
		bipartiteCheckEntry(),
		boruvkaEntry(),
		bridgesEntry(),
		criticalPathEntry(),
		cycleDetectionEntry(),
		dagShortestPathsEntry(),
		dinicEntry(),
		eulerPathEntry(),
		floydWarshallEntry(),
		graphBfsEntry(),
		graphDfsEntry(),
		graphDijkstraEntry(),
		hamiltonianPathEntry(),
		hopcroftKarpEntry(),
		iterativeDeepeningDfsEntry(),
		johnsonEntry(),
		kCoreEntry(),
		kosarajuSccEntry(),
		kruskalEntry(),
		maxFlowEntry(),
		pagerankEntry(),
		primEntry(),
		stoerWagnerEntry(),
		tarjanSccEntry(),
		topologicalSortEntry(),
		transitiveClosureEntry(),
		tspHeldKarpEntry(),
		tspNearestNeighborEntry(),
		uniformCostSearchEntry(),
	}
}

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

func edgeName(g *graphs.Graph, i int) string {
	e := g.Edges[i]
	link := "-"
	if g.Directed {
		link = "->"
	}
	return fmt.Sprintf("%s %s %s (%d)", g.Names[e.From], link, g.Names[e.To], e.Weight)
}

func graphsOrder(f func(*graphs.Graph, int, trace.Recorder) ([]int, error)) graphRun {
	return func(g *graphs.Graph, start, _ int, r trace.Recorder) (Output, error) {
		o, err := f(g, start, r)
		if err != nil {
			return Output{}, err
		}
		return Output{Values: o, Number: ptr(len(o)), Lines: []string{"order: " + nodeNames(g, o)}}, nil
	}
}

func graphsSpanning(f func(*graphs.Graph, trace.Recorder) ([]int, int, error)) graphRun {
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

func graphBfsEntry() Entry {
	return graphEntry("graph-bfs", "Breadth-First Traversal", "traversal.go", "BreadthFirst",
		"Visit every node reachable from the start, nearest first, level by level.", "O(V + E)", "O(V)", graphWeighted, true,
		"graphs.BreadthFirst(g, start)", graphsOrder(graphs.TraceBreadthFirst))
}

func graphDfsEntry() Entry {
	return graphEntry("graph-dfs", "Depth-First Traversal", "traversal.go", "DepthFirst",
		"Visit every node reachable from the start, going as deep as possible before backing up.", "O(V + E)", "O(V)", graphWeighted, true,
		"graphs.DepthFirst(g, start)", graphsOrder(graphs.TraceDepthFirst))
}

func graphDijkstraEntry() Entry {
	return graphEntry("graph-dijkstra", "Dijkstra's Shortest Paths", "dijkstra.go", "Dijkstra",
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
		})
}

func bellmanFordEntry() Entry {
	return graphEntry("bellman-ford", "Bellman-Ford", "bellmanford.go", "BellmanFord",
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
		})
}

func floydWarshallEntry() Entry {
	return graphEntry("floyd-warshall", "Floyd-Warshall", "floydwarshall.go", "FloydWarshall",
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
		})
}

func topologicalSortEntry() Entry {
	return graphEntry("topological-sort", "Topological Sort", "topological.go", "TopologicalSort",
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
		})
}

func kruskalEntry() Entry {
	return graphEntry("kruskal", "Kruskal's Minimum Spanning Tree", "mst.go", "Kruskal",
		"Take the edges lightest first, skipping any that would close a cycle (1956).", "O(E log E)", "O(V + E)", graphMST, false,
		"graphs.Kruskal(g)", graphsSpanning(graphs.TraceKruskal))
}

func primEntry() Entry {
	return graphEntry("prim", "Prim's Minimum Spanning Tree", "mst.go", "Prim",
		"Grow a tree from one node, always adding the lightest edge to a node outside it (Jarnik 1930, Prim 1957).", "O(E log V)", "O(V + E)", graphMST, false,
		"graphs.Prim(g)", graphsSpanning(graphs.TracePrim))
}

func tarjanSccEntry() Entry {
	return graphEntry("tarjan-scc", "Tarjan's Strongly Connected Components", "tarjan.go", "TarjanSCC",
		"The groups of nodes that can all reach each other, from one depth-first search (1972).", "O(V + E)", "O(V)", graphSCC, false,
		"graphs.TarjanSCC(g)",
		func(g *graphs.Graph, _, _ int, r trace.Recorder) (Output, error) {
			comps := graphs.TraceTarjanSCC(g, r)
			lines := make([]string, len(comps))
			for i, c := range comps {
				lines[i] = fmt.Sprintf("component %d: %s", i, nodeNames(g, c))
			}
			return Output{Number: ptr(len(comps)), Lines: lines}, nil
		})
}

func bipartiteCheckEntry() Entry {
	return graphEntry("bipartite-check", "Bipartite Check", "bipartite.go", "Bipartite",
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
		})
}

func maxFlowEntry() Entry {
	return graphEntry("max-flow", "Maximum Flow (Edmonds-Karp)", "maxflow.go", "MaxFlow",
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
		})
}

func cycleDetectionEntry() Entry {
	return graphEntry("cycle-detection", "Cycle Detection", "cycle.go", "CycleDetect",
		"Find a cycle by depth-first search: an edge back to a node still on the search path.", "O(V + E)", "O(V)", graphCycle, false,
		"graphs.CycleDetect(g)",
		func(g *graphs.Graph, _, _ int, r trace.Recorder) (Output, error) {
			c := graphs.TraceCycleDetect(g, r)
			if c == nil {
				return Output{Bool: ptr(false), Lines: []string{"no cycle"}}, nil
			}
			return Output{Bool: ptr(true), Path: c, Lines: []string{"cycle: " + nodeNames(g, c) + " " + g.Names[c[0]]}}, nil
		})
}

// More example graphs.
const (
	graphCities  = "A B 5\nA C 4\nB D 5\nC D 5\nC E 6\nD F 4\nE F 5\n# pos A 0 0\n# pos B 4 3\n# pos C 3 -2\n# pos D 7 0\n# pos E 7 -5\n# pos F 10 -2"
	graphBridges = "A B\nB C\nC A\nC D\nD E\nE F\nF D\nB G"
	graphEuler   = "A B\nB C\nC A\nC D\nD A"
	graphTour    = "A B 2\nA C 9\nA D 10\nB C 6\nB D 4\nC D 8\nA E 3\nB E 7\nC E 5\nD E 1"
	graphMatch   = "L1 R1\nL1 R2\nL2 R1\nL3 R2\nL3 R3\nL4 R3"
	graphWeb     = "directed\nA B\nA C\nB C\nC A\nD C"
	graphCore    = "A B\nA C\nA D\nB C\nB D\nC D\nD E\nE F"
	graphCut     = "1 2 2\n1 5 3\n2 3 3\n2 5 2\n2 6 2\n3 4 4\n3 7 2\n4 7 2\n4 8 2\n5 6 3\n6 7 1\n7 8 3"
	graphDAG     = "directed\nr s 5\nr t 3\ns t 2\ns x 6\nt x 7\nt y 4\nt z 2\nx y -1\nx z 1\ny z -2"
	graphProject = "directed\nstart design 3\nstart buy 1\ndesign build 5\nbuy build 2\nbuild test 4\nbuild docs 1\ntest ship 2\ndocs ship 1"
)

// optRun is a graph entry's run that also reads the input (for its options
// and node positions).
type optRun func(g *graphs.Graph, start, end int, in Input, r trace.Recorder) (Output, error)

// optEntry is graphEntry for entries with options: it builds the same entry
// and replaces Run so run sees the whole input. extra is more setup for the
// snippet (it may use in.Text); call is the Go expression.
func optEntry(id, name, file, fn, summary, time, space, example string, ends bool, opts []Option, extra func(in Input) string, call func(in Input) string, run optRun) Entry {
	e := graphEntry(id, name, file, fn, summary, time, space, example, ends, "", nil)
	e.Options = opts
	if len(opts) > 0 {
		e.Example.Options = map[string]int{}
		for _, o := range opts {
			e.Example.Options[o.Name] = o.Default
		}
	}
	e.Run = func(in Input, r trace.Recorder) (Output, error) {
		g, start, end, err := graphEnds(in)
		if err != nil {
			return Output{}, err
		}
		out, err := run(g, start, end, in, r)
		if err != nil {
			return Output{}, wrap(err)
		}
		out.Graph = graphOutput(g)
		return out, nil
	}
	base := e.Call
	e.Call = func(in Input) (string, string) {
		setup, _ := base(in)
		if extra != nil {
			setup += extra(in)
		}
		return setup, call(in)
	}
	return e
}

// positionsSetup reads node positions in a snippet.
func positionsSetup(in Input) string {
	return fmt.Sprintf("pos, err := graphs.ParsePositions(g, %s)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", "`"+in.Text+"`")
}

// pathOutput describes a path from start to end and its cost.
func pathOutput(g *graphs.Graph, path []int, cost int) Output {
	if path == nil {
		return Output{Bool: ptr(false), Lines: []string{"no path"}}
	}
	out := Output{Path: path, Bool: ptr(true), Lines: []string{"path: " + nodeNames(g, path)}}
	if cost >= 0 {
		out.Number = &cost
		out.Lines = append(out.Lines, fmt.Sprintf("cost %d, %d edges", cost, len(path)-1))
	} else {
		out.Number = ptr(len(path) - 1)
		out.Lines = append(out.Lines, fmt.Sprintf("%d edges", len(path)-1))
	}
	return out
}

func componentLines(g *graphs.Graph, comps [][]int) []string {
	lines := make([]string, len(comps))
	for i, c := range comps {
		lines[i] = fmt.Sprintf("component %d: %s", i, nodeNames(g, c))
	}
	return lines
}

// matrixOutput is an n x n table of a graph's nodes, Infinity shown blank.
func matrixOutput(g *graphs.Graph, m [][]int) *TableOutput {
	n := g.Len()
	t := &TableOutput{Rows: n, Cols: n, Cells: make([]int, 0, n*n), RowLabels: g.Names, ColLabels: g.Names, Blank: ptr(unreachable)}
	for _, row := range m {
		for _, v := range row {
			if v == graphs.Infinity {
				v = unreachable
			}
			t.Cells = append(t.Cells, v)
		}
	}
	return t
}

func edgeLines(g *graphs.Graph, edges []int) []string {
	lines := make([]string, len(edges))
	for i, e := range edges {
		lines[i] = edgeName(g, e)
	}
	return lines
}

func iterativeDeepeningDfsEntry() Entry {
	return optEntry("iterative-deepening-dfs", "Iterative Deepening DFS", "iddfs.go", "IterativeDeepening",
		"Depth-first searches limited to depth 0, 1, 2, ...: the memory of DFS with the fewest-edges path of BFS.", "O(b^d)", "O(d)", graphWeighted, true,
		[]Option{{"depth", "deepest depth (0 = number of nodes)", 0, 0, graphs.MaxNodes}}, nil,
		func(in Input) string {
			return fmt.Sprintf("graphs.IterativeDeepening(g, start, end, %d)", in.opt("depth", 0))
		},
		func(g *graphs.Graph, start, end int, in Input, r trace.Recorder) (Output, error) {
			path, err := graphs.TraceIterativeDeepening(g, start, end, in.opt("depth", 0), r)
			if err != nil {
				return Output{}, err
			}
			return pathOutput(g, path, -1), nil
		})
}

func bidirectionalBfsEntry() Entry {
	return optEntry("bidirectional-bfs", "Bidirectional BFS", "bidirectional.go", "BidirectionalBFS",
		"Breadth-first search from both ends at once until the frontiers meet: about 2b^(d/2) nodes instead of b^d.", "O(b^(d/2))", "O(b^(d/2))", graphWeighted, true,
		nil, nil, fixed("graphs.BidirectionalBFS(g, start, end)"),
		func(g *graphs.Graph, start, end int, _ Input, r trace.Recorder) (Output, error) {
			path, err := graphs.TraceBidirectionalBFS(g, start, end, r)
			if err != nil {
				return Output{}, err
			}
			return pathOutput(g, path, -1), nil
		})
}

func uniformCostSearchEntry() Entry {
	return optEntry("uniform-cost-search", "Uniform-Cost Search", "ucs.go", "UniformCost",
		"Expand the cheapest frontier node first and stop when the goal is settled: Dijkstra's algorithm as a search.", "O((V + E) log V)", "O(V)", graphWeighted, true,
		nil, nil, fixed("graphs.UniformCost(g, start, end)"),
		func(g *graphs.Graph, start, end int, _ Input, r trace.Recorder) (Output, error) {
			path, cost, err := graphs.TraceUniformCost(g, start, end, r)
			if err != nil {
				return Output{}, err
			}
			return pathOutput(g, path, cost), nil
		})
}

func aStarGraphEntry() Entry {
	return optEntry("a-star-graph", "A* Search (graph)", "astargraph.go", "AStarGraph",
		"Uniform-cost search guided by the straight-line distance to the goal, from \"# pos NAME X Y\" lines; still the cheapest path when edges are at least as long as the straight line.", "O((V + E) log V)", "O(V)", graphCities, true,
		nil, positionsSetup, fixed("graphs.AStarGraph(g, pos, start, end)"),
		func(g *graphs.Graph, start, end int, in Input, r trace.Recorder) (Output, error) {
			pos, err := graphs.ParsePositions(g, in.Text)
			if err != nil {
				return Output{}, err
			}
			path, cost, err := graphs.TraceAStarGraph(g, pos, start, end, r)
			if err != nil {
				return Output{}, err
			}
			out := pathOutput(g, path, cost)
			out.Lines = append(out.Lines, fmt.Sprintf("%d of %d nodes have positions", len(pos), g.Len()))
			return out, nil
		})
}

func beamSearchEntry() Entry {
	return optEntry("beam-search", "Beam Search", "beam.go", "BeamSearch",
		"Keep only the k most promising nodes of each level, by straight-line distance to the goal: little memory, no guarantee.", "O(d * k * b log(k * b))", "O(V)", graphCities, true,
		[]Option{{"width", "beam width", 2, 1, graphs.MaxNodes}}, positionsSetup,
		func(in Input) string {
			return fmt.Sprintf("graphs.BeamSearch(g, pos, start, end, %d)", in.opt("width", 2))
		},
		func(g *graphs.Graph, start, end int, in Input, r trace.Recorder) (Output, error) {
			pos, err := graphs.ParsePositions(g, in.Text)
			if err != nil {
				return Output{}, err
			}
			path, cost, err := graphs.TraceBeamSearch(g, pos, start, end, in.opt("width", 2), r)
			if err != nil {
				return Output{}, err
			}
			out := pathOutput(g, path, cost)
			if path == nil {
				out.Lines = append(out.Lines, "the beam dropped every way to the goal; try a wider beam")
			}
			return out, nil
		})
}

func dagShortestPathsEntry() Entry {
	return optEntry("dag-shortest-paths", "Shortest Paths in a DAG", "dagpaths.go", "DAGShortestPaths",
		"Relax every edge in topological order: linear time, and negative weights are fine.", "O(V + E)", "O(V)", graphDAG, true,
		nil, nil, fixed("graphs.DAGShortestPaths(g, start)"),
		func(g *graphs.Graph, start, end int, _ Input, r trace.Recorder) (Output, error) {
			dist, prev, err := graphs.TraceDAGShortestPaths(g, start, r)
			if err != nil {
				return Output{}, err
			}
			path := graphs.PathTo(prev, start, end)
			r.Record(trace.Event{Kind: trace.Mark, Indices: path, Note: "path"})
			out := Output{Path: path, Lines: distances(g, start, dist, prev)}
			if dist[end] != graphs.Infinity {
				out.Number = ptr(dist[end])
			}
			return out, nil
		})
}

func criticalPathEntry() Entry {
	return optEntry("critical-path", "Critical Path (longest path in a DAG)", "dagpaths.go", "DAGLongestPath",
		"The heaviest path in a DAG of tasks with durations: no amount of parallel work finishes the project sooner.", "O(V + E)", "O(V)", graphProject, false,
		nil, nil, fixed("graphs.DAGLongestPath(g)"),
		func(g *graphs.Graph, _, _ int, _ Input, r trace.Recorder) (Output, error) {
			path, length, err := graphs.TraceDAGLongestPath(g, r)
			if err != nil {
				return Output{}, err
			}
			return Output{Path: path, Number: &length, Lines: []string{"critical path: " + nodeNames(g, path), fmt.Sprintf("length %d", length)}}, nil
		})
}

func tspHeldKarpEntry() Entry {
	return optEntry("tsp-held-karp", "Traveling Salesperson (Held-Karp)", "tsp.go", "HeldKarp",
		"The cheapest tour through every node, exactly, by dynamic programming over subsets (1962).", "O(2^n * n^2)", "O(2^n * n)", graphTour, false,
		nil, nil, fixed("graphs.HeldKarp(g)"),
		func(g *graphs.Graph, _, _ int, _ Input, r trace.Recorder) (Output, error) {
			tour, cost, err := graphs.TraceHeldKarp(g, r)
			if err != nil {
				return Output{}, err
			}
			return Output{Path: tour, Number: &cost, Lines: []string{"tour: " + nodeNames(g, tour), fmt.Sprintf("cost %d (optimal)", cost)}}, nil
		})
}

func tspNearestNeighborEntry() Entry {
	return optEntry("tsp-nearest-neighbor", "Traveling Salesperson (nearest neighbor)", "tsp.go", "NearestNeighborTour",
		"Always go to the nearest unvisited node, then return: fast, but not always the cheapest tour.", "O(n^2)", "O(n^2)", graphTour, true,
		nil, nil, fixed("graphs.NearestNeighborTour(g, start)"),
		func(g *graphs.Graph, start, _ int, _ Input, r trace.Recorder) (Output, error) {
			tour, cost, err := graphs.TraceNearestNeighborTour(g, start, r)
			if err != nil {
				return Output{}, err
			}
			lines := []string{"tour: " + nodeNames(g, tour), fmt.Sprintf("cost %d", cost)}
			if g.Len() <= graphs.MaxHeldKarp {
				if _, best, err := graphs.HeldKarp(g); err == nil {
					lines = append(lines, fmt.Sprintf("the optimal tour costs %d", best))
				}
			}
			return Output{Path: tour, Number: &cost, Lines: lines}, nil
		})
}

func hamiltonianPathEntry() Entry {
	return optEntry("hamiltonian-path", "Hamiltonian Path", "hamiltonian.go", "HamiltonianPath",
		"A path through every node exactly once, by backtracking; NP-complete, so small graphs only.", "O(n!)", "O(n)", graphTour, false,
		[]Option{{"cycle", "return to the start (1 yes, 0 no)", 0, 0, 1}}, nil,
		func(in Input) string { return fmt.Sprintf("graphs.HamiltonianPath(g, %v)", in.opt("cycle", 0) == 1) },
		func(g *graphs.Graph, _, _ int, in Input, r trace.Recorder) (Output, error) {
			path, err := graphs.TraceHamiltonianPath(g, in.opt("cycle", 0) == 1, r)
			if err != nil {
				return Output{}, err
			}
			if path == nil {
				return Output{Bool: ptr(false), Lines: []string{"none exists"}}, nil
			}
			return Output{Path: path, Bool: ptr(true), Lines: []string{nodeNames(g, path)}}, nil
		})
}

func eulerPathEntry() Entry {
	return optEntry("euler-path", "Euler Path (Hierholzer)", "euler.go", "EulerPath",
		"A walk using every edge exactly once, when the degrees allow it, splicing in detours (1873).", "O(V + E)", "O(V + E)", graphEuler, false,
		nil, nil, fixed("graphs.EulerPath(g)"),
		func(g *graphs.Graph, _, _ int, _ Input, r trace.Recorder) (Output, error) {
			nodes, edges, err := graphs.TraceEulerPath(g, r)
			if errors.Is(err, graphs.ErrNoEuler) { // a normal answer: the reason is the result
				return Output{Bool: ptr(false), Lines: []string{err.Error()}}, nil
			}
			if err != nil {
				return Output{}, err
			}
			return Output{Path: nodes, Values: edges, Bool: ptr(true), Lines: []string{"walk: " + nodeNames(g, nodes)}}, nil
		})
}

func graphsFlow(id, name, file, fn, summary, time, call string, f func(*graphs.Graph, int, int, trace.Recorder) (int, []int, error)) Entry {
	return optEntry(id, name, file, fn, summary, time, "O(V + E)", graphFlow, true, nil, nil, fixed(call),
		func(g *graphs.Graph, start, end int, _ Input, r trace.Recorder) (Output, error) {
			total, flows, err := f(g, start, end, r)
			if err != nil {
				return Output{}, err
			}
			lines := []string{fmt.Sprintf("maximum flow %d", total)}
			for i, fl := range flows {
				if fl != 0 {
					lines = append(lines, fmt.Sprintf("%s carries %d", edgeName(g, i), fl))
				}
			}
			return Output{Number: &total, Values: flows, Lines: lines}, nil
		})
}

func kosarajuSccEntry() Entry {
	return optEntry("kosaraju-scc", "Kosaraju's Strongly Connected Components", "kosaraju.go", "Kosaraju",
		"Finish order from one depth-first search, then a second over the reversed edges collects each component (1978).", "O(V + E)", "O(V + E)", graphSCC, false,
		nil, nil, fixed("graphs.Kosaraju(g)"),
		func(g *graphs.Graph, _, _ int, _ Input, r trace.Recorder) (Output, error) {
			comps := graphs.TraceKosaraju(g, r)
			return Output{Number: ptr(len(comps)), Lines: componentLines(g, comps)}, nil
		})
}

func articulationPointsEntry() Entry {
	return optEntry("articulation-points", "Articulation Points", "bridges.go", "ArticulationPoints",
		"The nodes whose removal disconnects the graph, from one low-link depth-first search (Tarjan, 1973).", "O(V + E)", "O(V)", graphBridges, false,
		nil, nil, fixed("graphs.ArticulationPoints(g)"),
		func(g *graphs.Graph, _, _ int, _ Input, r trace.Recorder) (Output, error) {
			points, err := graphs.TraceArticulationPoints(g, r)
			if err != nil {
				return Output{}, err
			}
			return Output{Values: points, Number: ptr(len(points)), Lines: []string{"articulation points: " + nodeNames(g, points)}}, nil
		})
}

func bridgesEntry() Entry {
	return optEntry("bridges", "Bridges", "bridges.go", "Bridges",
		"The edges whose removal disconnects the graph, from the same low-link search.", "O(V + E)", "O(V)", graphBridges, false,
		nil, nil, fixed("graphs.Bridges(g)"),
		func(g *graphs.Graph, _, _ int, _ Input, r trace.Recorder) (Output, error) {
			bridges, err := graphs.TraceBridges(g, r)
			if err != nil {
				return Output{}, err
			}
			return Output{Values: bridges, Number: ptr(len(bridges)), Lines: edgeLines(g, bridges)}, nil
		})
}

func boruvkaEntry() Entry {
	return optEntry("boruvka", "Boruvka's Minimum Spanning Tree", "boruvka.go", "Boruvka",
		"Every component takes its cheapest outgoing edge at once, round after round (1926); the components at least halve each round.", "O(E log V)", "O(V)", graphMST, false,
		nil, nil, fixed("graphs.Boruvka(g)"),
		func(g *graphs.Graph, _, _ int, _ Input, r trace.Recorder) (Output, error) {
			edges, total, err := graphs.TraceBoruvka(g, r)
			if err != nil {
				return Output{}, err
			}
			return Output{Values: edges, Number: &total, Lines: append(edgeLines(g, edges), fmt.Sprintf("total weight %d", total))}, nil
		})
}

func dinicEntry() Entry {
	return graphsFlow("dinic", "Dinic's Maximum Flow", "dinic.go", "Dinic",
		"Level the residual graph breadth-first, push a blocking flow depth-first, repeat (1970); the same answer as Edmonds-Karp in fewer phases.", "O(V^2 * E)",
		"graphs.Dinic(g, start, end)", graphs.TraceDinic)
}

func hopcroftKarpEntry() Entry {
	return optEntry("hopcroft-karp", "Hopcroft-Karp Bipartite Matching", "matching.go", "HopcroftKarp",
		"The most pairs of a bipartite graph with no node in two pairs, augmenting along many shortest paths per phase (1973).", "O(E * sqrt V)", "O(V)", graphMatch, false,
		nil, nil, fixed("graphs.HopcroftKarp(g)"),
		func(g *graphs.Graph, _, _ int, _ Input, r trace.Recorder) (Output, error) {
			m, err := graphs.TraceHopcroftKarp(g, r)
			if err != nil {
				return Output{}, err
			}
			return Output{Values: m, Number: ptr(len(m)), Lines: edgeLines(g, m)}, nil
		})
}

func johnsonEntry() Entry {
	return optEntry("johnson", "Johnson's All-Pairs Shortest Paths", "johnson.go", "Johnson",
		"Reweight with Bellman-Ford so every weight is non-negative, then Dijkstra from every node (1977); negative edges allowed.", "O(V * E log V)", "O(V^2)", graphNegative, false,
		nil, nil, fixed("graphs.Johnson(g)"),
		func(g *graphs.Graph, _, _ int, _ Input, r trace.Recorder) (Output, error) {
			d, err := graphs.TraceJohnson(g, r)
			if err != nil {
				return Output{}, err
			}
			return Output{Table: matrixOutput(g, d)}, nil
		})
}

func transitiveClosureEntry() Entry {
	return optEntry("transitive-closure", "Transitive Closure (Warshall)", "closure.go", "TransitiveClosure",
		"Which nodes reach which: j is reachable from i through k when i reaches k and k reaches j (1962).", "O(V^3)", "O(V^2)", graphCycle, false,
		nil, nil, fixed("graphs.TransitiveClosure(g)"),
		func(g *graphs.Graph, _, _ int, _ Input, r trace.Recorder) (Output, error) {
			return Output{Table: matrixOutput(g, graphs.TraceTransitiveClosure(g, r))}, nil
		})
}

func pagerankEntry() Entry {
	return optEntry("pagerank", "PageRank", "pagerank.go", "PageRank",
		"How often a random surfer is at each node, following links with probability d and jumping anywhere otherwise (1998).", "O(k * (V + E))", "O(V)", graphWeb, false,
		[]Option{{"damping", "damping, in thousandths", 850, 0, 1000}, {"iterations", "most iterations", 100, 1, 1000}}, nil,
		func(in Input) string {
			return fmt.Sprintf("graphs.PageRank(g, %g, %d, 1e-9)", float64(in.opt("damping", 850))/1000, in.opt("iterations", 100))
		},
		func(g *graphs.Graph, _, _ int, in Input, r trace.Recorder) (Output, error) {
			rank, ran := graphs.TracePageRank(g, float64(in.opt("damping", 850))/1000, in.opt("iterations", 100), 1e-9, r)
			out := Output{Summary: map[string]float64{"iterations": float64(ran)}}
			for v, x := range rank {
				out.Lines = append(out.Lines, fmt.Sprintf("%s: %.4f", g.Names[v], x))
				out.Summary["rank of "+g.Names[v]] = x
			}
			return out, nil
		})
}

func kCoreEntry() Entry {
	return optEntry("k-core", "k-Core Decomposition", "kcore.go", "KCore",
		"Peel off the node of least degree, again and again: each node's core number is how deeply embedded it is (1983).", "O(V + E)", "O(V + E)", graphCore, false,
		nil, nil, fixed("graphs.KCore(g)"),
		func(g *graphs.Graph, _, _ int, _ Input, r trace.Recorder) (Output, error) {
			core := graphs.TraceKCore(g, r)
			lines := make([]string, len(core))
			k := 0
			for v, c := range core {
				lines[v] = fmt.Sprintf("%s: core %d", g.Names[v], c)
				k = max(k, c)
			}
			return Output{Values: core, Number: &k, Lines: lines}, nil
		})
}

func stoerWagnerEntry() Entry {
	return optEntry("stoer-wagner", "Stoer-Wagner Minimum Cut", "mincut.go", "StoerWagner",
		"The lightest set of edges that splits the graph in two, by merging the last two nodes of each maximum-adjacency phase (1997).", "O(V^3)", "O(V^2)", graphCut, false,
		nil, nil, fixed("graphs.StoerWagner(g)"),
		func(g *graphs.Graph, _, _ int, _ Input, r trace.Recorder) (Output, error) {
			cut, side, err := graphs.TraceStoerWagner(g, r)
			if err != nil {
				return Output{}, err
			}
			return Output{Number: &cut, Values: side, Lines: []string{fmt.Sprintf("minimum cut %d", cut), "one side: " + nodeNames(g, side)}}, nil
		})
}
