package databases

import (
	"fmt"
	"slices"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// PrecedenceEdge is a conflict: an operation of From came before a
// conflicting operation of To on Item, so From must come first in any
// equivalent serial order.
type PrecedenceEdge struct {
	From, To int
	Item     string
	Why      string // "W1(x) before R2(x)"
}

// PrecedenceGraph is the conflict graph of a schedule.
type PrecedenceGraph struct {
	Txns  []string
	Edges []PrecedenceEdge
	// Serializable is whether the graph has no cycle. Order is then an
	// equivalent serial order (indexes into Txns); otherwise Cycle is one
	// cycle of transactions.
	Serializable bool
	Order        []int
	Cycle        []int
}

// Precedence tests a schedule for conflict serializability. Two operations
// conflict when they belong to different transactions, touch the same item,
// and at least one writes; each conflict, in the order the schedule runs
// them, is an edge of the precedence graph. The schedule is equivalent to a
// serial one exactly when the graph has no cycle, and a topological order of
// the graph is that serial order. Transactions that abort are left out.
//
// Each edge found is reported to r as a Mark "edge" with its index, each
// transaction placed in the serial order as a Mark "order", and a cycle as a
// Mark "cycle". Time: O(n^2) for n operations.
func Precedence(ops []TxOp, r trace.Recorder) PrecedenceGraph {
	aborted := map[string]bool{}
	for _, o := range ops {
		if o.Action == Abort {
			aborted[o.Txn] = true
		}
	}
	var g PrecedenceGraph
	index := map[string]int{}
	for _, t := range txnOrder(ops) {
		if !aborted[t] {
			index[t] = len(g.Txns)
			g.Txns = append(g.Txns, t)
		}
	}
	seen := map[[2]int]bool{}
	for i, a := range ops {
		if aborted[a.Txn] || (a.Action != Read && a.Action != Write) {
			continue
		}
		for _, b := range ops[i+1:] {
			if aborted[b.Txn] || b.Txn == a.Txn || b.Item != a.Item || (b.Action != Read && b.Action != Write) {
				continue
			}
			if a.Action == Read && b.Action == Read {
				continue
			}
			e := [2]int{index[a.Txn], index[b.Txn]}
			if seen[e] {
				continue
			}
			seen[e] = true
			g.Edges = append(g.Edges, PrecedenceEdge{From: e[0], To: e[1], Item: a.Item, Why: fmt.Sprintf("%s before %s", a, b)})
			r.Record(trace.Event{Kind: trace.Mark, Indices: []int{len(g.Edges) - 1}, Note: "edge"})
		}
	}
	// Kahn's algorithm: repeatedly take a transaction nothing must precede.
	in := make([]int, len(g.Txns))
	for _, e := range g.Edges {
		in[e.To]++
	}
	var ready []int
	for t, d := range in {
		if d == 0 {
			ready = append(ready, t)
		}
	}
	for len(ready) > 0 {
		t := ready[0]
		ready = ready[1:]
		g.Order = append(g.Order, t)
		r.Record(trace.Event{Kind: trace.Mark, Indices: []int{t}, Note: "order"})
		for _, e := range g.Edges {
			if e.From == t {
				if in[e.To]--; in[e.To] == 0 {
					ready = append(ready, e.To)
				}
			}
		}
	}
	g.Serializable = len(g.Order) == len(g.Txns)
	if !g.Serializable {
		g.Order = nil
		g.Cycle = findCycle(len(g.Txns), g.Edges)
		r.Record(trace.Event{Kind: trace.Mark, Indices: slices.Clone(g.Cycle), Note: "cycle"})
	}
	return g
}

func findCycle(n int, edges []PrecedenceEdge) []int {
	state := make([]int, n)
	var path, found []int
	var dfs func(u int) bool
	dfs = func(u int) bool {
		state[u] = 1
		path = append(path, u)
		for _, e := range edges {
			if e.From != u {
				continue
			}
			if state[e.To] == 1 {
				found = slices.Clone(path[slices.Index(path, e.To):])
				return true
			}
			if state[e.To] == 0 && dfs(e.To) {
				return true
			}
		}
		path = path[:len(path)-1]
		state[u] = 2
		return false
	}
	for u := 0; u < n; u++ {
		if state[u] == 0 && dfs(u) {
			return found
		}
	}
	return nil
}
