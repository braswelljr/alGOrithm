package core

import (
	"errors"
	"math"
	"testing"
)

func TestFitGrowthFindsTheClass(t *testing.T) {
	sizes := []int{100, 200, 400, 800, 1600}
	for _, g := range growths {
		if g.class == "O(2^n)" {
			continue // 2^1600 overflows; checked with small sizes below
		}
		samples := make([]Sample, len(sizes))
		ys := make([]float64, len(sizes))
		for i, n := range sizes {
			samples[i] = Sample{N: n}
			ys[i] = 3*g.f(float64(n)) + 5 // a constant factor and a lower term
		}
		if got := fitGrowth(samples, ys)[0].Class; got != g.class {
			t.Errorf("3*%s+5 fitted as %s", g.class, got)
		}
	}
	small := []Sample{{N: 10}, {N: 14}, {N: 18}, {N: 22}}
	ys := []float64{1 << 10, 1 << 14, 1 << 18, 1 << 22}
	if got := fitGrowth(small, ys)[0]; got.Class != "O(2^n)" || got.Error > 0.01 {
		t.Errorf("2^n fitted as %+v", got)
	}
	if p := (Fit{Class: "O(n)", Scale: 2}).Predict(50); p != 100 {
		t.Errorf("Predict = %v", p)
	}
}

func TestStatedRank(t *testing.T) {
	for _, c := range []struct {
		stated string
		rank   float64
		ok     bool
	}{
		{"O(n log n) average, O(n^2) worst", 4, true},
		{"O(log n)", 1, true},
		{"O(V + E)", 2, true},
		{"O((V + E) log V)", 3, true},
		{"about O(n^(4/3)), O(n^2) worst", 4, true},
		{"O(n * m)", 0, false},
		{"exponential worst", 0, false},
	} {
		rank, ok := StatedRank(c.stated)
		if ok != c.ok || (ok && rank != c.rank) {
			t.Errorf("StatedRank(%q) = %v, %v; want %v, %v", c.stated, rank, ok, c.rank, c.ok)
		}
	}
}

func TestMeasureRejects(t *testing.T) {
	for _, c := range []struct {
		id    string
		sizes []int
		order string
	}{
		{"nope", nil, "random"},
		{"stack", nil, "random"},                      // a structure: one event per operation
		{"quick-sort", []int{100, 50, 200}, "random"}, // sizes must grow
		{"quick-sort", []int{100, 200}, "random"},     // too few
		{"quick-sort", nil, "zigzag"},
	} {
		if _, err := MeasureOrder(c.id, c.order, c.sizes, 1); !errors.Is(err, ErrInput) {
			t.Errorf("MeasureOrder(%s, %v, %s): err = %v, want an input error", c.id, c.sizes, c.order, err)
		}
	}
}

func TestMeasureIsRepeatable(t *testing.T) {
	a, err := Measure("merge-sort", nil, 7)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := Measure("merge-sort", nil, 7)
	for i := range a.Samples {
		if a.Samples[i].Ops != b.Samples[i].Ops {
			t.Fatalf("size %d: %d then %d ops", a.Samples[i].N, a.Samples[i].Ops, b.Samples[i].Ops)
		}
	}
	if a.Best != "O(n log n)" || a.Quality != "good" || a.Agrees == nil || !*a.Agrees {
		t.Errorf("merge sort measured as %s (%s, agrees %v)", a.Best, a.Quality, a.Agrees)
	}
}

// honestExtras are entries outside sorting and searching whose input grows
// cleanly with one size, measured with them.
var honestExtras = []string{
	"breadth-first-search", "dijkstra", "graph-bfs", "graph-dfs", "kruskal", "prim", "tarjan-scc",
	"longest-increasing-subsequence", "lis-fast", "rod-cutting", "manacher", "linear-search",
}

// TestStatedCostsAreHonest measures every traced sort and search, and the
// extras, by their operation counts at doubling sizes on seeded random
// values, and fails when one grows faster than the worst cost its entry
// states. Counts, not time, so it never flakes.
func TestStatedCostsAreHonest(t *testing.T) {
	if testing.Short() {
		t.Skip("measures dozens of entries")
	}
	ids := append([]string(nil), honestExtras...)
	for _, e := range Entries() {
		if e.Traced && (e.Category == "sorting" || e.Category == "searching") {
			ids = append(ids, e.ID)
		}
	}
	measured := 0
	for _, id := range ids {
		m, err := Measure(id, nil, 1)
		if err != nil {
			t.Logf("%s: not measured: %v", id, err) // bogo and stooge sort cap their size
			continue
		}
		measured++
		if m.Agrees != nil && !*m.Agrees {
			t.Errorf("%s grows like %s (off by %.0f%%), faster than its stated %q", id, m.Best, 100*m.Error, m.Stated)
		}
		if m.Error > 0.5 || math.IsNaN(m.Error) {
			t.Errorf("%s: no class fits (best %s, off by %.0f%%)", id, m.Best, 100*m.Error)
		}
	}
	if measured < 30 {
		t.Errorf("only %d entries measured; did the inputs stop generating?", measured)
	}
}

// TestWhyNamesRealEntries keeps every explanation attached to a real entry,
// and checks there are enough of them to be worth reading.
func TestWhyNamesRealEntries(t *testing.T) {
	for id, text := range why {
		if _, ok := Lookup(id); !ok {
			t.Errorf("why[%q]: no such entry", id)
		}
		if text == "" || text[len(text)-1] != '.' {
			t.Errorf("why[%q] should be a sentence ending in a period", id)
		}
	}
	if len(why) < 120 {
		t.Errorf("%d explanations; want at least 120", len(why))
	}
	if e, _ := Lookup("merge-sort"); e.Why == "" {
		t.Error("Entries() does not copy Why")
	}
}
