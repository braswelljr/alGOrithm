package cache_test

import (
	"errors"
	"math"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/os/cache"
)

func TestSplit(t *testing.T) {
	c := cache.Config{Lines: 8, BlockSize: 16, Ways: 2} // 4 sets: 4 offset bits, 2 set bits
	if got := c.Split(0b1011_10_0110); got != (cache.Split{Tag: 0b1011, Set: 0b10, Offset: 0b0110}) {
		t.Errorf("Split = %+v", got)
	}
}

func TestDirectMappedConflicts(t *testing.T) {
	// 4 one-word lines; addresses 0 and 16 map to set 0 and keep evicting each other.
	c := cache.Config{Lines: 4, BlockSize: 4, Ways: 1}
	r, err := cache.Simulate(c, []int{0, 16, 0, 16, 4})
	if err != nil {
		t.Fatal(err)
	}
	kinds := []string{}
	for _, a := range r.Accesses {
		kinds = append(kinds, a.Result)
	}
	want := []string{cache.Compulsory, cache.Compulsory, cache.Conflict, cache.Conflict, cache.Compulsory}
	for i := range want {
		if kinds[i] != want[i] {
			t.Fatalf("results %v, want %v", kinds, want)
		}
	}
	if r.Hits != 0 || r.Misses != 5 {
		t.Errorf("hits %d misses %d", r.Hits, r.Misses)
	}
	// Two ways remove the conflict: the third and fourth accesses hit.
	r2, _ := cache.Simulate(cache.Config{Lines: 4, BlockSize: 4, Ways: 2}, []int{0, 16, 0, 16})
	if r2.Hits != 2 {
		t.Errorf("2-way hits %d, want 2", r2.Hits)
	}
}

func TestCapacityAndPolicies(t *testing.T) {
	// Fully associative, 2 lines, looping over 3 blocks: LRU misses every time
	// (capacity misses after the first three), FIFO too; a reused block helps LRU.
	c := cache.Config{Lines: 2, BlockSize: 1, Ways: 2, Policy: cache.LRU}
	r, _ := cache.Simulate(c, []int{0, 1, 2, 0, 1, 2})
	if r.Hits != 0 || r.ByKind[cache.Capacity] != 3 || r.ByKind[cache.Compulsory] != 3 {
		t.Errorf("loop: hits %d kinds %v", r.Hits, r.ByKind)
	}
	trace := []int{0, 1, 0, 2, 0, 1}
	lru, _ := cache.Simulate(cache.Config{Lines: 2, BlockSize: 1, Ways: 2, Policy: cache.LRU}, trace)
	fifo, _ := cache.Simulate(cache.Config{Lines: 2, BlockSize: 1, Ways: 2, Policy: cache.FIFO}, trace)
	if lru.Hits != 2 || fifo.Hits != 1 {
		t.Errorf("LRU hits %d (want 2), FIFO hits %d (want 1)", lru.Hits, fifo.Hits)
	}
	if lru.Final[0][0] < 0 || len(lru.Final) != 1 {
		t.Errorf("final %v", lru.Final)
	}
}

func TestErrorsAndAMAT(t *testing.T) {
	for _, c := range []cache.Config{{Lines: 3, BlockSize: 4, Ways: 1}, {Lines: 4, BlockSize: 3, Ways: 1}, {Lines: 4, BlockSize: 4, Ways: 3}, {Lines: 4, BlockSize: 4, Ways: 1, Policy: "random"}} {
		if _, err := cache.Simulate(c, []int{1}); !errors.Is(err, cache.ErrConfig) {
			t.Errorf("%+v accepted", c)
		}
	}
	if _, err := cache.Simulate(cache.Config{Lines: 4, BlockSize: 4, Ways: 1}, []int{-1}); !errors.Is(err, cache.ErrConfig) {
		t.Error("negative address accepted")
	}
	if got := cache.AMAT(1, 0.05, 100); math.Abs(got-6) > 1e-9 {
		t.Errorf("AMAT = %v, want 6", got)
	}
}
