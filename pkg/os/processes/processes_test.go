package processes

import "testing"

func TestLifecycle(t *testing.T) {
	tr, err := Lifecycle("admit dispatch wait wake dispatch timeout dispatch exit")
	if err != nil || tr[len(tr)-1].To != Terminated {
		t.Fatalf("%v %v", tr, err)
	}
	if _, err := Lifecycle("admit dispatch wait dispatch"); err == nil {
		t.Error("a waiting process cannot be dispatched")
	}
}

func TestForkTreeZombiesAndOrphans(t *testing.T) {
	r, err := ForkTree("2 fork\n3 fork\n3 exit 0\n4 exit 1\n2 wait")
	if err != nil {
		t.Fatal(err)
	}
	state := map[int]string{}
	for _, p := range r.Procs {
		state[p.PID] = p.State
	}
	if state[3] != "reaped" || state[4] != "reaped" {
		t.Errorf("states %v: 3 reaped by the shell, 4 adopted and reaped by init", state)
	}
}

func TestSignalsAndThreadModels(t *testing.T) {
	r, _ := Signals("handle SIGINT\nblock SIGINT\nsend SIGINT\nsend SIGINT\nignore SIGKILL\nunblock SIGINT")
	if !r.Alive {
		t.Error("a handled SIGINT keeps the process alive")
	}
	r, _ = Signals("ignore SIGKILL\nsend SIGKILL")
	if r.Alive {
		t.Error("SIGKILL cannot be ignored")
	}
	end := func(tl Timeline) int { return tl[len(tl)-1].Tick }
	one, _ := ThreadModels(ManyToOne, 3, 1, 2, 3, 4)
	oto, _ := ThreadModels(OneToOne, 3, 3, 2, 3, 4)
	if end(oto) >= end(one) {
		t.Errorf("one-to-one finishes at %d, many-to-one at %d", end(oto), end(one))
	}
	m, _ := ThreadsVsProcesses(3, 5, true)
	p, _ := ThreadsVsProcesses(3, 5, false)
	if m.Parent != 15 || p.Parent != 0 {
		t.Errorf("threads %d, processes %d", m.Parent, p.Parent)
	}
}
