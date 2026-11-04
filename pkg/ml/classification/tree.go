package classification

import (
	"fmt"
	"sort"
	"strings"
)

// Node is a decision tree: a leaf has a Label; an inner node sends a row
// Left when row[Feature] <= Threshold, else Right.
type Node struct {
	Feature     int
	Threshold   float64
	Left, Right *Node
	Label       string
	Samples     int
	Gini        float64
}

// Leaf reports whether n is a leaf.
func (n *Node) Leaf() bool { return n.Left == nil }

// BuildTree grows a classification tree by CART (Leo Breiman and others,
// 1984): at each node, try every feature and every threshold between two
// neighboring values, keep the split whose two sides have the lowest
// weighted Gini impurity (the chance two random examples of a side differ),
// and recurse, until a node is pure, has fewer than two examples, or is at
// maxDepth. It is greedy: a split that helps only together with the next one
// is never taken, so on XOR, where no single split lowers the impurity, the
// tree stays a single leaf.
//
// Time: O(d * n^2 * depth) as written (each threshold is scored by a scan).
// Space: O(n) per level.
func BuildTree(xs [][]float64, labels []string, maxDepth int) (*Node, error) {
	if err := check(xs, labels); err != nil {
		return nil, err
	}
	if maxDepth < 0 || maxDepth > 30 {
		return nil, fmt.Errorf("%w: max depth must be from 0 to 30", ErrInput)
	}
	idx := make([]int, len(xs))
	for i := range idx {
		idx[i] = i
	}
	return grow(xs, labels, idx, maxDepth), nil
}

func gini(labels []string, idx []int) float64 {
	count := map[string]int{}
	for _, i := range idx {
		count[labels[i]]++
	}
	g := 1.0
	for _, c := range count {
		p := float64(c) / float64(len(idx))
		g -= p * p
	}
	return g
}

func grow(xs [][]float64, labels []string, idx []int, depth int) *Node {
	ls := make([]string, len(idx))
	for k, i := range idx {
		ls[k] = labels[i]
	}
	n := &Node{Label: vote(ls), Samples: len(idx), Gini: gini(labels, idx)}
	if depth == 0 || n.Gini == 0 || len(idx) < 2 {
		return n
	}
	best, bf, bt := n.Gini, -1, 0.0
	for f := range xs[0] {
		vals := make([]float64, 0, len(idx))
		for _, i := range idx {
			vals = append(vals, xs[i][f])
		}
		sort.Float64s(vals)
		for k := 1; k < len(vals); k++ {
			if vals[k] == vals[k-1] {
				continue
			}
			t := (vals[k] + vals[k-1]) / 2
			l, r := split(xs, idx, f, t)
			w := (float64(len(l))*gini(labels, l) + float64(len(r))*gini(labels, r)) / float64(len(idx))
			if w < best-1e-12 {
				best, bf, bt = w, f, t
			}
		}
	}
	if bf < 0 {
		return n
	}
	l, r := split(xs, idx, bf, bt)
	n.Feature, n.Threshold = bf, bt
	n.Left, n.Right = grow(xs, labels, l, depth-1), grow(xs, labels, r, depth-1)
	return n
}

func split(xs [][]float64, idx []int, f int, t float64) (l, r []int) {
	for _, i := range idx {
		if xs[i][f] <= t {
			l = append(l, i)
		} else {
			r = append(r, i)
		}
	}
	return l, r
}

// Predict follows row down the tree to a leaf and returns its label.
func (n *Node) Predict(row []float64) string {
	for !n.Leaf() {
		if row[n.Feature] <= n.Threshold {
			n = n.Left
		} else {
			n = n.Right
		}
	}
	return n.Label
}

// Lines draws the tree as indented text, one node a line.
func (n *Node) Lines() []string {
	var out []string
	var walk func(n *Node, depth int, prefix string)
	walk = func(n *Node, depth int, prefix string) {
		pad := strings.Repeat("  ", depth)
		if n.Leaf() {
			out = append(out, fmt.Sprintf("%s%s%s (%d examples, gini %.2f)", pad, prefix, n.Label, n.Samples, n.Gini))
			return
		}
		out = append(out, fmt.Sprintf("%s%sx%d <= %.3g? (%d examples, gini %.2f)", pad, prefix, n.Feature, n.Threshold, n.Samples, n.Gini))
		walk(n.Left, depth+1, "yes: ")
		walk(n.Right, depth+1, "no: ")
	}
	walk(n, 0, "")
	return out
}
