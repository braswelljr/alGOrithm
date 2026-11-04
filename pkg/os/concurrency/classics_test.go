package concurrency_test

import (
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/os/concurrency"
)

func TestReadersWriters(t *testing.T) {
	readerPref, writerPref := 0, 0
	for seed := int64(1); seed <= 30; seed++ {
		r, err := concurrency.ReadersWriters(5, 1, 4, false, seed)
		if err != nil {
			t.Fatal(err)
		}
		w, _ := concurrency.ReadersWriters(5, 1, 4, true, seed)
		if r.Violations != 0 || w.Violations != 0 {
			t.Fatalf("seed %d: a writer shared the data", seed)
		}
		readerPref += r.MaxWriterWait
		writerPref += w.MaxWriterWait
	}
	if readerPref <= writerPref {
		t.Errorf("writers waited %d ticks in all under reader preference, %d under writer preference; reader preference should starve writers more", readerPref, writerPref)
	}
	if r, _ := concurrency.ReadersWriters(4, 1, 3, false, 3); r.MaxReaders < 2 {
		t.Error("readers never read together")
	}
}

func TestSleepingBarber(t *testing.T) {
	for seed := int64(1); seed <= 20; seed++ {
		r, err := concurrency.SleepingBarber(8, 2, seed)
		if err != nil {
			t.Fatal(err)
		}
		if r.Served+r.TurnedAway != 8 || r.Served == 0 {
			t.Fatalf("seed %d: served %d, turned away %d", seed, r.Served, r.TurnedAway)
		}
	}
	slept := false
	for seed := int64(1); seed <= 10; seed++ {
		r, _ := concurrency.SleepingBarber(3, 3, seed)
		if r.TurnedAway != 0 {
			t.Fatalf("seed %d: a customer left with %d free chairs", seed, 3)
		}
		for _, a := range r.Timeline {
			slept = slept || (a.Thread == 0 && a.What == "sleep")
		}
	}
	r, _ := concurrency.SleepingBarber(3, 3, 1)
	if !slept || r.TurnedAway != 0 {
		t.Errorf("with enough chairs nobody leaves, and the barber sleeps: slept %v, turned away %d", slept, r.TurnedAway)
	}
}

func TestPeterson(t *testing.T) {
	broke := false
	for seed := int64(1); seed <= 200; seed++ {
		ok, err := concurrency.Peterson(5, false, seed)
		if err != nil {
			t.Fatal(err)
		}
		if ok.Violations != 0 || ok.Entries != 10 {
			t.Fatalf("seed %d: Peterson's algorithm had %d violations, %d entries", seed, ok.Violations, ok.Entries)
		}
		if b, _ := concurrency.Peterson(5, true, seed); b.Violations > 0 {
			broke = true
		}
	}
	if !broke {
		t.Error("the broken lock never let both threads in")
	}
}

func TestBarrierAndSemaphore(t *testing.T) {
	b, err := concurrency.Barrier(4, 3, 1)
	if err != nil {
		t.Fatal(err)
	}
	passes := map[int]int{}
	for _, a := range b.Timeline {
		if len(a.What) > 4 && a.What[:4] == "pass" {
			passes[a.Thread]++
		}
	}
	if len(passes) != 4 || passes[0] != 3 || b.Waited == 0 {
		t.Errorf("passes %v, waited %d", passes, b.Waited)
	}
	for seed := int64(1); seed <= 20; seed++ {
		s, _ := concurrency.SemaphoreSim(6, 2, 2, seed)
		if s.MaxInside > 2 || s.MaxInside == 0 {
			t.Fatalf("seed %d: %d inside a semaphore of 2", seed, s.MaxInside)
		}
	}
	if _, err := concurrency.SemaphoreSim(99, 1, 1, 1); err == nil {
		t.Error("too many threads accepted")
	}
}
