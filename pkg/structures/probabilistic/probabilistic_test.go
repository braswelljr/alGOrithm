package probabilistic_test

import (
	"math"
	"math/rand"
	"strconv"
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
