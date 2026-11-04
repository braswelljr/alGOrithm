package randomized_test

import (
	"math"
	"math/rand"
	"reflect"
	"sort"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/math/randomized"
)

func TestShuffleIsUniform(t *testing.T) {
	counts := map[[3]int]int{}
	for seed := int64(0); seed < 6000; seed++ {
		s := []int{0, 1, 2}
		randomized.Shuffle(s, seed)
		counts[[3]int(s)]++
	}
	var c []int
	for _, v := range counts {
		c = append(c, v)
	}
	stat, df := randomized.ChiSquare(c)
	if len(c) != 6 || df != 5 || stat > 25 {
		t.Errorf("shuffle of 3: %d orders, chi-square %.1f with %d df", len(c), stat, df)
	}
}

func TestChiSquare(t *testing.T) {
	if s, df := randomized.ChiSquare([]int{10, 10, 10}); s != 0 || df != 2 {
		t.Errorf("even counts: %v, %d", s, df)
	}
	if s, _ := randomized.ChiSquare([]int{30, 0, 0}); s != 60 {
		t.Errorf("one bucket: %v", s)
	}
	if s, df := randomized.ChiSquare([]int{5}); s != 0 || df != 0 {
		t.Error("one bucket has no test")
	}
	if s, df := randomized.ChiSquare([]int{0, 0}); s != 0 || df != 0 {
		t.Error("no counts has no test")
	}
}

func TestReservoirSampling(t *testing.T) {
	s := make([]int, 10)
	for i := range s {
		s[i] = i
	}
	counts := make([]int, 10)
	for seed := int64(0); seed < 5000; seed++ {
		got := randomized.ReservoirSampling(s, 3, seed)
		if len(got) != 3 {
			t.Fatalf("sample of %d", len(got))
		}
		seen := map[int]bool{}
		for _, v := range got {
			if seen[v] {
				t.Fatalf("repeated %d in %v", v, got)
			}
			seen[v] = true
			counts[v]++
		}
	}
	if stat, df := randomized.ChiSquare(counts); stat > float64(df)+6*math.Sqrt(2*float64(df)) {
		t.Errorf("reservoir counts %v: chi-square %.1f", counts, stat)
	}
	if got := randomized.ReservoirSampling(s, 20, 1); !reflect.DeepEqual(got, s) {
		t.Errorf("k beyond n = %v", got)
	}
	if randomized.ReservoirSampling(s, 0, 1) != nil {
		t.Error("k = 0")
	}
}

func TestMonteCarloPi(t *testing.T) {
	est, pts := randomized.MonteCarloPi(200000, 3)
	if math.Abs(est-math.Pi) > 0.02 || len(pts) != 2000 {
		t.Errorf("estimate %v with %d samples", est, len(pts))
	}
	for _, p := range pts {
		if (p.X*p.X+p.Y*p.Y <= 1) != p.Inside {
			t.Fatalf("sample %+v", p)
		}
	}
	if e, p := randomized.MonteCarloPi(0, 1); e != 0 || p != nil {
		t.Error("n = 0")
	}
}

func TestKargerMinCut(t *testing.T) {
	// two 4-cliques joined by two edges: the minimum cut is 2
	var edges [][2]int
	for _, base := range []int{0, 4} {
		for i := 0; i < 4; i++ {
			for j := i + 1; j < 4; j++ {
				edges = append(edges, [2]int{base + i, base + j})
			}
		}
	}
	edges = append(edges, [2]int{0, 4}, [2]int{3, 7})
	for seed := int64(0); seed < 20; seed++ {
		cut, side, err := randomized.KargerMinCut(8, edges, 100, seed)
		if err != nil || cut != 2 {
			t.Fatalf("seed %d: cut %d, %v", seed, cut, err)
		}
		for i := range side {
			if side[i] != (i < 4) {
				t.Fatalf("seed %d: sides %v", seed, side)
			}
		}
	}
	if cut, _, _ := randomized.KargerMinCut(4, [][2]int{{0, 1}, {2, 3}}, 5, 1); cut != 0 {
		t.Errorf("disconnected cut = %d", cut)
	}
	if cut, side, _ := randomized.KargerMinCut(5, [][2]int{{0, 1}, {1, 2}}, 5, 1); cut != 0 || side[0] == side[4] {
		t.Errorf("three components: cut %d, sides %v", cut, side)
	}
	if _, _, err := randomized.KargerMinCut(1, nil, 1, 1); err == nil {
		t.Error("one node")
	}
	if _, _, err := randomized.KargerMinCut(2, [][2]int{{0, 2}}, 1, 1); err == nil {
		t.Error("edge out of range")
	}
}

func TestRandomizedQuickselect(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for n := 0; n < 60; n++ {
		s := make([]int, n)
		for i := range s {
			s[i] = r.Intn(15)
		}
		sorted := append([]int(nil), s...)
		sort.Ints(sorted)
		for k := -1; k <= n; k++ {
			got, ok := randomized.RandomizedQuickselect(s, k, int64(k))
			if want := k >= 0 && k < n; ok != want || (ok && got != sorted[k]) {
				t.Fatalf("RandomizedQuickselect(%v, %d) = %d, %v", s, k, got, ok)
			}
		}
	}
}
