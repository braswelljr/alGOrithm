package synchronization

import "testing"

func TestCheckFindsWhatEachVariantGetsWrong(t *testing.T) {
	cases := []struct {
		v           Variant
		mutex, live bool
	}{
		{Peterson, true, true},
		{CheckThenSet, false, true},
		{FlagsFirst, true, false},
	}
	for _, c := range cases {
		r, err := Check(c.v)
		if err != nil {
			t.Fatal(err)
		}
		if r.MutualExcl != c.mutex || r.DeadlockFree != c.live {
			t.Errorf("%s: %v, want mutual exclusion %v, deadlock-free %v", c.v, r, c.mutex, c.live)
		}
		if (c.mutex && c.live) != (len(r.Counterexample) == 0) {
			t.Errorf("%s: counterexample %v", c.v, r.Counterexample)
		}
	}
	if _, err := Check("nope"); err == nil {
		t.Error("an unknown variant is an error")
	}
}

func TestMutualExclusionHoldsOverManySeeds(t *testing.T) {
	for seed := int64(1); seed <= 40; seed++ {
		if r, _ := Bakery(4, 2, 2, seed); r.Violated || r.Deadlocked {
			t.Fatalf("bakery seed %d: %+v", seed, r.MaxInside)
		}
		for _, k := range []Lock{TestAndSet, CompareAndSwap} {
			if r, _ := Spinlock(k, 3, 3, 2, seed); r.Violated || r.MaxInside != 1 {
				t.Fatalf("%s seed %d: max inside %d", k, seed, r.MaxInside)
			}
		}
		if r, _ := ConditionVariable(true, 3, 6, seed); r.Counts["underflows"] != 0 {
			t.Fatalf("while loop seed %d underflowed", seed)
		}
	}
	broken, underflow := false, false
	for seed := int64(1); seed <= 200; seed++ {
		if r, _ := Spinlock(NonAtomic, 3, 3, 2, seed); r.Violated {
			broken = true
		}
		if r, _ := ConditionVariable(false, 3, 6, seed); r.Counts["underflows"] > 0 {
			underflow = true
		}
	}
	if !broken || !underflow {
		t.Errorf("some seed should show the non-atomic lock failing (%v) and the if-wait underflowing (%v)", broken, underflow)
	}
}

func TestPriorityInheritanceRemovesTheInversion(t *testing.T) {
	without, _ := PriorityInversion(8, false)
	with, _ := PriorityInversion(8, true)
	if with.Counts["high finish"] >= without.Counts["high finish"] {
		t.Errorf("high finishes at %d with inheritance and %d without", with.Counts["high finish"], without.Counts["high finish"])
	}
	if without.Counts["high finish"] < without.Counts["medium finish"] {
		t.Error("without inheritance, high should wait for medium")
	}
}

func TestABA(t *testing.T) {
	if !ABA(false).Corrupted {
		t.Error("an untagged CAS should corrupt the stack")
	}
	if r := ABA(true); r.Corrupted || r.Final != "C" {
		t.Errorf("tagged: %+v", r)
	}
}

func TestSmokers(t *testing.T) {
	if r, _ := CigaretteSmokers(10, false, 1); r.Deadlocked {
		t.Error("the fixed version never deadlocks")
	}
	dead := false
	for seed := int64(1); seed < 20; seed++ {
		if r, _ := CigaretteSmokers(10, true, seed); r.Deadlocked {
			dead = true
		}
	}
	if !dead {
		t.Error("some seed should deadlock the naive smokers")
	}
}
