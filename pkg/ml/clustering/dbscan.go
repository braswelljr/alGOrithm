package clustering

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Noise is the label DBSCAN gives a point in no cluster.
const Noise = -1

// DBSCAN clusters pts by density (Martin Ester and others, 1996): a point with
// at least minPts points (itself included) within eps is a core point, a
// cluster is everything reachable from a core point through core points, and
// a point reachable from none is noise (label Noise). Unlike k-means it finds
// the number of clusters itself, and clusters of any shape.
//
// Time: O(n^2 * d) (every neighborhood is found by a scan).
// Space: O(n).
func DBSCAN(pts [][]float64, eps float64, minPts int) ([]int, error) {
	return dbscan(pts, eps, minPts, trace.Nop{})
}

// TraceDBSCAN is DBSCAN, reporting Visit for each point expanded and Mark
// "group" for each cluster when it is complete, and for the noise.
func TraceDBSCAN(pts [][]float64, eps float64, minPts int, r trace.Recorder) ([]int, error) {
	return dbscan(pts, eps, minPts, r)
}

func dbscan(pts [][]float64, eps float64, minPts int, r trace.Recorder) ([]int, error) {
	if err := check(pts); err != nil {
		return nil, err
	}
	if eps <= 0 || minPts < 1 {
		return nil, fmt.Errorf("%w: eps must be positive and minPts at least 1", ErrInput)
	}
	const unseen = -2
	labels := make([]int, len(pts))
	for i := range labels {
		labels[i] = unseen
	}
	neighbors := func(i int) []int {
		var out []int
		for j, q := range pts {
			if Distance(pts[i], q) <= eps {
				out = append(out, j)
			}
		}
		return out
	}
	cluster := 0
	for i := range pts {
		if labels[i] != unseen {
			continue
		}
		r.Record(trace.Event{Kind: trace.Visit, Indices: []int{i}})
		nb := neighbors(i)
		if len(nb) < minPts {
			labels[i] = Noise
			continue
		}
		labels[i] = cluster
		members := []int{i}
		queue := append([]int(nil), nb...)
		for len(queue) > 0 {
			j := queue[0]
			queue = queue[1:]
			if labels[j] == Noise {
				labels[j] = cluster // a border point
				members = append(members, j)
			}
			if labels[j] != unseen {
				continue
			}
			labels[j] = cluster
			members = append(members, j)
			r.Record(trace.Event{Kind: trace.Visit, Indices: []int{j}})
			if more := neighbors(j); len(more) >= minPts {
				queue = append(queue, more...)
			}
		}
		r.Record(trace.Event{Kind: trace.Mark, Note: "group", Indices: members, Values: []int{cluster}})
		cluster++
	}
	var noise []int
	for i, l := range labels {
		if l == Noise {
			noise = append(noise, i)
		}
	}
	if len(noise) > 0 {
		r.Record(trace.Event{Kind: trace.Mark, Note: "group", Indices: noise, Values: []int{Noise}})
	}
	return labels, nil
}
