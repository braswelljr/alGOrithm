package clustering

import (
	"fmt"
	"math"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Linkage is how the distance between two clusters is measured.
type Linkage int

const (
	// Single linkage: the distance between the closest pair of points, one
	// in each cluster. It follows chains and makes long, thin clusters.
	Single Linkage = iota
	// Complete linkage: the distance between the farthest pair. It makes
	// compact, round clusters.
	Complete
	// Average linkage: the mean distance over all pairs.
	Average
)

// Merge is one step of agglomerative clustering: the clusters A and B (by
// their numbers: points are 0 to n-1, and the merge at step s makes cluster
// n+s) joined at Distance into a cluster of Size points.
type Merge struct {
	A, B     int
	Distance float64
	Size     int
}

// Agglomerative builds a hierarchy of clusters from the bottom up: every point
// starts as its own cluster, and the two nearest clusters (by linkage) are
// joined until one is left. The n-1 merges are the dendrogram; Cut turns them
// into k clusters.
//
// Time: O(n^3) (the naive algorithm; the matrix is rescanned each merge).
// Space: O(n^2) for the distance matrix.
func Agglomerative(pts [][]float64, linkage Linkage) ([]Merge, error) {
	return agglomerative(pts, linkage, trace.Nop{})
}

// TraceAgglomerative is Agglomerative, reporting each merge as Mark "group"
// with the points of the new cluster and Values[0] its number.
func TraceAgglomerative(pts [][]float64, linkage Linkage, r trace.Recorder) ([]Merge, error) {
	return agglomerative(pts, linkage, r)
}

func agglomerative(pts [][]float64, linkage Linkage, r trace.Recorder) ([]Merge, error) {
	if err := check(pts); err != nil {
		return nil, err
	}
	if linkage < Single || linkage > Average {
		return nil, fmt.Errorf("%w: unknown linkage %d", ErrInput, linkage)
	}
	n := len(pts)
	members := map[int][]int{}
	for i := 0; i < n; i++ {
		members[i] = []int{i}
	}
	active := make([]int, 0, n)
	for i := 0; i < n; i++ {
		active = append(active, i)
	}
	dist := func(a, b []int) float64 {
		best := math.Inf(1)
		if linkage != Single {
			best = 0
		}
		for _, i := range a {
			for _, j := range b {
				d := Distance(pts[i], pts[j])
				switch linkage {
				case Single:
					best = math.Min(best, d)
				case Complete:
					best = math.Max(best, d)
				default:
					best += d
				}
			}
		}
		if linkage == Average {
			best /= float64(len(a) * len(b))
		}
		return best
	}
	var merges []Merge
	for len(active) > 1 {
		bi, bj, bd := 0, 1, math.Inf(1)
		for x := 0; x < len(active); x++ {
			for y := x + 1; y < len(active); y++ {
				if d := dist(members[active[x]], members[active[y]]); d < bd {
					bi, bj, bd = x, y, d
				}
			}
		}
		a, b := active[bi], active[bj]
		id := n + len(merges)
		members[id] = append(append([]int(nil), members[a]...), members[b]...)
		merges = append(merges, Merge{A: a, B: b, Distance: bd, Size: len(members[id])})
		r.Record(trace.Event{Kind: trace.Mark, Note: "group", Indices: members[id], Values: []int{id}})
		delete(members, a)
		delete(members, b)
		active = append(append(active[:bi:bi], active[bi+1:bj]...), active[bj+1:]...)
		active = append(active, id)
	}
	return merges, nil
}

// Cut undoes the last k-1 of the n-1 merges, leaving k clusters, and returns
// each of the n points' cluster, numbered from 0 in order of first point.
//
// Time: O(n).
// Space: O(n).
func Cut(merges []Merge, n, k int) ([]int, error) {
	if n < 1 || len(merges) != n-1 || k < 1 || k > n {
		return nil, fmt.Errorf("%w: want n-1 merges and k from 1 to n", ErrInput)
	}
	parent := make([]int, 2*n-1)
	for i := range parent {
		parent[i] = i
	}
	for s, m := range merges[:n-k] {
		parent[m.A], parent[m.B] = n+s, n+s
	}
	root := func(i int) int {
		for parent[i] != i {
			i = parent[i]
		}
		return i
	}
	labels := make([]int, n)
	number := map[int]int{}
	for i := 0; i < n; i++ {
		rt := root(i)
		if _, ok := number[rt]; !ok {
			number[rt] = len(number)
		}
		labels[i] = number[rt]
	}
	return labels, nil
}

// Dendrogram describes the merges in words, one line each, from the first
// (closest) to the last.
func Dendrogram(merges []Merge, n int) []string {
	name := func(c int) string {
		if c < n {
			return fmt.Sprintf("point %d", c)
		}
		return fmt.Sprintf("cluster %d", c)
	}
	out := make([]string, len(merges))
	for s, m := range merges {
		out[s] = fmt.Sprintf("cluster %d = %s + %s at distance %.3g (%d points)", n+s, name(m.A), name(m.B), m.Distance, m.Size)
	}
	return out
}
