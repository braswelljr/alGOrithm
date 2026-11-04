package scheduling_test

import (
	"errors"
	"math"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/os/scheduling"
)

// The classic textbook set (Silberschatz): arrivals 0-3, bursts 8 4 9 5.
const classic = "name arrival burst priority\nP1 0 8 3\nP2 1 4 1\nP3 2 9 4\nP4 3 5 2"

func waiting(t *testing.T, s scheduling.Schedule) []int {
	t.Helper()
	out := make([]int, len(s.Stats))
	busy := 0
	for i, st := range s.Stats {
		out[i] = st.Waiting
		busy += st.Burst
		if st.Turnaround != st.Finish-st.Arrival || st.Waiting != st.Turnaround-st.Burst {
			t.Errorf("%s: inconsistent stats %+v", st.Name, st)
		}
	}
	ran := 0
	for _, sl := range s.Timeline {
		if sl.Process != "" {
			ran += sl.End - sl.Start
		}
	}
	if ran != busy {
		t.Errorf("timeline runs %d ticks, bursts total %d", ran, busy)
	}
	return out
}

func eq(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestClassicExample(t *testing.T) {
	ps, err := scheduling.Parse(classic)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		run  func() (scheduling.Schedule, error)
		want []int
		avg  float64
	}{
		{"fcfs", func() (scheduling.Schedule, error) { return scheduling.FCFS(ps) }, []int{0, 7, 10, 18}, 8.75},
		{"sjf", func() (scheduling.Schedule, error) { return scheduling.SJF(ps) }, []int{0, 7, 15, 9}, 7.75},
		{"srtf", func() (scheduling.Schedule, error) { return scheduling.SRTF(ps) }, []int{9, 0, 15, 2}, 6.5},
		{"priority", func() (scheduling.Schedule, error) { return scheduling.Priority(ps, false) }, []int{0, 7, 15, 9}, 7.75},
		{"priority preemptive", func() (scheduling.Schedule, error) { return scheduling.Priority(ps, true) }, []int{9, 0, 15, 2}, 6.5},
	}
	for _, c := range cases {
		s, err := c.run()
		if err != nil {
			t.Fatal(err)
		}
		if got := waiting(t, s); !eq(got, c.want) || s.AvgWaiting != c.avg {
			t.Errorf("%s: waiting %v (avg %v), want %v (avg %v)", c.name, got, s.AvgWaiting, c.want, c.avg)
		}
	}
}

func TestRoundRobin(t *testing.T) {
	ps, _ := scheduling.Parse("P1 0 5\nP2 1 3\nP3 2 1\nP4 3 2\nP5 4 3")
	s, err := scheduling.RoundRobin(ps, 2)
	if err != nil {
		t.Fatal(err)
	}
	waiting(t, s)
	if s.Makespan != 14 {
		t.Errorf("makespan %d", s.Makespan)
	}
	for _, sl := range s.Timeline {
		if sl.End-sl.Start > 2 && sl.Process != "" {
			// consecutive quanta of the same process merge only if nothing else was ready
			continue
		}
	}
	big, _ := scheduling.RoundRobin(ps, 100)
	fcfs, _ := scheduling.FCFS(ps)
	if big.AvgWaiting != fcfs.AvgWaiting {
		t.Error("a quantum longer than every burst should be FCFS")
	}
}

func TestMLFQAndIdle(t *testing.T) {
	ps, _ := scheduling.Parse("long 0 20\nshort 2 2\nlate 30 3")
	s, err := scheduling.MLFQ(ps, []int{2, 4, 8}, 0)
	if err != nil {
		t.Fatal(err)
	}
	waiting(t, s)
	if st := s.Stats[1]; st.Finish > 6 {
		t.Errorf("the short job finished at %d; MLFQ should favor it", st.Finish)
	}
	idle := false
	for _, sl := range s.Timeline {
		idle = idle || sl.Process == ""
	}
	if !idle || s.Stats[2].Start != 30 {
		t.Errorf("expected an idle gap before the late arrival, start %d", s.Stats[2].Start)
	}
	if _, err := scheduling.MLFQ(ps, nil, 10); err != nil {
		t.Error(err)
	}
}

func TestParseErrors(t *testing.T) {
	for _, bad := range []string{"", "P1 0", "P1 x 3\nP2 0 1", "P1 0 0", "P1 0 1\nP1 2 3", "P1 -1 2"} {
		if _, err := scheduling.Parse(bad); !errors.Is(err, scheduling.ErrInput) {
			t.Errorf("Parse(%q) error %v", bad, err)
		}
	}
}

func TestLottery(t *testing.T) {
	ps, _ := scheduling.Parse("rich 0 300 9\npoor 0 300 1")
	s, err := scheduling.Lottery(ps, 1)
	if err != nil {
		t.Fatal(err)
	}
	// Over the first 100 ticks the rich process should run about 90% of the time.
	rich := 0
	for _, sl := range s.Timeline {
		for tick := sl.Start; tick < sl.End && tick < 100; tick++ {
			if sl.Process == "rich" {
				rich++
			}
		}
	}
	if rich < 75 || rich > 99 {
		t.Errorf("rich ran %d of the first 100 ticks, want about 90", rich)
	}
	again, _ := scheduling.Lottery(ps, 1)
	if len(again.Timeline) != len(s.Timeline) {
		t.Error("the same seed gave a different schedule")
	}
}

func busy(rt scheduling.RealTime, name string) int {
	n := 0
	for _, sl := range rt.Timeline {
		if sl.Process == name {
			n += sl.End - sl.Start
		}
	}
	return n
}

func TestEDFAndRateMonotonic(t *testing.T) {
	// The classic set where RM fails and EDF does not: U = 2/5 + 4/7 = 0.971.
	ts, err := scheduling.ParseTasks("name period burst\nT1 5 2\nT2 7 4")
	if err != nil {
		t.Fatal(err)
	}
	edf, _ := scheduling.EDF(ts)
	rm, _ := scheduling.RateMonotonic(ts)
	if edf.Horizon != 35 || len(edf.Missed) != 0 {
		t.Errorf("EDF missed %v over %d ticks", edf.Missed, edf.Horizon)
	}
	if len(rm.Missed) == 0 {
		t.Error("rate-monotonic should miss a deadline of T2")
	}
	if math.Abs(edf.Utilization-(0.4+4.0/7)) > 1e-9 || math.Abs(edf.Bound-0.8284) > 1e-3 {
		t.Errorf("utilization %v bound %v", edf.Utilization, edf.Bound)
	}
	if busy(edf, "T1") != 14 || busy(edf, "T2") != 20 {
		t.Errorf("EDF ran T1 %d and T2 %d ticks, want 14 and 20", busy(edf, "T1"), busy(edf, "T2"))
	}
	// Under the Liu-Layland bound RM meets every deadline.
	easy, _ := scheduling.ParseTasks("A 4 1\nB 5 1\nC 10 2")
	if r, _ := scheduling.RateMonotonic(easy); len(r.Missed) != 0 || r.Utilization > r.Bound {
		t.Errorf("RM under the bound missed %v", r.Missed)
	}
	// Overloaded: even EDF misses.
	over, _ := scheduling.ParseTasks("A 2 1\nB 3 2")
	if r, _ := scheduling.EDF(over); len(r.Missed) == 0 {
		t.Error("an overloaded set met every deadline")
	}
	for _, bad := range []string{"", "A 0 1", "A 2 3", "A 2 1\nA 3 1", "A x 1\nB 2 1"} {
		if _, err := scheduling.ParseTasks(bad); err == nil {
			t.Errorf("ParseTasks(%q) succeeded", bad)
		}
	}
}
