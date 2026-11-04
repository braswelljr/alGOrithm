package virtualmemory

import (
	"math"
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
