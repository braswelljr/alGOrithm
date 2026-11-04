// Package graphs holds algorithms on weighted graphs, one file each:
// breadth-first and depth-first traversal, Dijkstra's and the Bellman-Ford
// shortest paths, Floyd-Warshall all-pairs distances, topological sort,
// Kruskal's and Prim's minimum spanning trees, Tarjan's strongly connected
// components, the bipartite check, Edmonds-Karp maximum flow, and cycle
// detection.
//
// A Graph has named nodes, numbered in the order they first appear, and
// weighted edges, numbered the same way; it is directed or not. Parse reads
// one edge a line:
//
//	g, _ := graphs.Parse("A B 4\nA C 1\nC B 2")
//	dist, prev, _ := graphs.Dijkstra(g, 0)
//	path := graphs.PathTo(prev, 0, 1) // A C B
//
// Every algorithm X has a TraceX that reports what it does to a
// trace.Recorder: nodes visited and queued (Visit, Push, Indices[0] the
// node), distances set (Set, Values[0] the distance), and its answer as Mark
// events: "path", "order", "group" (Values[0] the group), "cycle" over nodes,
// and "tree" and "flow" over edges (Indices[0] the edge, Values[0] the flow).
package graphs

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Limits on what Parse and Generate accept.
const (
	MaxNodes = 500
	MaxEdges = 5000
)

// Infinity is the distance to a node that cannot be reached.
const Infinity = math.MaxInt

// Errors the algorithms report.
var (
	ErrInput         = errors.New("graphs: invalid graph")
	ErrNegative      = errors.New("graphs: a negative edge weight (use Bellman-Ford)")
	ErrNegativeCycle = errors.New("graphs: a negative cycle makes some distances unbounded")
	ErrCycle         = errors.New("graphs: the graph has a cycle, so it has no topological order")
	ErrUndirected    = errors.New("graphs: the algorithm needs a directed graph")
	ErrDirected      = errors.New("graphs: the algorithm needs an undirected graph")
	ErrNode          = errors.New("graphs: no such node")
)

// Edge joins From to To (both directions when the graph is undirected).
type Edge struct {
	From, To, Weight int
}

// Graph is a set of named nodes and weighted edges. The zero Graph is an empty
// undirected graph, ready to use.
type Graph struct {
	Directed bool
	Names    []string // node i is Names[i]
	Edges    []Edge
	index    map[string]int
}

// New returns an empty graph.
func New(directed bool) *Graph { return &Graph{Directed: directed} }

// Len is the number of nodes.
func (g *Graph) Len() int { return len(g.Names) }

// Node returns the index of the node called name, adding it if it is new.
func (g *Graph) Node(name string) int {
	if i, ok := g.index[name]; ok {
		return i
	}
	if g.index == nil {
		g.index = map[string]int{}
	}
	g.index[name] = len(g.Names)
	g.Names = append(g.Names, name)
	return len(g.Names) - 1
}

// Lookup returns the index of the node called name.
func (g *Graph) Lookup(name string) (int, bool) {
	i, ok := g.index[name]
	return i, ok
}

// AddEdge joins a to b with weight w, adding the nodes if they are new, and
// returns the edge's index.
func (g *Graph) AddEdge(a, b string, w int) int {
	g.Edges = append(g.Edges, Edge{From: g.Node(a), To: g.Node(b), Weight: w})
	return len(g.Edges) - 1
}

// arc is one way along an edge.
type arc struct{ to, weight, edge int }

// adjacency lists every node's arcs, in edge order: both ways for an
// undirected graph.
func (g *Graph) adjacency() [][]arc {
	adj := make([][]arc, len(g.Names))
	for i, e := range g.Edges {
		adj[e.From] = append(adj[e.From], arc{e.To, e.Weight, i})
		if !g.Directed && e.From != e.To {
			adj[e.To] = append(adj[e.To], arc{e.From, e.Weight, i})
		}
	}
	return adj
}

// Parse reads a graph, one edge a line: "A B" or "A B 4" (the weight, 1 when
// left out; it may be negative). A line with a single name adds a node with no
// edges. A first line "directed" (or "undirected") sets the kind; the default
// is undirected. Lines starting with # are comments.
func Parse(text string) (*Graph, error) {
	g := &Graph{}
	first := true
	for i, line := range strings.Split(text, "\n") {
		f := strings.FieldsFunc(line, func(r rune) bool { return r == ' ' || r == '\t' || r == ',' || r == '\r' })
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		if first && len(f) == 1 && (strings.EqualFold(f[0], "directed") || strings.EqualFold(f[0], "undirected")) {
			g.Directed = strings.EqualFold(f[0], "directed")
			first = false
			continue
		}
		first = false
		switch len(f) {
		case 1:
			g.Node(f[0])
		case 2, 3:
			w := 1
			if len(f) == 3 {
				n, err := strconv.Atoi(f[2])
				if err != nil {
					return nil, fmt.Errorf("%w: line %d: the weight %q is not a whole number", ErrInput, i+1, f[2])
				}
				w = n
			}
			g.AddEdge(f[0], f[1], w)
		default:
			return nil, fmt.Errorf("%w: line %d: want \"FROM TO [WEIGHT]\"", ErrInput, i+1)
		}
		if len(g.Names) > MaxNodes || len(g.Edges) > MaxEdges {
			return nil, fmt.Errorf("%w: at most %d nodes and %d edges", ErrInput, MaxNodes, MaxEdges)
		}
	}
	if len(g.Names) == 0 {
		return nil, fmt.Errorf("%w: no nodes", ErrInput)
	}
	return g, nil
}

// String writes g in the form Parse reads.
func (g *Graph) String() string {
	var b strings.Builder
	if g.Directed {
		b.WriteString("directed\n")
	}
	used := make([]bool, len(g.Names))
	for _, e := range g.Edges {
		used[e.From], used[e.To] = true, true
		fmt.Fprintf(&b, "%s %s %d\n", g.Names[e.From], g.Names[e.To], e.Weight)
	}
	for i, u := range used {
		if !u {
			fmt.Fprintf(&b, "%s\n", g.Names[i])
		}
	}
	return b.String()
}

// Kinds of generated graph.
const (
	Connected = "connected" // undirected, connected, weights 1 to 9
	Directed  = "directed"  // directed, weights 1 to 9, every node reachable from the first
	DAG       = "dag"       // directed and acyclic: edges only go from lower to higher nodes
)

// Generate makes a graph of n nodes named A, B, ... (then A1, B1, ...) from
// seed: a random spanning tree for connectivity plus about n more edges.
func Generate(kind string, n int, seed int64) (*Graph, error) {
	if n < 1 || n > 100 {
		return nil, fmt.Errorf("%w: generate 1 to 100 nodes", ErrInput)
	}
	r := rand.New(rand.NewSource(seed))
	g := &Graph{Directed: kind != Connected}
	if kind != Connected && kind != Directed && kind != DAG {
		return nil, fmt.Errorf("%w: unknown kind %q (connected, directed, dag)", ErrInput, kind)
	}
	name := func(i int) string {
		s := string(rune('A' + i%26))
		if i >= 26 {
			s += strconv.Itoa(i / 26)
		}
		return s
	}
	for i := 0; i < n; i++ {
		g.Node(name(i))
	}
	seen := map[[2]int]bool{}
	add := func(a, b int) {
		if a == b || seen[[2]int{a, b}] || (!g.Directed && seen[[2]int{b, a}]) {
			return
		}
		seen[[2]int{a, b}] = true
		g.Edges = append(g.Edges, Edge{From: a, To: b, Weight: 1 + r.Intn(9)})
	}
	for i := 1; i < n; i++ {
		p := r.Intn(i) // a node before i: a tree rooted at 0, and acyclic for a DAG
		add(p, i)
	}
	for k := 0; k < n; k++ {
		a, b := r.Intn(n), r.Intn(n)
		if kind == DAG && a > b {
			a, b = b, a
		}
		add(a, b)
	}
	return g, nil
}

// PathTo follows prev back from target and returns the nodes from source to
// target, or nil when target was not reached (prev[source] is -1).
func PathTo(prev []int, source, target int) []int {
	if target < 0 || target >= len(prev) || (target != source && prev[target] < 0) {
		return nil
	}
	path := []int{target}
	for at := target; at != source; {
		at = prev[at]
		if at < 0 || len(path) > len(prev) {
			return nil // prev does not lead back to source
		}
		path = append(path, at)
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}

func (g *Graph) checkNode(n int) error {
	if n < 0 || n >= len(g.Names) {
		return fmt.Errorf("%w: %d", ErrNode, n)
	}
	return nil
}

// rec wraps a Recorder with the events graph algorithms report.
type rec struct{ r trace.Recorder }

var nop = rec{trace.Nop{}}

func (x rec) visit(n int) { x.r.Record(trace.Event{Kind: trace.Visit, Indices: []int{n}}) }
func (x rec) push(n int)  { x.r.Record(trace.Event{Kind: trace.Push, Indices: []int{n}}) }
func (x rec) set(n, d int) {
	x.r.Record(trace.Event{Kind: trace.Set, Indices: []int{n}, Values: []int{d}})
}
func (x rec) mark(note string, indices []int, values ...int) {
	x.r.Record(trace.Event{Kind: trace.Mark, Indices: indices, Values: values, Note: note})
}

func filled(n, v int) []int {
	s := make([]int, n)
	for i := range s {
		s[i] = v
	}
	return s
}
