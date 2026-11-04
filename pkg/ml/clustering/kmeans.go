package clustering

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// KMeansResult is a k-means clustering: each point's cluster, the clusters'
// centers, how many rounds it took, and the inertia (the sum of squared
// distances from each point to its center, which k-means makes small).
type KMeansResult struct {
	Labels     []int
	Centroids  [][]float64
	Iterations int
	Inertia    float64
}

// KMeans splits pts into k clusters with Lloyd's algorithm: put every point
// with its nearest center, move every center to the mean of its points, and
// repeat until nothing moves (at most 100 rounds). The starting centers come
// from KMeansPlusPlus with seed, so a seed always gives the same answer.
// k-means finds a local optimum of the inertia, not always the best one.
//
// Time: O(n * k * d) per round, for n points of dimension d.
// Space: O(n + k * d).
func KMeans(pts [][]float64, k int, seed int64) (KMeansResult, error) {
	return kmeans(pts, k, seed, trace.Nop{})
}

// TraceKMeans is KMeans, reporting each round's clusters as Mark "group".
func TraceKMeans(pts [][]float64, k int, seed int64, r trace.Recorder) (KMeansResult, error) {
	return kmeans(pts, k, seed, r)
}

func kmeans(pts [][]float64, k int, seed int64, r trace.Recorder) (KMeansResult, error) {
	centers, err := KMeansPlusPlus(pts, k, seed)
	if err != nil {
		return KMeansResult{}, err
	}
	labels := make([]int, len(pts))
	for i := range labels {
		labels[i] = -1
	}
	res := KMeansResult{}
	for res.Iterations < 100 {
		res.Iterations++
		moved := false
		for i, p := range pts {
			best := nearest(p, centers)
			if best != labels[i] {
				labels[i], moved = best, true
			}
		}
		markGroups(r, labels)
		if !moved {
			break
		}
		for c := range centers {
			sum := make([]float64, len(pts[0]))
			n := 0
			for i, p := range pts {
				if labels[i] == c {
					n++
					for j, v := range p {
						sum[j] += v
					}
				}
			}
			if n == 0 { // an empty cluster keeps its center
				continue
			}
			for j := range sum {
				sum[j] /= float64(n)
			}
			centers[c] = sum
		}
	}
	for i, p := range pts {
		d := Distance(p, centers[labels[i]])
		res.Inertia += d * d
	}
	res.Labels, res.Centroids = labels, centers
	return res, nil
}

func nearest(p []float64, centers [][]float64) int {
	best, bestD := 0, math.Inf(1)
	for c, ctr := range centers {
		if d := Distance(p, ctr); d < bestD {
			best, bestD = c, d
		}
	}
	return best
}

// KMeansPlusPlus picks k starting centers from pts (David Arthur and Sergei
// Vassilvitskii, 2007): the first at random, each next one with probability
// proportional to its squared distance from the nearest center chosen so far,
// so the centers start spread out. The same seed gives the same centers.
//
// Time: O(n * k * d).
// Space: O(n + k * d).
func KMeansPlusPlus(pts [][]float64, k int, seed int64) ([][]float64, error) {
	if err := check(pts); err != nil {
		return nil, err
	}
	if k < 1 || k > len(pts) {
		return nil, fmt.Errorf("%w: k must be from 1 to the number of points (%d)", ErrInput, len(pts))
	}
	rng := rand.New(rand.NewSource(seed))
	centers := [][]float64{clone(pts[rng.Intn(len(pts))])}
	d2 := make([]float64, len(pts))
	for len(centers) < k {
		total := 0.0
		for i, p := range pts {
			d := Distance(p, centers[nearest(p, centers)])
			d2[i] = d * d
			total += d2[i]
		}
		pick := 0
		if total > 0 {
			x := rng.Float64() * total
			for i, w := range d2 {
				x -= w
				if x < 0 {
					pick = i
					break
				}
			}
		}
		centers = append(centers, clone(pts[pick]))
	}
	return centers, nil
}

func clone(p []float64) []float64 { return append([]float64(nil), p...) }
