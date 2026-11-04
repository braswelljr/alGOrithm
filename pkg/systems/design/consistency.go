package design

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

// CAPConfig is two replicas, a network partition between them, and writes.
type CAPConfig struct {
	Ops         int   // operations, alternating between the two sides
	PartitionAt int   // operation at which the network splits
	HealAt      int   // operation at which it heals
	SyncMs      int   // the extra milliseconds a synchronous (consistent) write waits for the other replica
	Seed        int64 // which side writes
}

// CAP plays the same operations against a consistent (CP) and an available
// (AP) pair of replicas through a network partition. During the partition
// CP refuses writes on the side that cannot reach the other (unavailable but
// never stale), and AP accepts them on both sides (available, but the sides
// diverge and must reconcile on healing, with conflicts). PACELC adds the
// rest of the time: Else, without a partition, the choice is Latency vs
// Consistency: CP waits SyncMs for the other replica on every write, AP acks
// at once. It leaves out quorums of more than two.
//
// Time: O(Ops). Space: O(1).
func CAP(c CAPConfig) (Report, error) {
	if err := check(c.Ops >= 1 && c.Ops <= 100_000 && c.PartitionAt >= 0 && c.HealAt >= c.PartitionAt && c.SyncMs >= 0, "1 to 100,000 ops, the partition healing after it starts"); err != nil {
		return Report{}, err
	}
	rng := rand.New(rand.NewSource(c.Seed))
	var cpRejected, apDiverged, apConflicts, cpWait int
	wroteA, wroteB := false, false
	for op := 0; op < c.Ops; op++ {
		split := op >= c.PartitionAt && op < c.HealAt
		sideB := rng.Intn(2) == 1
		if op == c.HealAt && wroteA && wroteB {
			apConflicts++
		}
		if !split {
			cpWait += c.SyncMs
			wroteA, wroteB = false, false
			continue
		}
		if sideB { // B is the minority side that cannot reach the leader
			cpRejected++
			wroteB = true
		} else {
			wroteA = true
		}
		apDiverged++
	}
	a := arch([]Component{
		{ID: "c1", Kind: Client, Label: "Clients (side A)"},
		{ID: "a", Kind: Database, Label: "Replica A"},
		{ID: "b", Kind: Database, Label: "Replica B"},
		{ID: "c2", Kind: Client, Label: "Clients (side B)"},
	}, "c1>a", "c2>b", "a>b:replicate")
	f := a.frame("partition: A and B cannot reach each other")
	a.send(f, "c1", "a", 1, OK)
	a.send(f, "c2", "b", 2, Dropped)
	a.send(f, "a", "b", 3, Failed)
	healed := a.frame("healed: replicas exchange their writes")
	a.send(healed, "a", "b", 4, OK)
	var r Report
	r.Arch = a
	r.say("partition from op %d to %d:", c.PartitionAt, c.HealAt)
	r.say("  CP: %d writes refused on the cut-off side; every read stays current", cpRejected)
	r.say("  AP: %d writes accepted while the sides disagreed; %d conflicts to reconcile on healing", apDiverged, apConflicts)
	r.say("no partition (PACELC's Else): CP waited %d ms in all for synchronous acks; AP waited none but reads may be stale", cpWait)
	r.put("CP refused writes", float64(cpRejected))
	r.put("AP divergent writes", float64(apDiverged))
	r.put("AP conflicts", float64(apConflicts))
	r.put("CP extra latency (ms)", float64(cpWait))
	return r, nil
}

// Op is one operation of a history: a client writes or reads a value of a
// register, between a start and an end time.
type Op struct {
	Client     string
	Write      bool
	Key        string
	Value      int
	Start, End int
}

// ParseHistory reads one operation a line: "client w|r key value start end",
// for example "A w x 1 0 2" (A wrote 1 to x between times 0 and 2) or
// "B r x 0 3 4" (B read 0 from x). Every key starts at 0.
func ParseHistory(text string) ([]Op, error) {
	var ops []Op
	for i, line := range strings.Split(text, "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		if err := check(len(f) == 6 && (f[1] == "w" || f[1] == "r"), "line %d: want \"client w|r key value start end\"", i+1); err != nil {
			return nil, err
		}
		v, e1 := strconv.Atoi(f[3])
		s, e2 := strconv.Atoi(f[4])
		e, e3 := strconv.Atoi(f[5])
		if err := check(e1 == nil && e2 == nil && e3 == nil && s <= e, "line %d: value, start and end are numbers with start <= end", i+1); err != nil {
			return nil, err
		}
		ops = append(ops, Op{f[0], f[1] == "w", f[2], v, s, e})
	}
	return ops, check(len(ops) > 0 && len(ops) <= 10, "1 to 10 operations")
}

// Models says which consistency models a history satisfies.
type Models struct {
	Linearizable bool // one order of all ops, respecting real time, where every read sees the last write
	Sequential   bool // one order of all ops, respecting each client's own order
	Causal       bool // each client sees all writes in some order respecting every client's order
	Eventual     bool // the weakest: always true of a finished history, if replicas converge
}

// CheckConsistency tests a small history of register reads and writes
// against the consistency models, strongest first, by searching for an
// order of the operations that each model allows: linearizable (respects
// real time: an op that ended before another began comes first), sequential
// (respects only each client's program order), causal (checked here as each
// client seeing every write in an order that keeps all clients' program
// order, which is PRAM consistency; full causal consistency also orders a
// write after the writes its client had read) and eventual (any read may be
// stale for a while). It explains the first model that
// fails. The search tries orders exhaustively, so histories are kept to 10
// operations.
//
// Time: O(n!) in the worst case for n operations. Space: O(n).
func CheckConsistency(ops []Op) (Models, Report, error) {
	if err := check(len(ops) > 0 && len(ops) <= 10, "1 to 10 operations"); err != nil {
		return Models{}, Report{}, err
	}
	before := func(rule string) func(i, j int) bool {
		return func(i, j int) bool {
			switch rule {
			case "real time":
				return ops[i].End < ops[j].Start
			default:
				return ops[i].Client == ops[j].Client && ops[i].Start < ops[j].Start
			}
		}
	}
	all := make([]int, len(ops))
	for i := range all {
		all[i] = i
	}
	m := Models{Eventual: true}
	m.Linearizable = orderExists(ops, all, before("real time"))
	m.Sequential = orderExists(ops, all, before("program"))
	m.Causal = true
	clients := map[string]bool{}
	for _, o := range ops {
		clients[o.Client] = true
	}
	for cl := range clients {
		var sub []int
		for i, o := range ops {
			if o.Write || o.Client == cl {
				sub = append(sub, i)
			}
		}
		if !orderExists(ops, sub, before("program")) {
			m.Causal = false
		}
	}
	var r Report
	for _, o := range ops {
		verb := "read"
		if o.Write {
			verb = "wrote"
		}
		r.say("%s %s %s=%d during [%d, %d]", o.Client, verb, o.Key, o.Value, o.Start, o.End)
	}
	yes := map[bool]string{true: "yes", false: "no"}
	r.say("linearizable: %s; sequential: %s; causal: %s; eventual: %s", yes[m.Linearizable], yes[m.Sequential], yes[m.Causal], yes[m.Eventual])
	switch {
	case !m.Causal:
		r.say("not even causal: some client saw writes in an order that breaks another client's program order")
	case !m.Sequential:
		r.say("causal but not sequential: clients disagree on the order of concurrent writes")
	case !m.Linearizable:
		r.say("sequential but not linearizable: a read missed a write that had already finished in real time")
	default:
		r.say("linearizable: it behaves like a single copy")
	}
	for name, ok := range map[string]bool{"linearizable": m.Linearizable, "sequential": m.Sequential, "causal": m.Causal} {
		r.put(name, map[bool]float64{true: 1, false: 0}[ok])
	}
	sortMetrics(r.Metrics)
	return m, r, nil
}

func sortMetrics(ms []Metric) {
	for i := 1; i < len(ms); i++ {
		for j := i; j > 0 && ms[j].Name < ms[j-1].Name; j-- {
			ms[j], ms[j-1] = ms[j-1], ms[j]
		}
	}
}

// orderExists searches for an order of the ops in set that keeps every pair
// before(i, j) in order and where each read returns the latest value written
// to its key so far (0 at first).
func orderExists(ops []Op, set []int, before func(i, j int) bool) bool {
	used := make([]bool, len(ops))
	val := map[string]int{}
	var try func(n int) bool
	try = func(n int) bool {
		if n == len(set) {
			return true
		}
		for _, i := range set {
			if used[i] {
				continue
			}
			ready := true
			for _, j := range set {
				if !used[j] && j != i && before(j, i) {
					ready = false
					break
				}
			}
			if !ready {
				continue
			}
			o := ops[i]
			if !o.Write && val[o.Key] != o.Value {
				continue
			}
			old, had := val[o.Key]
			if o.Write {
				val[o.Key] = o.Value
			}
			used[i] = true
			if try(n + 1) {
				return true
			}
			used[i] = false
			if had {
				val[o.Key] = old
			} else {
				delete(val, o.Key)
			}
		}
		return false
	}
	return try(0)
}

// String writes the models a history satisfies.
func (m Models) String() string {
	return fmt.Sprintf("linearizable %v, sequential %v, causal %v, eventual %v", m.Linearizable, m.Sequential, m.Causal, m.Eventual)
}
