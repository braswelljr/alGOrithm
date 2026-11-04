package clustering

import (
	"errors"
	"math"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// blobs are three well-separated groups of four points.
var blobs = [][]float64{
	{0, 0}, {1, 0}, {0, 1}, {1, 1},
	{10, 10}, {11, 10}, {10, 11}, {11, 11},
	{20, 0}, {21, 0}, {20, 1}, {21, 1},
}

// sameGroups reports whether labels split points 0-3, 4-7 and 8-11 apart.
func sameGroups(t *testing.T, labels []int) {
	t.Helper()
	for g := 0; g < 3; g++ {
		for i := 1; i < 4; i++ {
			if labels[4*g+i] != labels[4*g] {
				t.Fatalf("blob %d split: %v", g, labels)
			}
		}
	}
	if labels[0] == labels[4] || labels[4] == labels[8] || labels[0] == labels[8] {
		t.Fatalf("blobs merged: %v", labels)
	}
}

func TestKMeans(t *testing.T) {
	for seed := int64(0); seed < 5; seed++ {
		var log trace.Log
		res, err := TraceKMeans(blobs, 3, seed, &log)
		if err != nil {
			t.Fatal(err)
		}
		sameGroups(t, res.Labels)
		if math.Abs(res.Inertia-12*0.5) > 1e-9 {
			t.Errorf("seed %d: inertia %v, want 6", seed, res.Inertia)
		}
		if log.Count(trace.Mark) == 0 {
			t.Error("no group events")
		}
	}
	a, _ := KMeans(blobs, 3, 7)
	b, _ := KMeans(blobs, 3, 7)
	if a.Inertia != b.Inertia || a.Iterations != b.Iterations {
		t.Error("the same seed gave different runs")
	}
	if _, err := KMeans(blobs, 0, 1); !errors.Is(err, ErrInput) {
		t.Error("k = 0 accepted")
	}
	if _, err := KMeans([][]float64{{1}, {1, 2}}, 1, 1); !errors.Is(err, ErrInput) {
		t.Error("mixed dimensions accepted")
	}
}

func TestDBSCAN(t *testing.T) {
	pts := append(append([][]float64(nil), blobs...), []float64{50, 50})
	labels, err := DBSCAN(pts, 1.5, 3)
	if err != nil {
		t.Fatal(err)
	}
	sameGroups(t, labels[:12])
	if labels[12] != Noise {
		t.Errorf("the outlier got %d, want noise", labels[12])
	}
}

func TestAgglomerative(t *testing.T) {
	for _, l := range []Linkage{Single, Complete, Average} {
		merges, err := Agglomerative(blobs, l)
		if err != nil {
			t.Fatal(err)
		}
		if len(merges) != len(blobs)-1 || merges[len(merges)-1].Size != len(blobs) {
			t.Fatalf("linkage %d: %d merges", l, len(merges))
		}
		labels, err := Cut(merges, len(blobs), 3)
		if err != nil {
			t.Fatal(err)
		}
		sameGroups(t, labels)
		if len(Dendrogram(merges, len(blobs))) != len(merges) {
			t.Error("dendrogram lines")
		}
	}
}

func TestSilhouette(t *testing.T) {
	good := []int{0, 0, 0, 0, 1, 1, 1, 1, 2, 2, 2, 2}
	bad := []int{0, 1, 2, 0, 1, 2, 0, 1, 2, 0, 1, 2}
	g, _, err := Silhouette(blobs, good)
	if err != nil {
		t.Fatal(err)
	}
	b, _, _ := Silhouette(blobs, bad)
	if g < 0.9 || b > 0 {
		t.Errorf("good %v, bad %v", g, b)
	}
}

func TestParse(t *testing.T) {
	pts, err := Parse("# x y\n1 2\n3,4\n")
	if err != nil || len(pts) != 2 || pts[1][1] != 4 {
		t.Fatalf("%v %v", pts, err)
	}
	if _, err := Parse("1 x"); !errors.Is(err, ErrInput) {
		t.Error("bad number accepted")
	}
}
