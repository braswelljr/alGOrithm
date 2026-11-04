package probabilistic_test

import (
	"math"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/structures/probabilistic"
)

func TestBloomFilter(t *testing.T) {
	f := probabilistic.NewBloomFilter(2000, 4)
	for i := 0; i < 200; i++ {
		f.Add("item" + strconv.Itoa(i))
	}
	for i := 0; i < 200; i++ {
		if !f.MightContain("item" + strconv.Itoa(i)) {
			t.Fatalf("false negative for item%d", i)
		}
	}
	fp := 0
	const probes = 20000
	for i := 0; i < probes; i++ {
		if f.MightContain("other" + strconv.Itoa(i)) {
			fp++
		}
	}
	rate, want := float64(fp)/probes, f.FalsePositiveRate()
	if rate > 3*want+0.01 || rate < want/3-0.01 {
		t.Errorf("measured false-positive rate %.4f, estimate %.4f", rate, want)
	}
	if len(f.Snapshot().Nodes) != 2000 || f.Len() != 200 || f.Bits() != 2000 || f.Hashes() != 4 || len(f.Positions("x")) != 4 {
		t.Error("accessors wrong")
	}
}

func TestCountMinSketch(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	s := probabilistic.NewCountMinSketch(64, 4)
	truth := map[string]int{}
	for i := 0; i < 5000; i++ {
		k := "k" + strconv.Itoa(int(r.ExpFloat64()*20)) // a skewed stream
		n := 1 + r.Intn(3)
		s.Add(k, n)
		truth[k] += n
	}
	over := 0
	for k, n := range truth {
		est := s.Estimate(k)
		if est < n {
			t.Fatalf("Estimate(%s) = %d below the true %d", k, est, n)
		}
		if float64(est-n) > s.ErrorBound() {
			over++
		}
	}
	if over > len(truth)/10 { // the bound holds with probability 1 - e^-4 per key
		t.Errorf("%d of %d keys beyond the error bound", over, len(truth))
	}
	if s.Estimate("never") > int(s.ErrorBound())+1 {
		t.Errorf("an unseen key is estimated at %d", s.Estimate("never"))
	}
}

func TestHyperLogLog(t *testing.T) {
	for _, n := range []int{0, 10, 100, 1000, 20000, 200000} {
		h := probabilistic.NewHyperLogLog(12)
		for i := 0; i < n; i++ {
			h.Add("item-" + strconv.Itoa(i))
			h.Add("item-" + strconv.Itoa(i)) // duplicates change nothing
		}
		got := h.Count()
		if n == 0 {
			if got != 0 {
				t.Errorf("empty count %d", got)
			}
			continue
		}
		if rel := math.Abs(float64(got-n)) / float64(n); rel > 4*h.StandardError() {
			t.Errorf("n=%d: estimate %d is %.1f%% off (standard error %.1f%%)", n, got, 100*rel, 100*h.StandardError())
		}
	}
}

func TestCuckooFilter(t *testing.T) {
	f := probabilistic.NewCuckooFilter(256, 1)
	var added []string
	for i := 0; i < 900; i++ {
		k := "a" + strconv.Itoa(i)
		if f.Add(k) {
			added = append(added, k)
		}
	}
	if len(added) < 850 {
		t.Fatalf("only %d of 900 fit in 1024 slots", len(added))
	}
	for _, k := range added {
		if !f.Contains(k) {
			t.Fatalf("false negative for %s", k)
		}
	}
	fp := 0
	for i := 0; i < 10000; i++ {
		if f.Contains("b" + strconv.Itoa(i)) {
			fp++
		}
	}
	if rate := float64(fp) / 10000; rate > 0.06 {
		t.Errorf("false-positive rate %.3f", rate)
	}
	for _, k := range added[:400] {
		if !f.Remove(k) {
			t.Fatalf("Remove(%s) found nothing", k)
		}
	}
	for _, k := range added[400:] {
		if !f.Contains(k) {
			t.Fatalf("removing others lost %s", k)
		}
	}
	if f.Len() != len(added)-400 {
		t.Errorf("Len %d, want %d", f.Len(), len(added)-400)
	}
}

func TestCountMinSketchTotalsAndSnapshot(t *testing.T) {
	s := probabilistic.NewCountMinSketch(0, 0) // held to 1 x 1
	s.Add("a", 3)
	s.Add("b", 0)
	s.Add("c", -2)
	if s.Total() != 3 || s.Estimate("anything") != 3 {
		t.Errorf("total %d, estimate %d: one counter holds everything", s.Total(), s.Estimate("anything"))
	}

	s = probabilistic.NewCountMinSketch(5, 3)
	for i, w := range []string{"x", "y", "z", "x", "w"} {
		s.Add(w, i+1)
	}
	if s.Total() != 15 {
		t.Errorf("total %d, want 15", s.Total())
	}
	snap := s.Snapshot()
	if snap.Layout != "row" || len(snap.Nodes) != 3 {
		t.Fatalf("%+v", snap)
	}
	// every row's counters add up to the total: each Add touches one counter a row
	for i, n := range snap.Nodes {
		prefix := "hash " + strconv.Itoa(i+1) + ": ["
		if !strings.HasPrefix(n.Label, prefix) {
			t.Fatalf("label %q", n.Label)
		}
		sum := 0
		for _, f := range strings.Fields(strings.Trim(strings.TrimPrefix(n.Label, prefix), "]")) {
			v, err := strconv.Atoi(f)
			if err != nil {
				t.Fatal(err)
			}
			sum += v
		}
		if sum != 15 {
			t.Errorf("row %d sums to %d, want the total 15", i+1, sum)
		}
	}
	if s.Estimate("x") < 5 {
		t.Errorf("x was added 1 + 4 times, estimate %d", s.Estimate("x"))
	}
}

func TestCuckooFilterFull(t *testing.T) {
	f := probabilistic.NewCuckooFilter(2, 3) // 2 buckets of 4
	var added []string
	for i := 0; i < 100; i++ {
		item := "item" + strconv.Itoa(i)
		if !f.Add(item) {
			break
		}
		added = append(added, item)
	}
	if len(added) == 0 || len(added) > 8 || f.Len() != len(added) {
		t.Fatalf("added %d, len %d", len(added), f.Len())
	}
	if lf := f.LoadFactor(); lf != float64(len(added))/8 {
		t.Errorf("load factor %v, want %d/8", lf, len(added))
	}
	// a failed Add undoes its kicks, so nothing stored is lost
	for _, item := range added {
		if !f.Contains(item) {
			t.Errorf("%s was lost by a failed Add", item)
		}
	}
	snap := f.Snapshot()
	used := 0
	for i, n := range snap.Nodes {
		if !strings.HasPrefix(n.Label, strconv.Itoa(i)+":") {
			t.Errorf("label %q", n.Label)
		}
		used += 4 - strings.Count(n.Label, "..")
	}
	if len(snap.Nodes) != 2 || used != f.Len() {
		t.Errorf("%d buckets, %d used slots, len %d", len(snap.Nodes), used, f.Len())
	}
	for _, item := range added {
		if !f.Remove(item) {
			t.Errorf("Remove(%s) found nothing", item)
		}
	}
	if f.Len() != 0 || f.LoadFactor() != 0 || f.Remove("item0") {
		t.Errorf("after removing everything: len %d", f.Len())
	}
	one := probabilistic.NewCuckooFilter(1, 1)
	for i := 0; i < 4; i++ {
		if !one.Add("k" + strconv.Itoa(i)) {
			t.Fatalf("a 4-slot filter refused item %d", i)
		}
	}
	if one.Add("k4") || one.Len() != 4 {
		t.Errorf("a full one-bucket filter took a fifth item: len %d", one.Len())
	}
}

func TestHyperLogLogLargeCountsAndRegisters(t *testing.T) {
	for _, p := range []int{0, 5, 6, 10} { // 0 is held to 4
		h := probabilistic.NewHyperLogLog(p)
		const n = 20_000
		for i := 0; i < n; i++ {
			h.Add("user-" + strconv.Itoa(i))
		}
		est := h.Count()
		if rel := math.Abs(float64(est)-n) / n; rel > 4*h.StandardError() {
			t.Errorf("p %d: estimate %d for %d, off by %.1f%% (standard error %.1f%%)", p, est, n, 100*rel, 100*h.StandardError())
		}
		regs := h.Registers()
		bits := max(p, 4)
		if len(regs) != 1<<bits {
			t.Errorf("p %d: %d registers", p, len(regs))
		}
		for _, r := range regs {
			if r < 0 || r > 64-bits+1 {
				t.Errorf("register %d out of range", r)
			}
		}
		regs[0] = 99
		if h.Registers()[0] == 99 {
			t.Error("Registers returned the internal slice")
		}
		snap := h.Snapshot()
		if want := min(1<<bits, 64); len(snap.Nodes) != want || !strings.HasPrefix(snap.Nodes[0].Label, "r0=") {
			t.Errorf("p %d: %d snapshot nodes, want %d", p, len(snap.Nodes), want)
		}
	}
	if probabilistic.NewHyperLogLog(4).Count() != 0 {
		t.Error("an empty estimator counts 0")
	}
}
