// Package clustering groups points that are close together without being
// told what the groups are (unsupervised learning), one algorithm per file:
// k-means with k-means++ seeding, DBSCAN, agglomerative hierarchical
// clustering, and the silhouette score that says how good a grouping is.
//
// A point is a []float64 of any dimension; every point in one call has the
// same dimension. Distances are Euclidean. Nothing here changes its input.
//
// The Trace functions report to a trace.Recorder: Visit for a point looked
// at, and Mark "group" with Indices the points of one cluster and Values[0]
// that cluster's number (-1 for noise).
package clustering

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// ErrInput is returned for points that are missing, of mixed dimension, or
// settings out of range.
var ErrInput = errors.New("clustering: invalid input")

// MaxPoints is the most points Parse accepts.
const MaxPoints = 5_000

// Parse reads one point a line, its coordinates separated by spaces, tabs or
// commas; blank lines and lines starting with # are skipped.
func Parse(text string) ([][]float64, error) {
	var pts [][]float64
	for i, line := range strings.Split(text, "\n") {
		f := strings.FieldsFunc(line, func(r rune) bool { return r == ' ' || r == '\t' || r == ',' || r == '\r' })
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		p := make([]float64, len(f))
		for j, s := range f {
			v, err := strconv.ParseFloat(s, 64)
			if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
				return nil, fmt.Errorf("%w: line %d: %q is not a number", ErrInput, i+1, s)
			}
			p[j] = v
		}
		if len(pts) > 0 && len(p) != len(pts[0]) {
			return nil, fmt.Errorf("%w: line %d has %d coordinates, the first point has %d", ErrInput, i+1, len(p), len(pts[0]))
		}
		pts = append(pts, p)
		if len(pts) > MaxPoints {
			return nil, fmt.Errorf("%w: more than %d points", ErrInput, MaxPoints)
		}
	}
	if len(pts) == 0 {
		return nil, fmt.Errorf("%w: no points", ErrInput)
	}
	return pts, nil
}

// Distance is the Euclidean distance between a and b.
func Distance(a, b []float64) float64 {
	s := 0.0
	for i := range a {
		d := a[i] - b[i]
		s += d * d
	}
	return math.Sqrt(s)
}

func check(pts [][]float64) error {
	if len(pts) == 0 {
		return fmt.Errorf("%w: no points", ErrInput)
	}
	for i, p := range pts {
		if len(p) == 0 || len(p) != len(pts[0]) {
			return fmt.Errorf("%w: point %d has %d coordinates, the first has %d", ErrInput, i, len(p), len(pts[0]))
		}
	}
	return nil
}

// markGroups reports each cluster of labels as one Mark "group".
func markGroups(r trace.Recorder, labels []int) {
	groups := map[int][]int{}
	var order []int
	for i, l := range labels {
		if _, ok := groups[l]; !ok {
			order = append(order, l)
		}
		groups[l] = append(groups[l], i)
	}
	for _, l := range order {
		r.Record(trace.Event{Kind: trace.Mark, Note: "group", Indices: groups[l], Values: []int{l}})
	}
}
