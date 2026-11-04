package ranges_test

import (
	"math/rand"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/structures/ranges"
)

func TestRangeStructures(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	vals := make([]int, 57)
	for i := range vals {
		vals[i] = r.Intn(200) - 100
	}
	seg, fen := ranges.NewSegmentTree(vals), ranges.NewFenwickTree(vals)
	for step := 0; step < 2000; step++ {
		if r.Intn(3) == 0 {
			i, v := r.Intn(len(vals)), r.Intn(200)-100
			seg.Update(i, v)
			fen.Add(i, v-vals[i])
			vals[i] = v
		}
		l := r.Intn(len(vals))
		h := l + r.Intn(len(vals)-l)
		sum, mn := 0, vals[l]
		for _, v := range vals[l : h+1] {
			sum += v
			mn = min(mn, v)
		}
		if got, ok := seg.Sum(l, h); !ok || got != sum {
			t.Fatalf("segment Sum(%d,%d) = %d, want %d", l, h, got, sum)
		}
		if got, ok := seg.Min(l, h); !ok || got != mn {
			t.Fatalf("segment Min(%d,%d) = %d, want %d", l, h, got, mn)
		}
		if got, ok := fen.RangeSum(l, h); !ok || got != sum {
			t.Fatalf("fenwick RangeSum(%d,%d) = %d, want %d", l, h, got, sum)
		}
		sp := ranges.NewSparseTable(vals)
		if got, ok := sp.Min(l, h); !ok || got != mn {
			t.Fatalf("sparse Min(%d,%d) = %d, want %d", l, h, got, mn)
		}
	}
	if _, ok := seg.Sum(5, 2); ok {
		t.Error("an inverted range succeeded")
	}
	if _, ok := fen.PrefixSum(len(vals)); ok || fen.Add(-1, 1) || seg.Update(99, 0) {
		t.Error("out-of-range access succeeded")
	}
	if _, ok := ranges.NewSparseTable(nil).Min(0, 0); ok {
		t.Error("empty sparse table answered")
	}
	if n := len(ranges.NewSegmentTree([]int{1, 2, 3}).Snapshot().Nodes); n != 5 {
		t.Errorf("segment snapshot of 3 values has %d nodes, want 5", n)
	}
	if len(ranges.NewFenwickTree([]int{1, 2, 3, 4}).Snapshot().Nodes) != 4 || len(ranges.NewSparseTable([]int{3, 1, 2}).Table()) != 2 {
		t.Error("fenwick or sparse snapshot wrong")
	}
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	if a < 0 {
		return -a
	}
	return a
}

func TestDisjointSparseTable(t *testing.T) {
	r := rand.New(rand.NewSource(5))
	ops := map[string]func(a, b int) int{
		"sum": func(a, b int) int { return a + b },
		"min": func(a, b int) int { return min(a, b) },
		"gcd": gcd,
	}
	for n := 0; n <= 40; n++ {
		values := make([]int, n)
		for i := range values {
			values[i] = r.Intn(100) - 30
		}
		for name, op := range ops {
			tab := ranges.NewDisjointSparseTable(values, op)
			for l := 0; l < n; l++ {
				for h := l; h < n; h++ {
					want := values[l]
					for i := l + 1; i <= h; i++ {
						want = op(want, values[i])
					}
					if got, ok := tab.Query(l, h); !ok || got != want {
						t.Fatalf("n=%d %s(%d..%d) = %d, %v; want %d", n, name, l, h, got, ok, want)
					}
				}
			}
			if _, ok := tab.Query(0, n); ok {
				t.Fatalf("n=%d: a range past the end was answered", n)
			}
		}
	}
}
