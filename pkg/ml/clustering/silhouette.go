package clustering

import "fmt"

// Silhouette scores a clustering from -1 to 1 (Peter Rousseeuw, 1987). For
// each point, a is its mean distance to the rest of its cluster and b its
// mean distance to the nearest other cluster; its score is (b - a) /
// max(a, b). Near 1 means points sit well inside their clusters, near 0 on a
// border, below 0 in the wrong one. It returns the mean over the points and
// each point's score; a point alone in its cluster scores 0. Labels below 0
// (noise) are left out.
//
// Time: O(n^2 * d).
// Space: O(n).
func Silhouette(pts [][]float64, labels []int) (float64, []float64, error) {
	if err := check(pts); err != nil {
		return 0, nil, err
	}
	if len(labels) != len(pts) {
		return 0, nil, fmt.Errorf("%w: %d labels for %d points", ErrInput, len(labels), len(pts))
	}
	clusters := map[int]bool{}
	for _, l := range labels {
		if l >= 0 {
			clusters[l] = true
		}
	}
	if len(clusters) < 2 {
		return 0, nil, fmt.Errorf("%w: the silhouette needs at least two clusters", ErrInput)
	}
	scores := make([]float64, len(pts))
	sum, counted := 0.0, 0
	for i, p := range pts {
		if labels[i] < 0 {
			continue
		}
		total := map[int]float64{}
		count := map[int]int{}
		for j, q := range pts {
			if j == i || labels[j] < 0 {
				continue
			}
			total[labels[j]] += Distance(p, q)
			count[labels[j]]++
		}
		counted++
		if count[labels[i]] == 0 {
			continue
		}
		a := total[labels[i]] / float64(count[labels[i]])
		b := -1.0
		for c := range clusters {
			if c == labels[i] || count[c] == 0 {
				continue
			}
			if m := total[c] / float64(count[c]); b < 0 || m < b {
				b = m
			}
		}
		if s := max(a, b); s > 0 {
			scores[i] = (b - a) / s
		}
		sum += scores[i]
	}
	return sum / float64(counted), scores, nil
}
