package virtualmemory

import (
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestTLBEffectiveAccessTime(t *testing.T) {
	r, err := TLB([]int{1, 1, 1, 2, 1, 2, 3, 1}, 2, 10, 100, 1)
	if err != nil {
		t.Fatal(err)
	}
	h := r.HitRate
	want := h*110 + (1-h)*210
	if math.Abs(r.Effective-want) > 1e-9 || math.Abs(r.Measured-want) > 1e-9 {
		t.Errorf("EAT %.2f, measured %.2f, want %.2f", r.Effective, r.Measured, want)
	}
}

func TestPageWalk(t *testing.T) {
	w, err := PageWalk("0x401 7\n0x402 9", 0x401abc, []int{10, 10}, 12)
	if err != nil {
		t.Fatal(err)
	}
	if w.Physical != 7<<12|0xabc || w.Indices[0] != 1 || w.Indices[1] != 1 || w.Tables != 2 {
		t.Errorf("%+v", w)
	}
	if w, _ := PageWalk("0x401 7", 0x800000, []int{10, 10}, 12); w.Physical != -1 {
		t.Error("an unmapped page faults")
	}
}

func TestCopyOnWrite(t *testing.T) {
	r, _ := CopyOnWrite(4, "child read 0\nchild write 1\nchild write 1\nparent write 1\nparent write 2")
	if r.Copies != 2 {
		t.Errorf("copies %d, want 2", r.Copies)
	}
}

func TestThrashing(t *testing.T) {
	pts, _ := Thrashing(100, 20, 80, 10)
	if !(pts[4].Utilization > pts[0].Utilization && pts[9].Utilization < pts[4].Utilization) {
		t.Errorf("utilization should rise then collapse: %v", pts)
	}
}

func TestInvertedPageTable(t *testing.T) {
	t.Run("faults load frames in FIFO order and hits are found", func(t *testing.T) {
		r, err := InvertedPageTable(2, "1:0 1:1 1:0 2:0 1:1 1:0")
		if err != nil {
			t.Fatal(err)
		}
		// 1:0 fault -> f0, 1:1 fault -> f1, 1:0 hit, 2:0 fault evicts f0,
		// 1:1 hit, 1:0 fault evicts f1.
		if r.Faults != 4 {
			t.Errorf("faults %d, want 4", r.Faults)
		}
		if len(r.Tables) != 6 || len(r.Probes) != 6 || len(r.Log) != 6 {
			t.Fatalf("%d tables, %d probes, %d log lines", len(r.Tables), len(r.Probes), len(r.Log))
		}
		want := []InvertedEntry{{2, 0}, {1, 0}}
		if got := r.Tables[5]; got[0] != want[0] || got[1] != want[1] {
			t.Errorf("final table %v, want %v", got, want)
		}
		if first := r.Tables[0]; first[1] != (InvertedEntry{-1, -1}) {
			t.Errorf("frame 1 should be free after one reference: %v", first)
		}
		if !strings.Contains(r.Log[3], "evicting 1:0") {
			t.Errorf("log %q", r.Log[3])
		}
		if !strings.Contains(r.Log[2], "found in frame 0") {
			t.Errorf("log %q", r.Log[2])
		}
	})

	t.Run("a miss probes every frame and a hit at most that many", func(t *testing.T) {
		r, _ := InvertedPageTable(4, "0:0 0:1 0:2 0:3 0:0 0:2")
		for i := 0; i < 4; i++ {
			if r.Probes[i] != 4 {
				t.Errorf("miss %d took %d probes, want 4", i, r.Probes[i])
			}
		}
		for i := 4; i < 6; i++ {
			if r.Probes[i] < 1 || r.Probes[i] > 4 {
				t.Errorf("hit %d took %d probes", i, r.Probes[i])
			}
		}
		if r.Faults != 4 {
			t.Errorf("faults %d, want 4", r.Faults)
		}
	})

	t.Run("snapshots are copies", func(t *testing.T) {
		r, _ := InvertedPageTable(1, "1:1 2:2")
		if r.Tables[0][0] != (InvertedEntry{1, 1}) || r.Tables[1][0] != (InvertedEntry{2, 2}) {
			t.Errorf("tables %v", r.Tables)
		}
	})

	for name, c := range map[string]struct {
		frames int
		refs   string
	}{
		"no frames":      {0, "1:1"},
		"too many":       {33, "1:1"},
		"no references":  {2, "  "},
		"missing colon":  {2, "11"},
		"not a number":   {2, "a:1"},
		"negative page":  {2, "1:-1"},
		"negative pid":   {2, "-1:1"},
		"bad after good": {2, "1:1 1:x"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := InvertedPageTable(c.frames, c.refs); !errors.Is(err, ErrInput) {
				t.Errorf("err %v, want ErrInput", err)
			}
		})
	}
}

func TestMemoryMapped(t *testing.T) {
	t.Run("read copies every access, mmap faults once a page", func(t *testing.T) {
		r, err := MemoryMapped(8, []int{0, 1, 0, 2, 1, 0}, []int{1, 5, 5})
		if err != nil {
			t.Fatal(err)
		}
		want := MmapResult{ReadCopies: 6, ReadCalls: 6, MmapFaults: 4, MmapCopies: 0, DiskReads: 4, DirtyWritten: 2}
		r.Log = nil
		if !reflect.DeepEqual(r, want) {
			t.Errorf("got %+v, want %+v", r, want)
		}
	})

	t.Run("disk reads never exceed the distinct pages touched", func(t *testing.T) {
		acc := []int{3, 3, 3, 3}
		r, _ := MemoryMapped(4, acc, nil)
		if r.DiskReads != 1 || r.MmapFaults != 1 || r.ReadCalls != 4 || r.DirtyWritten != 0 || len(r.Log) != 3 {
			t.Errorf("%+v", r)
		}
	})

	for name, c := range map[string]struct {
		pages      int
		acc, write []int
	}{
		"no pages":           {0, []int{0}, nil},
		"no accesses":        {4, nil, nil},
		"access past end":    {4, []int{4}, nil},
		"write past end":     {4, []int{0}, []int{9}},
		"negative access":    {4, []int{-1}, nil},
		"too many file page": {257, []int{0}, nil},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := MemoryMapped(c.pages, c.acc, c.write); !errors.Is(err, ErrInput) {
				t.Errorf("err %v, want ErrInput", err)
			}
		})
	}
}

func TestWorkingSet(t *testing.T) {
	t.Run("the textbook reference string", func(t *testing.T) {
		refs := []int{1, 2, 1, 3, 4, 4, 4, 3, 4, 4}
		r, err := WorkingSet(refs, 3)
		if err != nil {
			t.Fatal(err)
		}
		wantSets := [][]int{{1}, {1, 2}, {1, 2}, {1, 2, 3}, {1, 3, 4}, {3, 4}, {4}, {3, 4}, {3, 4}, {3, 4}}
		if !reflect.DeepEqual(r.Sets, wantSets) {
			t.Errorf("sets %v, want %v", r.Sets, wantSets)
		}
		for i, s := range r.Sets {
			if r.Sizes[i] != len(s) {
				t.Errorf("size %d is %d, set has %d", i, r.Sizes[i], len(s))
			}
		}
		if r.Peak != 3 {
			t.Errorf("peak %d, want 3", r.Peak)
		}
	})

	t.Run("window 1 is the current page", func(t *testing.T) {
		r, _ := WorkingSet([]int{5, 6, 7}, 1)
		if r.Peak != 1 || !reflect.DeepEqual(r.Sets, [][]int{{5}, {6}, {7}}) {
			t.Errorf("%+v", r)
		}
	})

	t.Run("a window as long as the string grows monotonically", func(t *testing.T) {
		r, _ := WorkingSet([]int{1, 2, 3, 2, 1, 4}, 50)
		for i := 1; i < len(r.Sizes); i++ {
			if r.Sizes[i] < r.Sizes[i-1] {
				t.Errorf("sizes %v shrink", r.Sizes)
			}
		}
		if r.Peak != 4 {
			t.Errorf("peak %d", r.Peak)
		}
	})

	for name, c := range map[string]struct {
		refs   []int
		window int
	}{
		"no references": {nil, 3},
		"window 0":      {[]int{1}, 0},
		"window 51":     {[]int{1}, 51},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := WorkingSet(c.refs, c.window); !errors.Is(err, ErrInput) {
				t.Errorf("err %v, want ErrInput", err)
			}
		})
	}
}

func TestThrashingRejectsBadInput(t *testing.T) {
	if _, err := Thrashing(10, 11, 50, 5); !errors.Is(err, ErrInput) {
		t.Errorf("working set over frames: %v", err)
	}
	pts, _ := Thrashing(100, 10, 50, 1)
	if len(pts) != 1 || math.Abs(pts[0].Utilization-0.5) > 1e-9 || pts[0].Note != "working sets fit" {
		t.Errorf("%+v", pts)
	}
}
