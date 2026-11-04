package sets_test

import (
	"math/rand"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/structures/sets"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func randomSet(r *rand.Rand, n, span int) []int {
	m := map[int]bool{}
	for len(m) < n {
		m[r.Intn(span)] = true
	}
	out := make([]int, 0, n)
	for v := range m {
		out = append(out, v)
	}
	slices.Sort(out)
	return out
}

// model computes the four operations with maps.
func model(a, b []int) (union, inter, diff, sym []int) {
	ina, inb := map[int]bool{}, map[int]bool{}
	for _, v := range a {
		ina[v] = true
	}
	for _, v := range b {
		inb[v] = true
	}
	all := map[int]bool{}
	for v := range ina {
		all[v] = true
	}
	for v := range inb {
		all[v] = true
	}
	for v := range all {
		union = append(union, v)
		switch {
		case ina[v] && inb[v]:
			inter = append(inter, v)
		case ina[v]:
			diff = append(diff, v)
			sym = append(sym, v)
		default:
			sym = append(sym, v)
		}
	}
	for _, s := range [][]int{union, inter, diff, sym} {
		slices.Sort(s)
	}
	return union, inter, diff, sym
}

func eq(a, b []int) bool { return len(a) == len(b) && (len(a) == 0 || reflect.DeepEqual(a, b)) }

func TestSortedOperationsAgainstMaps(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for k := 0; k < 300; k++ {
		a, b := randomSet(r, r.Intn(15), 30), randomSet(r, r.Intn(15), 30)
		u, i, d, s := model(a, b)
		if !eq(sets.Union(a, b), u) || !eq(sets.Intersection(a, b), i) || !eq(sets.Difference(a, b), d) || !eq(sets.SymmetricDifference(a, b), s) {
			t.Fatalf("a=%v b=%v: union %v inter %v diff %v sym %v", a, b, sets.Union(a, b), sets.Intersection(a, b), sets.Difference(a, b), sets.SymmetricDifference(a, b))
		}
		if sets.IsSubset(a, b) != (len(d) == 0) || sets.IsDisjoint(a, b) != (len(i) == 0) {
			t.Fatalf("subset/disjoint wrong for %v %v", a, b)
		}
		var log trace.Log
		got := sets.TraceUnion(a, b, &log)
		if replay := trace.Replay(make([]int, len(got)), log.Events); !eq(replay, got) {
			t.Fatalf("replaying the union's Set events gives %v, want %v", replay, got)
		}
	}
	if j := sets.Jaccard([]int{1, 2, 3}, []int{2, 3, 4}); j != 0.5 {
		t.Errorf("Jaccard = %v", j)
	}
	if sets.Jaccard([]int{}, []int{}) != 1 {
		t.Error("Jaccard of two empty sets should be 1")
	}
	if got := sets.Union([]string{"a", "c"}, []string{"b"}); !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Errorf("string union = %v", got)
	}
}

func TestBitset(t *testing.T) {
	r := rand.New(rand.NewSource(2))
	for k := 0; k < 50; k++ {
		a, b := sets.NewBitset(), sets.NewBitset()
		ma, mb := make([]bool, 300), make([]bool, 300)
		for i := 0; i < 80; i++ {
			x, y := r.Intn(300), r.Intn(300)
			a.Set(x)
			b.Set(y)
			ma[x], mb[y] = true, true
		}
		for i := 0; i < 20; i++ {
			x := r.Intn(300)
			a.Clear(x)
			ma[x] = false
		}
		check := func(name string, got *sets.Bitset, f func(x, y bool) bool) {
			want := 0
			for i := 0; i < 300; i++ {
				if got.Test(i) != f(ma[i], mb[i]) {
					t.Fatalf("%s: bit %d is %v", name, i, got.Test(i))
				}
				if f(ma[i], mb[i]) {
					want++
				}
			}
			if got.Count() != want || len(got.Members()) != want {
				t.Fatalf("%s: count %d, want %d", name, got.Count(), want)
			}
		}
		check("a", a, func(x, _ bool) bool { return x })
		check("union", a.Union(b), func(x, y bool) bool { return x || y })
		check("intersect", a.Intersect(b), func(x, y bool) bool { return x && y })
		check("difference", a.Difference(b), func(x, y bool) bool { return x && !y })
		check("xor", a.Xor(b), func(x, y bool) bool { return x != y })
	}
	b := sets.NewBitset(1, 3, 64)
	if b.Set(3) || !b.Set(5) || b.Test(-1) || b.Set(-2) || b.Len() != 128 {
		t.Error("Set/Test edges wrong")
	}
	if s := b.Snapshot(); len(s.Nodes) != 65 || s.Nodes[64].Role != "set" || s.Nodes[2].Role != "" {
		t.Errorf("snapshot has %d nodes", len(s.Nodes))
	}
}

func TestMultiset(t *testing.T) {
	a := sets.NewMultiset("x", "x", "y")
	b := sets.NewMultiset("x", "z", "z")
	if a.Count("x") != 2 || a.Len() != 3 || a.Distinct() != 2 {
		t.Fatal("counts wrong")
	}
	if u := a.Union(b); u.Count("x") != 2 || u.Count("z") != 2 || u.Len() != 5 {
		t.Errorf("union counts x=%d z=%d len=%d", u.Count("x"), u.Count("z"), u.Len())
	}
	if s := a.Sum(b); s.Count("x") != 3 || s.Len() != 6 {
		t.Errorf("sum x=%d len=%d", s.Count("x"), s.Len())
	}
	if i := a.Intersection(b); i.Count("x") != 1 || i.Len() != 1 {
		t.Error("intersection wrong")
	}
	if !a.Remove("x") || a.Count("x") != 1 || a.Remove("q") {
		t.Error("Remove wrong")
	}
	if got := a.Values(); !reflect.DeepEqual(got, []string{"x", "y"}) {
		t.Errorf("values %v", got)
	}
	if s := a.Snapshot(); s.Nodes[0].Label != "x x1" {
		t.Errorf("snapshot %+v", s.Nodes)
	}
}

func TestOrderedSet(t *testing.T) {
	s := sets.NewOrderedSet(5, 1, 9, 3, 5)
	if s.Len() != 4 || s.Add(9) || !s.Contains(3) || s.Contains(4) {
		t.Fatal("membership wrong")
	}
	if s.Rank(5) != 2 || s.Rank(4) != 2 || s.Rank(100) != 4 {
		t.Errorf("ranks %d %d %d", s.Rank(5), s.Rank(4), s.Rank(100))
	}
	if f, ok := s.Floor(4); !ok || f != 3 {
		t.Errorf("Floor(4) = %d", f)
	}
	if _, ok := s.Floor(0); ok {
		t.Error("Floor(0) found a value")
	}
	if c, ok := s.Ceiling(6); !ok || c != 9 {
		t.Errorf("Ceiling(6) = %d", c)
	}
	if _, ok := s.Ceiling(10); ok {
		t.Error("Ceiling(10) found a value")
	}
	if got := s.Range(2, 9); !reflect.DeepEqual(got, []int{3, 5, 9}) {
		t.Errorf("Range = %v", got)
	}
	if !s.Remove(1) || s.Remove(1) || s.Snapshot().Nodes[0].Role != "min" {
		t.Error("Remove wrong")
	}
}

func TestPowerSetAndProduct(t *testing.T) {
	for n := 0; n <= 10; n++ {
		in := make([]int, n)
		for i := range in {
			in[i] = i
		}
		ps, err := sets.PowerSet(in)
		if err != nil || len(ps) != 1<<n {
			t.Fatalf("PowerSet of %d values: %d subsets, %v", n, len(ps), err)
		}
		seen := map[string]bool{}
		for _, sub := range ps {
			seen[fmtSlice(sub)] = true
		}
		if len(seen) != 1<<n {
			t.Fatalf("PowerSet of %d values has duplicates", n)
		}
	}
	if _, err := sets.PowerSet(make([]int, 21)); err == nil {
		t.Error("PowerSet of 21 values accepted")
	}
	p, err := sets.CartesianProduct([]int{1, 2}, []string{"a", "b", "c"})
	if err != nil || len(p) != 6 || p[0].First != 1 || p[5].Second != "c" {
		t.Errorf("product %v, %v", p, err)
	}
}

func fmtSlice(s []int) string {
	b := make([]byte, 0, len(s)*3)
	for _, v := range s {
		b = append(b, byte('a'+v), ',')
	}
	return string(b)
}

func TestSparseSet(t *testing.T) {
	r := rand.New(rand.NewSource(12))
	s := sets.NewSparseSet(50)
	model := map[int]bool{}
	for step := 0; step < 5000; step++ {
		v := r.Intn(60) - 5
		switch r.Intn(20) {
		case 0:
			s.Clear()
			model = map[int]bool{}
		case 1, 2, 3, 4, 5, 6:
			if s.Remove(v) != model[v] {
				t.Fatalf("Remove(%d) disagreed", v)
			}
			delete(model, v)
		default:
			in := v >= 0 && v < 50
			if s.Add(v) != (in && !model[v]) {
				t.Fatalf("Add(%d) disagreed", v)
			}
			if in {
				model[v] = true
			}
		}
		if s.Len() != len(model) {
			t.Fatalf("Len %d, want %d", s.Len(), len(model))
		}
		for _, x := range s.Values() {
			if !model[x] {
				t.Fatalf("Values has %d", x)
			}
		}
		if p := r.Intn(60) - 5; s.Contains(p) != model[p] {
			t.Fatalf("Contains(%d) disagreed", p)
		}
	}
}

func TestRoaring(t *testing.T) {
	r := rand.New(rand.NewSource(13))
	rb := sets.NewRoaring()
	model := map[uint32]bool{}
	random := func() uint32 {
		if r.Intn(2) == 0 {
			return uint32(r.Intn(6000)) // one dense group: turns into a bitmap
		}
		return uint32(r.Intn(1 << 20)) // sparse across many groups
	}
	for step := 0; step < 30000; step++ {
		v := random()
		if r.Intn(4) == 0 {
			if rb.Remove(v) != model[v] {
				t.Fatalf("Remove(%d) disagreed", v)
			}
			delete(model, v)
		} else {
			if rb.Add(v) == model[v] {
				t.Fatalf("Add(%d) disagreed", v)
			}
			model[v] = true
		}
		if p := random(); rb.Contains(p) != model[p] {
			t.Fatalf("Contains(%d) disagreed", p)
		}
	}
	want := make([]uint32, 0, len(model))
	for v := range model {
		want = append(want, v)
	}
	sort.Slice(want, func(i, j int) bool { return want[i] < want[j] })
	got := rb.Values()
	if len(got) != len(want) || rb.Len() != len(want) {
		t.Fatalf("%d values, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("value %d is %d, want %d", i, got[i], want[i])
		}
	}
	bitmap := false
	for _, c := range rb.Containers() {
		if strings.Contains(c, "bitmap") {
			bitmap = true
		}
	}
	if !bitmap {
		t.Error("the dense group never became a bitmap")
	}
}
