package graphs_test

import (
	"errors"
	"math"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/graphs"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func mustParse(t *testing.T, text string) *graphs.Graph {
	t.Helper()
	g, err := graphs.Parse(text)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func node(t *testing.T, g *graphs.Graph, name string) int {
	t.Helper()
	i, ok := g.Lookup(name)
	if !ok {
		t.Fatalf("no node %s", name)
	}
	return i
}

func names(g *graphs.Graph, ids []int) string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = g.Names[id]
	}
	return strings.Join(out, " ")
}

func TestParse(t *testing.T) {
	g := mustParse(t, "# a comment\ndirected\nA B 4\nB C\nD\n")
	if !g.Directed || g.Len() != 4 || len(g.Edges) != 2 || g.Edges[1].Weight != 1 {
		t.Fatalf("Parse = %+v", g)
	}
	back := mustParse(t, g.String())
	if !reflect.DeepEqual(back.Names, g.Names) || !reflect.DeepEqual(back.Edges, g.Edges) || !back.Directed {
		t.Errorf("String does not round-trip: %q", g.String())
	}
	for _, bad := range []string{"", "A B x", "A B 1 2", "# only a comment"} {
		if _, err := graphs.Parse(bad); !errors.Is(err, graphs.ErrInput) {
			t.Errorf("Parse(%q) error %v", bad, err)
		}
	}
	if _, err := graphs.Parse("A B x"); err == nil || !strings.Contains(err.Error(), "line 1") {
		t.Errorf("error does not name the line: %v", err)
	}
}

func TestGenerate(t *testing.T) {
	for _, k := range []string{graphs.Connected, graphs.Directed, graphs.DAG} {
		g, err := graphs.Generate(k, 12, 3)
		if err != nil {
			t.Fatal(err)
		}
		h, _ := graphs.Generate(k, 12, 3)
		if g.String() != h.String() {
			t.Errorf("%s not reproducible", k)
		}
		order, _ := graphs.BreadthFirst(g, 0)
		if len(order) != 12 {
			t.Errorf("%s: only %d nodes reachable from the first", k, len(order))
		}
	}
	dag, _ := graphs.Generate(graphs.DAG, 30, 9)
	if _, err := graphs.TopologicalSort(dag); err != nil {
		t.Errorf("a generated DAG has a cycle: %v", err)
	}
	if _, err := graphs.Generate("tree", 5, 1); err == nil {
		t.Error("unknown kind accepted")
	}
}

func TestTraversals(t *testing.T) {
	g := mustParse(t, "A B\nA C\nB D\nC D\nD E\nF")
	bfs, _ := graphs.BreadthFirst(g, 0)
	dfs, _ := graphs.DepthFirst(g, 0)
	if names(g, bfs) != "A B C D E" || names(g, dfs) != "A B D C E" {
		t.Errorf("bfs %s, dfs %s", names(g, bfs), names(g, dfs))
	}
	if _, err := graphs.BreadthFirst(g, 99); !errors.Is(err, graphs.ErrNode) {
		t.Error("an unknown start was accepted")
	}
	var log trace.Log
	if _, err := graphs.TraceBreadthFirst(g, 0, &log); err != nil {
		t.Fatal(err)
	}
	if log.Count(trace.Visit) != 5 || log.Events[len(log.Events)-1].Note != "order" {
		t.Error("traversal events wrong")
	}
}

// CLRS figure 24.6 (Dijkstra) and 24.4 (Bellman-Ford, negative weights).
const clrsDijkstra = "directed\ns t 10\ns y 5\nt x 1\nt y 2\ny t 3\ny x 9\ny z 2\nx z 4\nz x 6\nz s 7"
const clrsBellman = "directed\ns t 6\ns y 7\nt x 5\nt y 8\nt z -4\nx t -2\ny x -3\ny z 9\nz x 7\nz s 2"

func TestShortestPaths(t *testing.T) {
	g := mustParse(t, clrsDijkstra)
	dist, prev, err := graphs.Dijkstra(g, 0)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]int{"s": 0, "t": 8, "x": 9, "y": 5, "z": 7}
	for n, d := range want {
		if dist[node(t, g, n)] != d {
			t.Errorf("Dijkstra dist(%s) = %d, want %d", n, dist[node(t, g, n)], d)
		}
	}
	if p := names(g, graphs.PathTo(prev, 0, node(t, g, "x"))); p != "s y t x" {
		t.Errorf("path to x = %s", p)
	}
	neg := mustParse(t, clrsBellman)
	if _, _, err := graphs.Dijkstra(neg, 0); !errors.Is(err, graphs.ErrNegative) {
		t.Error("Dijkstra accepted a negative weight")
	}
	bd, _, _, err := graphs.BellmanFord(neg, 0)
	if err != nil {
		t.Fatal(err)
	}
	for n, d := range map[string]int{"s": 0, "t": 2, "x": 4, "y": 7, "z": -2} {
		if bd[node(t, neg, n)] != d {
			t.Errorf("Bellman-Ford dist(%s) = %d, want %d", n, bd[node(t, neg, n)], d)
		}
	}
	cyc := mustParse(t, "directed\nA B 1\nB C -3\nC B 1\nC D 1")
	_, _, c, err := graphs.BellmanFord(cyc, 0)
	if !errors.Is(err, graphs.ErrNegativeCycle) || len(c) != 2 {
		t.Errorf("negative cycle: %v %v", c, err)
	}
	if _, err := graphs.FloydWarshall(cyc); !errors.Is(err, graphs.ErrNegativeCycle) {
		t.Error("Floyd-Warshall missed the negative cycle")
	}
	if p := graphs.PathTo([]int{-1, 0, -1}, 0, 2); p != nil {
		t.Errorf("path to an unreached node = %v", p)
	}
}

// TestShortestPathsAgree cross-checks the three algorithms on random graphs.
func TestShortestPathsAgree(t *testing.T) {
	for seed := int64(1); seed <= 30; seed++ {
		kind := graphs.Connected
		if seed%2 == 0 {
			kind = graphs.Directed
		}
		g, _ := graphs.Generate(kind, 15, seed)
		all, err := graphs.FloydWarshall(g)
		if err != nil {
			t.Fatal(err)
		}
		for s := 0; s < g.Len(); s += 4 {
			dj, _, _ := graphs.Dijkstra(g, s)
			bf, _, _, _ := graphs.BellmanFord(g, s)
			if !reflect.DeepEqual(dj, bf) || !reflect.DeepEqual(dj, all[s]) {
				t.Fatalf("seed %d source %d: Dijkstra %v, Bellman-Ford %v, Floyd-Warshall %v", seed, s, dj, bf, all[s])
			}
		}
	}
}

func TestTopologicalSort(t *testing.T) {
	// CLRS figure 22.7: getting dressed.
	g := mustParse(t, "directed\nundershorts pants\nundershorts shoes\npants belt\npants shoes\nbelt jacket\nshirt belt\nshirt tie\ntie jacket\nsocks shoes\nwatch")
	order, err := graphs.TopologicalSort(g)
	if err != nil || len(order) != g.Len() {
		t.Fatalf("order %v, %v", order, err)
	}
	pos := map[int]int{}
	for i, v := range order {
		pos[v] = i
	}
	for _, e := range g.Edges {
		if pos[e.From] > pos[e.To] {
			t.Errorf("%s comes after %s", g.Names[e.From], g.Names[e.To])
		}
	}
	if _, err := graphs.TopologicalSort(mustParse(t, "directed\nA B\nB A")); !errors.Is(err, graphs.ErrCycle) {
		t.Error("a cycle was sorted")
	}
	if _, err := graphs.TopologicalSort(mustParse(t, "A B")); !errors.Is(err, graphs.ErrUndirected) {
		t.Error("an undirected graph was sorted")
	}
}

// CLRS figure 23.1: the minimum spanning tree weighs 37.
const clrsMST = "a b 4\na h 8\nb c 8\nb h 11\nc d 7\nc f 4\nc i 2\nd e 9\nd f 14\ne f 10\nf g 2\ng h 1\ng i 6\nh i 7"

func TestSpanningTrees(t *testing.T) {
	g := mustParse(t, clrsMST)
	ke, kw, _ := graphs.Kruskal(g)
	pe, pw, _ := graphs.Prim(g)
	if kw != 37 || pw != 37 || len(ke) != 8 || len(pe) != 8 {
		t.Errorf("Kruskal %d (%d edges), Prim %d (%d edges), want 37 with 8", kw, len(ke), pw, len(pe))
	}
	for seed := int64(1); seed <= 40; seed++ {
		r, _ := graphs.Generate(graphs.Connected, 20, seed)
		_, a, _ := graphs.Kruskal(r)
		_, b, _ := graphs.Prim(r)
		if a != b {
			t.Fatalf("seed %d: Kruskal %d, Prim %d", seed, a, b)
		}
	}
	forest := mustParse(t, "A B 1\nC D 2")
	if e, w, _ := graphs.Prim(forest); len(e) != 2 || w != 3 {
		t.Errorf("forest: %v %d", e, w)
	}
	if _, _, err := graphs.Kruskal(mustParse(t, "directed\nA B")); !errors.Is(err, graphs.ErrDirected) {
		t.Error("a directed graph was spanned")
	}
}

func TestTarjanSCC(t *testing.T) {
	// CLRS figure 22.9.
	g := mustParse(t, "directed\na b\nb c\nb e\nb f\nc d\nc g\nd c\nd h\ne a\ne f\nf g\ng f\ng h\nh h")
	var got []string
	for _, c := range graphs.TarjanSCC(g) {
		n := strings.Fields(names(g, c))
		sort.Strings(n)
		got = append(got, strings.Join(n, ""))
	}
	sort.Strings(got)
	if want := []string{"abe", "cd", "fg", "h"}; !reflect.DeepEqual(got, want) {
		t.Errorf("components %v, want %v", got, want)
	}
	if n := len(graphs.TarjanSCC(mustParse(t, "A B\nB C\nD E"))); n != 2 {
		t.Errorf("undirected components = %d, want 2", n)
	}
}

func TestBipartite(t *testing.T) {
	ok, side, _ := graphs.Bipartite(mustParse(t, "A B\nB C\nC D\nD A"))
	if !ok || side[0] == side[1] || side[0] != side[2] {
		t.Errorf("a square is bipartite: %v %v", ok, side)
	}
	g := mustParse(t, "A B\nB C\nC D\nD E\nE A")
	ok, _, cyc := graphs.Bipartite(g)
	if ok || len(cyc)%2 == 0 {
		t.Fatalf("a pentagon: ok %v, cycle %v", ok, cyc)
	}
	edge := map[[2]int]bool{}
	for _, e := range g.Edges {
		edge[[2]int{e.From, e.To}], edge[[2]int{e.To, e.From}] = true, true
	}
	for i := range cyc {
		if !edge[[2]int{cyc[i], cyc[(i+1)%len(cyc)]}] {
			t.Fatalf("odd cycle %v is not a cycle of the graph", cyc)
		}
	}
}

// minCut brute-forces the minimum s-t cut: the smallest total capacity of
// edges leaving a set that holds s but not t.
func minCut(g *graphs.Graph, s, sink int) int {
	best := graphs.Infinity
	n := g.Len()
	for mask := 0; mask < 1<<n; mask++ {
		if mask&(1<<s) == 0 || mask&(1<<sink) != 0 {
			continue
		}
		cut := 0
		for _, e := range g.Edges {
			in := func(v int) bool { return mask&(1<<v) != 0 }
			if in(e.From) && !in(e.To) {
				cut += e.Weight
			}
			if !g.Directed && in(e.To) && !in(e.From) {
				cut += e.Weight
			}
		}
		best = min(best, cut)
	}
	return best
}

func TestMaxFlow(t *testing.T) {
	// CLRS figure 26.6: the maximum flow is 23.
	g := mustParse(t, "directed\ns v1 16\ns v2 13\nv1 v3 12\nv2 v1 4\nv2 v4 14\nv3 v2 9\nv3 t 20\nv4 v3 7\nv4 t 4")
	total, flow, err := graphs.MaxFlow(g, node(t, g, "s"), node(t, g, "t"))
	if err != nil || total != 23 {
		t.Fatalf("max flow %d, %v", total, err)
	}
	net := make([]int, g.Len())
	for i, e := range g.Edges {
		if flow[i] < 0 || flow[i] > e.Weight {
			t.Errorf("edge %d carries %d of %d", i, flow[i], e.Weight)
		}
		net[e.From] -= flow[i]
		net[e.To] += flow[i]
	}
	for v, n := range net {
		if name := g.Names[v]; name != "s" && name != "t" && n != 0 {
			t.Errorf("flow is not conserved at %s: %d", name, n)
		}
	}
	for seed := int64(1); seed <= 25; seed++ {
		kind := graphs.Directed
		if seed%3 == 0 {
			kind = graphs.Connected
		}
		r, _ := graphs.Generate(kind, 7, seed)
		got, _, _ := graphs.MaxFlow(r, 0, r.Len()-1)
		if want := minCut(r, 0, r.Len()-1); got != want {
			t.Fatalf("seed %d (%s): max flow %d, min cut %d", seed, kind, got, want)
		}
	}
	if _, _, err := graphs.MaxFlow(g, 0, 0); err == nil {
		t.Error("source = sink accepted")
	}
}

func TestCycleDetect(t *testing.T) {
	cases := []struct {
		text  string
		cycle bool
	}{
		{"directed\nA B\nB C\nC A", true},
		{"directed\nA B\nB C\nA C", false},
		{"A B\nB C\nC A", true},
		{"A B\nB C\nC D", false},
		{"A A", true},
		{"A B\nA B", true},
	}
	for _, c := range cases {
		g := mustParse(t, c.text)
		cyc := graphs.CycleDetect(g)
		if (cyc != nil) != c.cycle {
			t.Errorf("%q: cycle %v", c.text, cyc)
			continue
		}
		if !c.cycle {
			continue
		}
		edge := map[[2]int]int{}
		for _, e := range g.Edges {
			edge[[2]int{e.From, e.To}]++
			if !g.Directed {
				edge[[2]int{e.To, e.From}]++
			}
		}
		for i := range cyc {
			if edge[[2]int{cyc[i], cyc[(i+1)%len(cyc)]}] == 0 {
				t.Errorf("%q: %v is not a cycle of the graph", c.text, names(g, cyc))
			}
		}
	}
}

func TestTraceEvents(t *testing.T) {
	g := mustParse(t, clrsMST)
	var log trace.Log
	if _, _, err := graphs.TraceKruskal(g, &log); err != nil {
		t.Fatal(err)
	}
	marks := 0
	for _, e := range log.Events {
		if e.Kind == trace.Mark && e.Note == "tree" {
			marks++
		}
	}
	if marks != 8 {
		t.Errorf("%d tree marks, want 8", marks)
	}
}

// bfsDistance is the fewest edges from s to t, or -1.
func bfsDistance(t *testing.T, g *graphs.Graph, s, target int) int {
	t.Helper()
	order, err := graphs.BreadthFirst(g, s)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, v := range order {
		found = found || v == target
	}
	if !found {
		return -1
	}
	unit := graphs.New(g.Directed)
	for _, n := range g.Names {
		unit.Node(n)
	}
	for _, e := range g.Edges {
		unit.Edges = append(unit.Edges, graphs.Edge{From: e.From, To: e.To, Weight: 1})
	}
	d, _, err := graphs.Dijkstra(unit, s)
	if err != nil {
		t.Fatal(err)
	}
	return d[target]
}

func validPath(t *testing.T, g *graphs.Graph, path []int, s, target int) {
	t.Helper()
	if len(path) == 0 || path[0] != s || path[len(path)-1] != target {
		t.Fatalf("path %v does not run %d to %d", path, s, target)
	}
	for i := 1; i < len(path); i++ {
		ok := false
		for _, e := range g.Edges {
			if (e.From == path[i-1] && e.To == path[i]) || (!g.Directed && e.To == path[i-1] && e.From == path[i]) {
				ok = true
			}
		}
		if !ok {
			t.Fatalf("path %v uses a missing edge %d-%d", path, path[i-1], path[i])
		}
	}
}

func TestSearchesAgree(t *testing.T) {
	for seed := int64(1); seed <= 40; seed++ {
		kind := graphs.Connected
		if seed%2 == 0 {
			kind = graphs.Directed
		}
		g, _ := graphs.Generate(kind, 12, seed)
		s, target := 0, g.Len()-1
		want := bfsDistance(t, g, s, target)
		for name, find := range map[string]func() ([]int, error){
			"iddfs": func() ([]int, error) { return graphs.IterativeDeepening(g, s, target, 0) },
			"bidir": func() ([]int, error) { return graphs.BidirectionalBFS(g, s, target) },
		} {
			path, err := find()
			if err != nil {
				t.Fatal(err)
			}
			if want < 0 {
				if path != nil {
					t.Fatalf("%s found %v to an unreachable node", name, path)
				}
				continue
			}
			validPath(t, g, path, s, target)
			if len(path)-1 != want {
				t.Fatalf("seed %d: %s path has %d edges, fewest is %d", seed, name, len(path)-1, want)
			}
		}
		dist, _, _ := graphs.Dijkstra(g, s)
		for name, find := range map[string]func() ([]int, int, error){
			"ucs":   func() ([]int, int, error) { return graphs.UniformCost(g, s, target) },
			"astar": func() ([]int, int, error) { return graphs.AStarGraph(g, nil, s, target) },
		} {
			path, cost, err := find()
			if err != nil {
				t.Fatal(err)
			}
			if dist[target] == graphs.Infinity {
				if path != nil || cost != -1 {
					t.Fatalf("%s reached an unreachable node", name)
				}
				continue
			}
			validPath(t, g, path, s, target)
			if cost != dist[target] {
				t.Fatalf("seed %d: %s cost %d, Dijkstra %d", seed, name, cost, dist[target])
			}
		}
	}
}

const cities = "A B 5\nA C 4\nB D 5\nC D 5\nC E 6\nD F 4\nE F 5\n# pos A 0 0\n# pos B 4 3\n# pos C 3 -2\n# pos D 7 0\n# pos E 7 -5\n# pos F 10 -2"

func TestAStarAndBeam(t *testing.T) {
	g := mustParse(t, cities)
	pos, err := graphs.ParsePositions(g, cities)
	if err != nil || len(pos) != 6 {
		t.Fatalf("positions %v, %v", pos, err)
	}
	var withH, without trace.Log
	path, cost, _ := graphs.TraceAStarGraph(g, pos, node(t, g, "A"), node(t, g, "F"), &withH)
	_, plain, _ := graphs.TraceAStarGraph(g, nil, node(t, g, "A"), node(t, g, "F"), &without)
	if cost != plain || cost != 13 || names(g, path) != "A C D F" {
		t.Errorf("A* %s cost %d (plain %d)", names(g, path), cost, plain)
	}
	if withH.Count(trace.Visit) > without.Count(trace.Visit) {
		t.Error("the heuristic explored more nodes than none")
	}
	bp, bc, _ := graphs.BeamSearch(g, pos, node(t, g, "A"), node(t, g, "F"), 1)
	if bp == nil || bc < cost {
		t.Errorf("beam %v cost %d", bp, bc)
	}
	if _, err := graphs.ParsePositions(g, "# pos Z 1 1"); !errors.Is(err, graphs.ErrNode) {
		t.Errorf("unknown node: %v", err)
	}
}

func canonical(comps [][]int) [][]int {
	out := make([][]int, len(comps))
	for i, c := range comps {
		out[i] = append([]int(nil), c...)
		sort.Ints(out[i])
	}
	sort.Slice(out, func(i, j int) bool { return out[i][0] < out[j][0] })
	return out
}

func TestKosarajuMatchesTarjan(t *testing.T) {
	for seed := int64(1); seed <= 30; seed++ {
		g, _ := graphs.Generate(graphs.Directed, 15, seed)
		if a, b := canonical(graphs.Kosaraju(g)), canonical(graphs.TarjanSCC(g)); !reflect.DeepEqual(a, b) {
			t.Fatalf("seed %d: Kosaraju %v, Tarjan %v", seed, a, b)
		}
	}
}

func TestBridgesAndArticulationPoints(t *testing.T) {
	g := mustParse(t, "A B\nB C\nC A\nC D\nD E\nE F\nF D\nB G")
	points, _ := graphs.ArticulationPoints(g)
	bridges, _ := graphs.Bridges(g)
	if names(g, points) != "B C D" {
		t.Errorf("points %s", names(g, points))
	}
	var got []string
	for _, i := range bridges {
		e := g.Edges[i]
		got = append(got, g.Names[e.From]+g.Names[e.To])
	}
	if !reflect.DeepEqual(got, []string{"CD", "BG"}) {
		t.Errorf("bridges %v", got)
	}
	twin := mustParse(t, "A B\nA B\nB C")
	if b, _ := graphs.Bridges(twin); len(b) != 1 {
		t.Errorf("parallel edges count as a bridge: %v", b)
	}
	if _, err := graphs.Bridges(mustParse(t, "directed\nA B")); !errors.Is(err, graphs.ErrDirected) {
		t.Error("directed graph accepted")
	}
}

func TestEulerPath(t *testing.T) {
	konigsberg := mustParse(t, "N I\nN I\nN E\nS I\nS I\nS E\nI E")
	if _, _, err := graphs.EulerPath(konigsberg); !errors.Is(err, graphs.ErrNoEuler) {
		t.Errorf("Königsberg has an Euler path? %v", err)
	}
	for _, text := range []string{"A B\nB C\nC A\nC D\nD A", "directed\nA B\nB C\nC A\nA D", "A B\nB C\nC D\nD A"} {
		g := mustParse(t, text)
		nodes, edges, err := graphs.EulerPath(g)
		if err != nil || len(edges) != len(g.Edges) || len(nodes) != len(edges)+1 {
			t.Fatalf("%q: %v %v %v", text, nodes, edges, err)
		}
		used := map[int]bool{}
		for i, e := range edges {
			used[e] = true
			ed := g.Edges[e]
			a, b := nodes[i], nodes[i+1]
			forward := ed.From == a && ed.To == b
			backward := !g.Directed && ed.From == b && ed.To == a
			if !forward && !backward {
				t.Fatalf("%q: step %d does not follow edge %d", text, i, e)
			}
		}
		if len(used) != len(g.Edges) {
			t.Fatalf("%q: an edge was reused", text)
		}
	}
	if _, _, err := graphs.EulerPath(mustParse(t, "A B\nC D")); !errors.Is(err, graphs.ErrNoEuler) {
		t.Error("disconnected edges accepted")
	}
}

func TestHamiltonianAndTours(t *testing.T) {
	g := mustParse(t, "A B 2\nA C 9\nA D 10\nB C 6\nB D 4\nC D 8\nA E 3\nB E 7\nC E 5\nD E 1")
	p, _ := graphs.HamiltonianPath(g, true)
	if len(p) != g.Len()+1 || p[0] != p[len(p)-1] {
		t.Errorf("cycle %v", p)
	}
	if p, _ := graphs.HamiltonianPath(mustParse(t, "A B\nA C\nA D"), false); p != nil {
		t.Errorf("a star has a Hamiltonian path? %v", p)
	}
	for seed := int64(1); seed <= 15; seed++ {
		r, _ := graphs.Generate(graphs.Connected, 7, seed)
		complete := graphs.New(false)
		for _, n := range r.Names {
			complete.Node(n)
		}
		for i := 0; i < r.Len(); i++ {
			for j := i + 1; j < r.Len(); j++ {
				complete.Edges = append(complete.Edges, graphs.Edge{From: i, To: j, Weight: 1 + (i*7+j*13+int(seed))%20})
			}
		}
		tour, cost, err := graphs.HeldKarp(complete)
		if err != nil || cost != bruteTour(complete) || len(tour) != complete.Len()+1 {
			t.Fatalf("seed %d: Held-Karp %v %d, brute force %d, %v", seed, tour, cost, bruteTour(complete), err)
		}
		_, nn, _ := graphs.NearestNeighborTour(complete, 0)
		if nn < cost {
			t.Fatalf("nearest neighbor %d beat the optimum %d", nn, cost)
		}
	}
	if _, _, err := graphs.HeldKarp(mustParse(t, "A B\nB C")); !errors.Is(err, graphs.ErrNoTour) {
		t.Error("a path graph has a tour?")
	}
}

func bruteTour(g *graphs.Graph) int {
	n := g.Len()
	w := make([][]int, n)
	for i := range w {
		w[i] = make([]int, n)
	}
	for _, e := range g.Edges {
		w[e.From][e.To], w[e.To][e.From] = e.Weight, e.Weight
	}
	perm := make([]int, n-1)
	for i := range perm {
		perm[i] = i + 1
	}
	best := math.MaxInt
	var rec func(k int)
	rec = func(k int) {
		if k == len(perm) {
			c, at := 0, 0
			for _, v := range perm {
				c += w[at][v]
				at = v
			}
			best = min(best, c+w[at][0])
			return
		}
		for i := k; i < len(perm); i++ {
			perm[k], perm[i] = perm[i], perm[k]
			rec(k + 1)
			perm[k], perm[i] = perm[i], perm[k]
		}
	}
	rec(0)
	return best
}

func TestBoruvkaDinicJohnson(t *testing.T) {
	for seed := int64(1); seed <= 30; seed++ {
		g, _ := graphs.Generate(graphs.Connected, 14, seed)
		_, kw, _ := graphs.Kruskal(g)
		edges, bw, err := graphs.Boruvka(g)
		if err != nil || bw != kw || len(edges) != g.Len()-1 {
			t.Fatalf("seed %d: Boruvka %d (%d edges), Kruskal %d", seed, bw, len(edges), kw)
		}
		d, _ := graphs.Generate(graphs.Directed, 10, seed)
		ek, _, _ := graphs.MaxFlow(d, 0, d.Len()-1)
		di, _, _ := graphs.Dinic(d, 0, d.Len()-1)
		if ek != di {
			t.Fatalf("seed %d: Dinic %d, Edmonds-Karp %d", seed, di, ek)
		}
		fw, _ := graphs.FloydWarshall(d)
		jo, err := graphs.Johnson(d)
		if err != nil || !reflect.DeepEqual(fw, jo) {
			t.Fatalf("seed %d: Johnson differs from Floyd-Warshall: %v", seed, err)
		}
	}
	neg := mustParse(t, clrsBellman)
	fw, _ := graphs.FloydWarshall(neg)
	jo, _ := graphs.Johnson(neg)
	if !reflect.DeepEqual(fw, jo) {
		t.Error("Johnson differs from Floyd-Warshall with negative weights")
	}
	if _, err := graphs.Johnson(mustParse(t, "directed\nA B 1\nB A -3")); !errors.Is(err, graphs.ErrNegativeCycle) {
		t.Errorf("negative cycle: %v", err)
	}
}

func TestHopcroftKarp(t *testing.T) {
	g := mustParse(t, "L1 R1\nL1 R2\nL2 R1\nL3 R2\nL3 R3\nL4 R3")
	m, err := graphs.HopcroftKarp(g)
	if err != nil || len(m) != 3 {
		t.Fatalf("matching %v, %v", m, err)
	}
	seen := map[int]bool{}
	for _, i := range m {
		e := g.Edges[i]
		if seen[e.From] || seen[e.To] {
			t.Fatal("two matched edges share a node")
		}
		seen[e.From], seen[e.To] = true, true
	}
	if _, err := graphs.HopcroftKarp(mustParse(t, "A B\nB C\nC A")); !errors.Is(err, graphs.ErrNotBipartite) {
		t.Error("a triangle was matched")
	}
}

func TestClosurePageRankKCore(t *testing.T) {
	g := mustParse(t, "directed\nA B\nB C\nD A")
	reach := graphs.TransitiveClosure(g)
	if reach[3][2] != 1 || reach[2][0] != 0 || reach[0][0] != 0 {
		t.Errorf("closure %v", reach)
	}
	web := mustParse(t, "directed\nA B\nA C\nB C\nC A\nD C")
	rank, ran := graphs.PageRank(web, 0.85, 100, 1e-9)
	sum := 0.0
	for _, r := range rank {
		sum += r
	}
	c, _ := web.Lookup("C")
	d, _ := web.Lookup("D")
	if math.Abs(sum-1) > 1e-9 || ran > 100 || rank[c] <= rank[d] {
		t.Errorf("ranks %v sum %v in %d iterations", rank, sum, ran)
	}
	k := mustParse(t, "A B\nA C\nA D\nB C\nB D\nC D\nD E\nE F")
	core := graphs.KCore(k)
	if want := []int{3, 3, 3, 3, 1, 1}; !reflect.DeepEqual(core, want) {
		t.Errorf("core numbers %v, want %v", core, want)
	}
}

func TestStoerWagner(t *testing.T) {
	// The example of Stoer and Wagner's paper: the minimum cut is 4.
	g := mustParse(t, "1 2 2\n1 5 3\n2 3 3\n2 5 2\n2 6 2\n3 4 4\n3 7 2\n4 7 2\n4 8 2\n5 6 3\n6 7 1\n7 8 3")
	cut, side, err := graphs.StoerWagner(g)
	if err != nil || cut != 4 || len(side) == 0 || len(side) == g.Len() {
		t.Fatalf("cut %d side %v %v", cut, side, err)
	}
	if c, _, _ := graphs.StoerWagner(mustParse(t, "A B 3\nC D 2")); c != 0 {
		t.Errorf("disconnected graph cut %d", c)
	}
}

func TestDAGPaths(t *testing.T) {
	// CLRS figure 24.5.
	g := mustParse(t, "directed\nr s 5\nr t 3\ns t 2\ns x 6\nt x 7\nt y 4\nt z 2\nx y -1\nx z 1\ny z -2")
	dist, _, err := graphs.DAGShortestPaths(g, node(t, g, "s"))
	if err != nil {
		t.Fatal(err)
	}
	if want := []int{graphs.Infinity, 0, 2, 6, 5, 3}; !reflect.DeepEqual(dist, want) {
		t.Errorf("distances %v, want %v", dist, want)
	}
	path, length, _ := graphs.DAGLongestPath(mustParse(t, "directed\nstart design 3\ndesign build 5\nstart buy 1\nbuy build 2\nbuild test 4"))
	if length != 12 || len(path) != 4 {
		t.Errorf("critical path %v length %d", path, length)
	}
	if _, _, err := graphs.DAGShortestPaths(mustParse(t, "directed\nA B\nB A"), 0); !errors.Is(err, graphs.ErrCycle) {
		t.Errorf("cycle: %v", err)
	}
}
