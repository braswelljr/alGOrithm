package boot

import "testing"

func TestNestedInterrupts(t *testing.T) {
	r, err := Interrupts("2 1 3\n3 14 4\n4 0 1", 6)
	if err != nil {
		t.Fatal(err)
	}
	// Line 14 preempts line 1's handler at once; line 0 waits for both.
	if r.Latency[0] != 0 || r.Latency[1] != 0 || r.Latency[2] == 0 {
		t.Errorf("latencies %v", r.Latency)
	}
}

func TestDMAFreesTheCPU(t *testing.T) {
	r, _ := DMA(4096, 4, 1024)
	if r.DMACPU >= r.PIOCPU || r.Interrupts != 4 {
		t.Errorf("%+v", r)
	}
}

func TestSequence(t *testing.T) {
	u, _ := Sequence(true, 4)
	b, _ := Sequence(false, 4)
	if u.Total == 0 || b.Total == 0 || u.Stages[len(u.Stages)-1].What != "login prompt" {
		t.Errorf("%+v", u.Stages)
	}
}
