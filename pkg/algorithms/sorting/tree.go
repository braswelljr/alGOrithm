package sorting

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Tree returns a sorted copy of s, in ascending order, made with tree sort:
// insert every value into a binary search tree (equal values go right, which
// keeps it stable), then read the tree in order. s is not modified.
//
// Time: O(n log n) on average, O(n^2) when the input is already sorted (the
// tree becomes a chain); a balanced tree would make it O(n log n) always.
// Space: O(n) for the tree.
func Tree[T cmp.Ordered](s []T) []T { return treeSort(s, nop) }

// TraceTree is Tree, reporting each comparison while inserting (Compare with
// the indexes of the value being inserted and the node's value) and each
// value written back in order (Set). Indices refer to positions in a copy of
// s.
func TraceTree[T cmp.Ordered](s []T, r trace.Recorder) []T { return treeSort(s, rec{r}) }

type treeNode struct{ index, left, right int } // children are node indexes, -1 for none

func treeSort[T cmp.Ordered](in []T, x rec) []T {
	s := clone(in)
	if len(s) < 2 {
		return s
	}
	nodes := make([]treeNode, 0, len(s))
	for i := range s {
		nodes = append(nodes, treeNode{i, -1, -1})
		if i == 0 {
			continue
		}
		at := 0
		for {
			x.compare(i, nodes[at].index)
			link := &nodes[at].right
			if s[i] < s[nodes[at].index] {
				link = &nodes[at].left
			}
			if *link < 0 {
				*link = i
				break
			}
			at = *link
		}
	}
	vals := clone(s)
	out := 0
	var walk func(n int)
	walk = func(n int) { // an explicit stack would avoid deep recursion on sorted input
		for n >= 0 {
			walk(nodes[n].left)
			s[out] = vals[nodes[n].index]
			set(x, out, s[out])
			out++
			n = nodes[n].right
		}
	}
	walk(0)
	return s
}
