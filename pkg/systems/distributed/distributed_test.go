package distributed

import (
	"errors"
	"slices"
	"testing"
)

const clockTrace = `A send B
A local
B recv A
B send C
C local
C recv B`

func TestClocks(t *testing.T) {
	tr, err := ParseTrace(clockTrace)
	if err != nil {
		t.Fatal(err)
	}
	if got := tr.Lamport(); !slices.Equal(got, []int{1, 2, 2, 3, 1, 4}) {
		t.Errorf("Lamport %v", got)
	}
	v := tr.Vector()
	want := [][]int{{1, 0, 0}, {2, 0, 0}, {1, 1, 0}, {1, 2, 0}, {0, 0, 1}, {1, 2, 2}}
	for i := range want {
		if !slices.Equal(v[i], want[i]) {
			t.Errorf("event %d: vector %v, want %v", i, v[i], want[i])
		}
	}
	if !HappensBefore(v[0], v[5]) || HappensBefore(v[5], v[0]) {
		t.Error("A's send happened before C's receive")
	}
	// A's local event and B's receive both have Lamport time 2 but are concurrent;
	// C's local event (Lamport 1) is concurrent with everything before C's receive
	if !Concurrent(v[1], v[2]) || !Concurrent(v[4], v[0]) || Concurrent(v[0], v[2]) {
		t.Error("concurrency is wrong")
	}
	if _, err := ParseTrace("B recv A"); !errors.Is(err, ErrInput) {
		t.Errorf("a receive with no send: %v", err)
	}
}

func TestBully(t *testing.T) {
	res, err := Bully(5, []int{4}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if res.Leader != 3 {
		t.Errorf("leader %d, want 3 (the highest alive)", res.Leader)
	}
	if _, err := Bully(5, []int{1}, 1); err == nil {
		t.Error("a crashed starter was accepted")
	}
	low, _ := Bully(6, nil, 0)
	high, _ := Bully(6, nil, 4)
	if low.Leader != 5 || high.Leader != 5 || low.Messages <= high.Messages {
		t.Errorf("lowest starter: %d messages, a high one: %d", low.Messages, high.Messages)
	}
}

func TestRing(t *testing.T) {
	res, err := Ring([]int{3, 1, 4, 5, 2})
	if err != nil {
		t.Fatal(err)
	}
	if res.Leader != 5 {
		t.Errorf("leader %d", res.Leader)
	}
	up, _ := Ring([]int{1, 2, 3, 4, 5})
	down, _ := Ring([]int{5, 4, 3, 2, 1})
	// increasing clockwise: 2n-1 election messages; decreasing: n(n+1)/2; plus n ELECTED
	if up.Messages != 9+5 || down.Messages != 15+5 {
		t.Errorf("messages: increasing %d, decreasing %d", up.Messages, down.Messages)
	}
	if _, err := Ring([]int{1, 1}); err == nil {
		t.Error("duplicate ids accepted")
	}
}

func TestTwoPhaseCommit(t *testing.T) {
	if r, _ := TwoPhaseCommit(3, 0, false); r.Decision != "commit" || r.Messages != 12 {
		t.Errorf("all yes: %s in %d messages", r.Decision, r.Messages)
	}
	if r, _ := TwoPhaseCommit(3, 2, false); r.Decision != "abort" {
		t.Errorf("one no: %s", r.Decision)
	}
	r, _ := TwoPhaseCommit(3, 0, true)
	if r.Decision != "blocked" || !slices.Equal(r.Blocked, []int{1, 2, 3}) {
		t.Errorf("coordinator crash: %s, blocked %v", r.Decision, r.Blocked)
	}
	r, _ = TwoPhaseCommit(3, 3, true)
	if !slices.Equal(r.Blocked, []int{1, 2}) {
		t.Errorf("crash with a no vote: blocked %v", r.Blocked)
	}
}

func TestRaftElectsOneLeaderPerTerm(t *testing.T) {
	for seed := int64(0); seed < 40; seed++ {
		for _, crash := range []int{0, 60} {
			res, err := Raft(RaftConfig{Nodes: 5, Ticks: 200, Seed: seed, CrashAt: crash, Every: 7})
			if err != nil {
				t.Fatal(err)
			}
			if len(res.Elections) == 0 {
				t.Fatalf("seed %d: no leader", seed)
			}
			leaders := map[int]int{}
			for _, e := range res.Elections {
				if l, ok := leaders[e.Term]; ok && l != e.Leader {
					t.Fatalf("seed %d: term %d has leaders %d and %d", seed, e.Term, l, e.Leader)
				}
				leaders[e.Term] = e.Leader
			}
			if crash > 0 && res.Elections[len(res.Elections)-1].Tick <= crash {
				t.Errorf("seed %d: no election after the crash", seed)
			}
			// log matching: committed entries are on a majority with the same terms
			if res.Committed == 0 {
				t.Errorf("seed %d: nothing committed", seed)
			}
			var leaderLog []int
			for _, l := range res.Logs {
				if len(l) >= res.Committed && (leaderLog == nil || len(l) > len(leaderLog)) {
					leaderLog = l
				}
			}
			same := 0
			for _, l := range res.Logs {
				if len(l) >= res.Committed && slices.Equal(l[:res.Committed], leaderLog[:res.Committed]) {
					same++
				}
			}
			if same < 3 {
				t.Errorf("seed %d: %d committed entries on only %d nodes", seed, res.Committed, same)
			}
		}
	}
	a, _ := Raft(RaftConfig{Nodes: 3, Ticks: 100, Seed: 7})
	b, _ := Raft(RaftConfig{Nodes: 3, Ticks: 100, Seed: 7})
	if a.Messages != b.Messages || len(a.Elections) != len(b.Elections) {
		t.Error("two runs with one seed differ")
	}
}

func TestPaxosChoosesOneValue(t *testing.T) {
	decided := 0
	for seed := int64(0); seed < 200; seed++ {
		res, err := Paxos([]int{10, 20, 30}, 5, seed)
		if err != nil {
			t.Fatal(err)
		}
		if res.Conflict {
			t.Fatalf("seed %d: two values chosen", seed)
		}
		if res.Decided {
			decided++
			if res.Chosen != 10 && res.Chosen != 20 && res.Chosen != 30 {
				t.Fatalf("seed %d: chose %d, which nobody proposed", seed, res.Chosen)
			}
		}
	}
	if decided < 150 {
		t.Errorf("only %d of 200 runs decided", decided)
	}
	one, _ := Paxos([]int{42}, 3, 1)
	if !one.Decided || one.Chosen != 42 {
		t.Errorf("one proposer: %+v", one)
	}
}

func TestGossip(t *testing.T) {
	push, err := Gossip(1000, 1, false, 1)
	if err != nil {
		t.Fatal(err)
	}
	pull, _ := Gossip(1000, 1, true, 1)
	if push[len(push)-1] != 1000 || pull[len(pull)-1] != 1000 {
		t.Fatal("the rumor did not reach everyone")
	}
	if len(pull) > len(push) || len(push) > 40 {
		t.Errorf("rounds: push %d, push-pull %d", len(push)-1, len(pull)-1)
	}
	if !slices.IsSorted(push) {
		t.Error("informed counts went down")
	}
}

func TestRendezvousMovesFewKeys(t *testing.T) {
	res, err := Join(4, 10_000)
	if err != nil {
		t.Fatal(err)
	}
	// a fifth node should take about a fifth; mod-N moves about four fifths
	if res.Rendezvous < 1500 || res.Rendezvous > 2500 || res.Modulo < 7000 {
		t.Errorf("moved: rendezvous %d, modulo %d", res.Rendezvous, res.Modulo)
	}
	if Owner("k", []string{"a", "b"}) != Owner("k", []string{"b", "a"}) {
		t.Error("the owner depends on the node order")
	}
}

func TestQuorum(t *testing.T) {
	strong, _ := Quorum(5, 3, 3, 2000, 1)
	weak, _ := Quorum(5, 1, 1, 2000, 1)
	if !strong.Overlap || strong.Stale != 0 {
		t.Errorf("R+W>N: %d stale reads", strong.Stale)
	}
	if weak.Overlap || weak.Stale == 0 {
		t.Errorf("R=W=1 of 5: %d stale reads", weak.Stale)
	}
}

func TestCRDTs(t *testing.T) {
	run, err := RunGCounter("A inc 3\nB inc 2\nC inc 5\nA merge B\nB merge C\nC merge A\nA merge C\nB merge A")
	if err != nil {
		t.Fatal(err)
	}
	for i, c := range run.Counters {
		if c.Value() != 10 {
			t.Errorf("%s ends at %d", run.Replicas[i], c.Value())
		}
	}
	// concurrent add and remove: the add B never saw survives
	set, err := RunORSet("A add x\nB merge A\nB remove x\nA add x\nA merge B\nB merge A")
	if err != nil {
		t.Fatal(err)
	}
	for i, s := range set.Sets {
		if !slices.Equal(s, []string{"x"}) {
			t.Errorf("%s holds %v, want [x] (add wins)", set.Replicas[i], s)
		}
	}
	gone, _ := RunORSet("A add y\nB merge A\nB remove y\nA merge B")
	if len(gone.Sets[0]) != 0 {
		t.Errorf("an observed remove did not remove: %v", gone.Sets[0])
	}
	if _, err := RunGCounter("A dec 1"); err == nil {
		t.Error("dec was accepted")
	}
}

func TestChandyLamportIsConsistent(t *testing.T) {
	for seed := int64(0); seed < 50; seed++ {
		res, err := ChandyLamport(4, 60, 20, seed)
		if err != nil {
			t.Fatal(err)
		}
		if res.Total != res.Expected {
			t.Fatalf("seed %d: snapshot total %d, want %d", seed, res.Total, res.Expected)
		}
	}
}

func TestByzantine(t *testing.T) {
	// four generals, one traitorous lieutenant: the loyal ones obey
	r, err := Byzantine(4, 1, Attack, []int{3})
	if err != nil {
		t.Fatal(err)
	}
	if !r.Agree || !r.Obey || r.Decisions[0] != Attack || r.Decisions[1] != Attack {
		t.Errorf("traitor lieutenant: %+v", r)
	}
	// a traitorous commander: the loyal lieutenants still agree
	if r, _ := Byzantine(4, 1, Attack, []int{0}); !r.Agree {
		t.Errorf("traitor commander: %+v", r)
	}
	// three generals, one traitor: the loyal lieutenant is fooled
	if r, _ := Byzantine(3, 1, Attack, []int{2}); r.Obey {
		t.Errorf("n=3 with a traitor should fail: %+v", r)
	}
	// OM(2) on seven generals with two traitors
	if r, _ := Byzantine(7, 2, Attack, []int{2, 5}); !r.Agree || !r.Obey {
		t.Errorf("OM(2), seven generals: %+v", r)
	}
}
