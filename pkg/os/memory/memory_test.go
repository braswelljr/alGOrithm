package memory_test

import (
	"errors"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/os/memory"
)

// The textbook reference string (Silberschatz) with three frames.
var refs = []int{7, 0, 1, 2, 0, 3, 0, 4, 2, 3, 0, 3, 2, 1, 2, 0, 1, 7, 0, 1}

func TestPageReplacement(t *testing.T) {
	for name, c := range map[string]struct {
		run  func([]int, int) (memory.Result, error)
		want int
	}{
		"fifo": {memory.FIFO, 15}, "lru": {memory.LRU, 12}, "optimal": {memory.Optimal, 9}, "clock": {memory.Clock, 14},
	} {
		r, err := c.run(refs, 3)
		if err != nil {
			t.Fatal(err)
		}
		if r.Faults != c.want || r.Faults+r.Hits != len(refs) || len(r.Steps) != len(refs) {
			t.Errorf("%s: %d faults, want %d", name, r.Faults, c.want)
		}
		for i, st := range r.Steps {
			if st.Frames[st.Frame] != refs[i] {
				t.Fatalf("%s step %d: page %d not in frame %d", name, i, refs[i], st.Frame)
			}
		}
	}
}

func TestBeladysAnomaly(t *testing.T) {
	s := []int{1, 2, 3, 4, 1, 2, 5, 1, 2, 3, 4, 5}
	three, _ := memory.FIFO(s, 3)
	four, _ := memory.FIFO(s, 4)
	if three.Faults != 9 || four.Faults != 10 {
		t.Errorf("FIFO faults: %d with 3 frames, %d with 4; Belady's anomaly is 9 and 10", three.Faults, four.Faults)
	}
	if l3, _ := memory.LRU(s, 3); l3.Faults < 0 {
		t.Fatal()
	}
}

func TestAllocation(t *testing.T) {
	parts := []int{100, 500, 200, 300, 600}
	procs := []int{212, 417, 112, 426}
	for name, c := range map[string]struct {
		run  func([]int, []int) (memory.Allocation, error)
		want []int
	}{
		"first": {memory.FirstFit, []int{1, 4, 2, -1}},
		"best":  {memory.BestFit, []int{3, 1, 2, 4}},
		"worst": {memory.WorstFit, []int{4, 1, 3, -1}},
	} {
		a, err := c.run(parts, procs)
		if err != nil {
			t.Fatal(err)
		}
		for i := range c.want {
			if a.Placed[i] != c.want[i] {
				t.Errorf("%s fit placed %v, want %v", name, a.Placed, c.want)
				break
			}
		}
	}
	if _, err := memory.FirstFit(nil, []int{1}); !errors.Is(err, memory.ErrInput) {
		t.Error("no partitions accepted")
	}
	if _, err := memory.LRU(nil, 3); !errors.Is(err, memory.ErrInput) {
		t.Error("no references accepted")
	}
}

func TestLFUAndAging(t *testing.T) {
	// Page 1 is popular; LFU keeps it through a scan of 2, 3, 4.
	refs := []int{1, 1, 1, 2, 3, 4, 1}
	lfu, err := memory.LFU(refs, 2)
	if err != nil {
		t.Fatal(err)
	}
	if lfu.Steps[len(refs)-1].Fault {
		t.Errorf("LFU evicted the popular page: steps %+v", lfu.Steps[len(refs)-1])
	}
	lru, _ := memory.LRU(refs, 2)
	if !lru.Steps[len(refs)-1].Fault {
		t.Error("LRU should have evicted page 1 during the scan")
	}
	textbook := []int{7, 0, 1, 2, 0, 3, 0, 4, 2, 3, 0, 3, 2, 1, 2, 0, 1, 7, 0, 1}
	aging, _ := memory.Aging(textbook, 3)
	lruBook, _ := memory.LRU(textbook, 3)
	if aging.Faults < lruBook.Faults || aging.Faults > lruBook.Faults+3 {
		t.Errorf("aging %d faults, LRU %d: aging should be close to LRU", aging.Faults, lruBook.Faults)
	}
	if _, err := memory.LFU(nil, 2); !errors.Is(err, memory.ErrInput) {
		t.Error("no references accepted")
	}
}

func TestBuddy(t *testing.T) {
	b, err := memory.NewBuddy(1024, 64)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := b.Alloc("A", 100) // 128 at 0
	c, _ := b.Alloc("B", 240) // 256 at 256
	d, _ := b.Alloc("C", 64)  // 64 at 128
	if a != 0 || c != 256 || d != 128 {
		t.Fatalf("offsets %d %d %d, want 0 256 128", a, c, d)
	}
	if b.Fragmentation() != 28+16 {
		t.Errorf("fragmentation %d", b.Fragmentation())
	}
	if _, err := b.Alloc("D", 600); !errors.Is(err, memory.ErrNoSpace) {
		t.Errorf("600 bytes fit: %v", err)
	}
	for _, n := range []string{"A", "B", "C"} {
		if err := b.Free(n); err != nil {
			t.Fatal(err)
		}
	}
	blocks := b.Blocks()
	if len(blocks) != 1 || blocks[0].Size != 1024 {
		t.Errorf("after freeing everything: %+v (buddies should merge back)", blocks)
	}
	if err := b.Free("A"); err == nil {
		t.Error("double free accepted")
	}
	if _, err := memory.NewBuddy(1000, 64); err == nil {
		t.Error("a non-power-of-two arena accepted")
	}
	if s := b.Snapshot(); len(s.Nodes) != 1 || s.Nodes[0].Role != "free" {
		t.Error("snapshot wrong")
	}
}

func TestTranslate(t *testing.T) {
	table, err := memory.ParsePageTable("0 5\n1 2\n3 7")
	if err != nil {
		t.Fatal(err)
	}
	r, err := memory.Translate(1024, table, []int{100, 1100, 2048, 3100, 1200, 100}, 2)
	if err != nil {
		t.Fatal(err)
	}
	if s := r.Steps[0]; s.Page != 0 || s.Offset != 100 || s.Physical != 5*1024+100 {
		t.Errorf("first step %+v", s)
	}
	if !r.Steps[2].Fault || r.Faults != 1 {
		t.Errorf("page 2 should fault: %+v", r.Steps[2])
	}
	// TLB of 2: 0, 1 cached; 3 evicts 0; 1 hits; 0 misses.
	if !r.Steps[4].TLBHit || r.Steps[5].TLBHit || r.TLBHits != 1 {
		t.Errorf("TLB hits %d, steps %+v", r.TLBHits, r.Steps)
	}
	if _, err := memory.Translate(1000, table, []int{1}, 0); err == nil {
		t.Error("a page size of 1000 accepted")
	}
	if _, err := memory.ParsePageTable("x y"); err == nil {
		t.Error("a bad page table accepted")
	}
}
