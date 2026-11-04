package lookup

import (
	"fmt"
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// UnionFind is a disjoint-set forest: elements split into groups, with
// near-constant-time "which group?" (Find) and "merge these groups" (Union),
// thanks to union by rank and path compression. It is how Kruskal's algorithm
// detects cycles. The zero UnionFind is not ready; use NewUnionFind.
type UnionFind[T comparable] struct {
	parent map[T]T
	rank   map[T]int
	order  []T // insertion order, for a stable snapshot
	groups int
}

// NewUnionFind returns a forest with each of elements in a group of its own.
func NewUnionFind[T comparable](elements ...T) *UnionFind[T] {
	u := &UnionFind[T]{parent: map[T]T{}, rank: map[T]int{}}
	for _, e := range elements {
		u.Add(e)
	}
	return u
}

// Add makes x a group of its own if it is new, and reports whether it was.
// Time: O(1).
func (u *UnionFind[T]) Add(x T) bool {
	if _, ok := u.parent[x]; ok {
		return false
	}
	u.parent[x] = x
	u.order = append(u.order, x)
	u.groups++
	return true
}

// Groups is the number of groups. Time: O(1).
func (u *UnionFind[T]) Groups() int { return u.groups }

// Find returns the representative of x's group; ok is false when x was never
// added. It compresses the path it walks. Time: O(alpha(n)) amortized,
// effectively constant.
func (u *UnionFind[T]) Find(x T) (root T, ok bool) {
	p, ok := u.parent[x]
	if !ok {
		return root, false
	}
	if p == x {
		return x, true
	}
	root, _ = u.Find(p)
	u.parent[x] = root
	return root, true
}

// Union merges the groups of a and b (adding either if new) and reports
// whether they were separate. Time: O(alpha(n)) amortized.
func (u *UnionFind[T]) Union(a, b T) bool {
	u.Add(a)
	u.Add(b)
	ra, _ := u.Find(a)
	rb, _ := u.Find(b)
	if ra == rb {
		return false
	}
	if u.rank[ra] < u.rank[rb] {
		ra, rb = rb, ra
	}
	u.parent[rb] = ra // the shorter tree goes under the taller
	if u.rank[ra] == u.rank[rb] {
		u.rank[ra]++
	}
	u.groups--
	return true
}

// Connected reports whether a and b are in the same group.
func (u *UnionFind[T]) Connected(a, b T) bool {
	ra, oka := u.Find(a)
	rb, okb := u.Find(b)
	return oka && okb && ra == rb
}

// Snapshot describes the forest as stored (nothing is compressed by drawing
// it): each element under its parent, each representative a "root".
func (u *UnionFind[T]) Snapshot() trace.Snapshot {
	elems := append([]T(nil), u.order...)
	sort.SliceStable(elems, func(i, j int) bool { return fmt.Sprint(elems[i]) < fmt.Sprint(elems[j]) })
	id := make(map[T]int, len(elems))
	for i, x := range elems {
		id[x] = i
	}
	s := trace.Snapshot{Layout: "tree", Nodes: make([]trace.Node, len(elems))}
	for i, x := range elems {
		n := trace.Node{ID: i, Label: fmt.Sprint(x), Parent: -1}
		if p := u.parent[x]; p == x {
			n.Role = "root"
		} else {
			n.Parent = id[p]
			s.Edges = append(s.Edges, trace.Edge{From: id[p], To: i})
		}
		s.Nodes[i] = n
	}
	return s
}
