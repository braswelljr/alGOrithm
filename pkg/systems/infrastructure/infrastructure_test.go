package infrastructure

import (
	"strings"
	"testing"
)

func TestCircuitBreaker(t *testing.T) {
	r, err := CircuitBreaker("fail fail fail ok ok ok ok ok fail ok ok", 3, 3)
	if err != nil {
		t.Fatal(err)
	}
	want := []BreakerState{Closed, Closed, Closed, Open, Open, HalfOpen, Closed}
	for i, s := range want {
		if r.States[i] != s {
			t.Fatalf("state %d is %s, want %s: %v", i, r.States[i], s, r.States)
		}
	}
	if r.Rejected != 2 {
		t.Errorf("rejected %d, want 2", r.Rejected)
	}
}

func TestBalancingAdaptsToSpeed(t *testing.T) {
	rr, _ := Balance(RoundRobin, []int{1, 4}, 100, 1)
	lc, _ := Balance(LeastConnections, []int{1, 4}, 100, 1)
	if lc.Latency >= rr.Latency {
		t.Errorf("least connections %.1f should beat round robin %.1f on uneven servers", lc.Latency, rr.Latency)
	}
}

func TestJitterSpreadsRetries(t *testing.T) {
	none, _ := Backoff(100, 4, 2, 64, NoJitter, 1)
	full, _ := Backoff(100, 4, 2, 64, FullJitter, 1)
	if none.Peak != 100 || full.Peak >= none.Peak {
		t.Errorf("peaks %d and %d", none.Peak, full.Peak)
	}
}

func TestDeliveryGuarantees(t *testing.T) {
	most, _ := Delivery(AtMostOnce, 200, 20, 1)
	least, _ := Delivery(AtLeastOnce, 200, 20, 1)
	exact, _ := Delivery(ExactlyOnce, 200, 20, 1)
	if most.Duplicates != 0 || most.Lost == 0 || least.Lost != 0 || least.Duplicates == 0 || exact.Lost != 0 || exact.Duplicates != 0 {
		t.Errorf("%+v\n%+v\n%+v", most, least, exact)
	}
}

func TestStampedeFixes(t *testing.T) {
	none, _ := Stampede(20, 60, 20, 3, NoFix, 1)
	lock, _ := Stampede(20, 60, 20, 3, LockFix, 1)
	if none.Peak != 20 || lock.Peak != 1 {
		t.Errorf("peaks %d and %d", none.Peak, lock.Peak)
	}
}

func TestCgroups(t *testing.T) {
	cs, err := Cgroups(2, "web 1024 2 512 300\nbatch 512 2 256 100\nleak 1024 1 128 400")
	if err != nil {
		t.Fatal(err)
	}
	if !cs[2].Killed || cs[2].Got != 0 {
		t.Errorf("leak should be OOM-killed: %v", cs[2])
	}
	if d := cs[0].Got + cs[1].Got; d < 1.999 || cs[0].Got < 1.33 || cs[0].Got > 1.34 {
		t.Errorf("web got %.3f, batch %.3f", cs[0].Got, cs[1].Got)
	}
	if !strings.Contains(cs[2].String(), "OOM") {
		t.Error(cs[2].String())
	}
}

func TestCanaryLimitsTheDamage(t *testing.T) {
	bg, _ := Deploy(BlueGreen, 10, 1000, 30, true)
	cn, _ := Deploy(Canary, 10, 1000, 30, true)
	if !bg.RolledBack || !cn.RolledBack || cn.Errors >= bg.Errors {
		t.Errorf("blue/green %d errors, canary %d", bg.Errors, cn.Errors)
	}
}
