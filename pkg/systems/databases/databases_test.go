package databases

import (
	"math/rand"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// checkBPlus verifies a B+ tree's invariants: sorted keys within separator
// bounds, all leaves at one depth, occupancy, and linked leaves in order.
func checkBPlus(t *testing.T, tr *BPlusTree) {
	t.Helper()
	leafDepth := -1
	var walk func(n *bpNode, depth int, lo, hi *int)
	walk = func(n *bpNode, depth int, lo, hi *int) {
		if !slices.IsSorted(n.keys) {
			t.Fatalf("unsorted node %v", n.keys)
		}
		if len(n.keys) >= tr.order {
			t.Fatalf("overfull node %v", n.keys)
		}
		if n != tr.root && tr.underflows(n) {
			t.Fatalf("underfull node %v (order %d)", n.keys, tr.order)
		}
		for _, k := range n.keys {
			if (lo != nil && k < *lo) || (hi != nil && k >= *hi) {
				t.Fatalf("key %d outside its separators", k)
			}
		}
		if n.leaf {
			if leafDepth < 0 {
				leafDepth = depth
			} else if depth != leafDepth {
				t.Fatalf("leaves at depths %d and %d", leafDepth, depth)
			}
			return
		}
		if len(n.children) != len(n.keys)+1 {
			t.Fatalf("%d keys but %d children", len(n.keys), len(n.children))
		}
		for i, c := range n.children {
			clo, chi := lo, hi
			if i > 0 {
				clo = &n.keys[i-1]
			}
			if i < len(n.keys) {
				chi = &n.keys[i]
			}
			walk(c, depth+1, clo, chi)
		}
	}
	walk(tr.root, 0, nil, nil)
}

func TestBPlusTreeAgainstAModel(t *testing.T) {
	for _, order := range []int{3, 4, 5, 8} {
		rng := rand.New(rand.NewSource(int64(order)))
		tr := NewBPlusTree(order)
		model := map[int]int{}
		for step := 0; step < 3000; step++ {
			k := rng.Intn(200)
			if rng.Intn(3) == 0 {
				_, had := model[k]
				if got := tr.Delete(k); got != had {
					t.Fatalf("order %d: Delete(%d) = %v, want %v", order, k, got, had)
				}
				delete(model, k)
			} else {
				_, had := model[k]
				if got := tr.Insert(k, step); got == had {
					t.Fatalf("order %d: Insert(%d) = %v, want %v", order, k, got, !had)
				}
				model[k] = step
			}
			checkBPlus(t, tr)
		}
		want := make([]int, 0, len(model))
		for k := range model {
			want = append(want, k)
		}
		sort.Ints(want)
		if got := tr.Keys(); !slices.Equal(got, want) || tr.Len() != len(want) {
			t.Fatalf("order %d: keys %v, want %v", order, got, want)
		}
		for k, v := range model {
			if got, ok := tr.Search(k, trace.Nop{}); !ok || got != v {
				t.Fatalf("Search(%d) = %d, %v", k, got, ok)
			}
		}
		var inRange []int
		for _, k := range want {
			if k >= 50 && k <= 120 {
				inRange = append(inRange, k)
			}
		}
		if got := tr.Range(50, 120, trace.Nop{}); !slices.Equal(got, inRange) {
			t.Fatalf("Range = %v, want %v", got, inRange)
		}
	}
}

func TestLSMTree(t *testing.T) {
	l := NewLSMTree(3, 2)
	for i := 1; i <= 10; i++ {
		l.Put(i, i*10)
	}
	l.Delete(4)
	l.Put(5, 55)
	for i := 1; i <= 10; i++ {
		v, ok := l.Get(i, trace.Nop{})
		switch i {
		case 4:
			if ok {
				t.Errorf("deleted key 4 found")
			}
		case 5:
			if !ok || v != 55 {
				t.Errorf("Get(5) = %d, %v; want the newest value 55", v, ok)
			}
		default:
			if !ok || v != i*10 {
				t.Errorf("Get(%d) = %d, %v", i, v, ok)
			}
		}
	}
	if l.Stats.Compactions == 0 || l.Runs() > 3 {
		t.Errorf("runs %d, compactions %d", l.Runs(), l.Stats.Compactions)
	}
	if _, ok := l.Get(999, trace.Nop{}); ok {
		t.Error("missing key found")
	}
}

func TestWALRecovery(t *testing.T) {
	rec, err := RecoverWAL("T1 write x 5, T2 write y 7, T1 commit, flush y, T3 write z 9, T3 abort, crash, T2 commit")
	if err != nil {
		t.Fatal(err)
	}
	if rec.Recovered["x"] != 5 {
		t.Errorf("committed x = %d, want 5 (redo)", rec.Recovered["x"])
	}
	if rec.Disk["y"] != 7 || rec.Recovered["y"] != 0 {
		t.Errorf("y on disk %d, recovered %d; want 7 then 0 (undo)", rec.Disk["y"], rec.Recovered["y"])
	}
	if rec.Recovered["z"] != 0 {
		t.Errorf("aborted z = %d", rec.Recovered["z"])
	}
	if !slices.Equal(rec.Losers, []string{"T2"}) {
		t.Errorf("losers %v", rec.Losers)
	}
	if _, err := RecoverWAL("nonsense"); err == nil {
		t.Error("bad script accepted")
	}
}

func TestSlottedPage(t *testing.T) {
	p := NewSlottedPage(40)
	a, _ := p.Insert("alpha")
	b, _ := p.Insert("bravo")
	c, _ := p.Insert("charlie")
	p.Delete(b)
	if p.Holes() != 5 {
		t.Errorf("holes %d", p.Holes())
	}
	d, err := p.Insert("dog")
	if err != nil || d != b {
		t.Fatalf("Insert reused slot %d, want %d (%v)", d, b, err)
	}
	if _, err := p.Insert(strings.Repeat("x", 30)); err == nil {
		t.Error("an oversized record fit")
	}
	p.Delete(a)
	p.Compact()
	if p.Holes() != 0 {
		t.Errorf("holes after compaction %d", p.Holes())
	}
	if r, ok := p.Get(c); !ok || r != "charlie" {
		t.Errorf("slot %d holds %q after compaction", c, r)
	}
}

func TestBufferPool(t *testing.T) {
	for _, policy := range []string{Clock, LRUK} {
		b, err := NewBufferPool(3, policy, 2)
		if err != nil {
			t.Fatal(err)
		}
		for _, p := range []int{1, 2, 3, 1, 4, 1, 5} {
			if _, err := b.Fetch(p, false, false, trace.Nop{}); err != nil {
				t.Fatal(err)
			}
		}
		if b.Stats.Hits+b.Stats.Misses != 7 || b.Stats.Hits < 1 {
			t.Errorf("%s: %+v", policy, b.Stats)
		}
		// Clock evicts page 1 at the first miss, every reference bit being
		// set; LRU-K keeps it, the only page used twice.
		if policy == LRUK && b.find(1) < 0 {
			t.Errorf("%s evicted page 1, the hot page", policy)
		}
		for _, p := range []int{6, 7, 8} {
			_, _ = b.Fetch(p, true, false, trace.Nop{})
		}
		if _, err := b.Fetch(9, false, false, trace.Nop{}); err == nil {
			t.Errorf("%s: fetched into a pool of pinned frames", policy)
		}
	}
	if _, err := NewBufferPool(1, "fifo", 1); err == nil {
		t.Error("unknown policy accepted")
	}
}

func TestHashIndexes(t *testing.T) {
	x := NewExtendibleHash(2)
	l := NewLinearHash(2, 2)
	keys := []int{1, 5, 9, 13, 2, 6, 3, 7, 11, 21, 40, 64}
	for _, k := range keys {
		if ok, err := x.Insert(k); !ok || err != nil {
			t.Fatalf("extendible Insert(%d) = %v, %v", k, ok, err)
		}
		if !l.Insert(k) {
			t.Fatalf("linear Insert(%d)", k)
		}
	}
	for _, k := range keys {
		if !x.Search(k, trace.Nop{}) || !l.Search(k, trace.Nop{}) {
			t.Fatalf("lost %d", k)
		}
	}
	if x.Search(100, trace.Nop{}) || l.Search(100, trace.Nop{}) {
		t.Error("found a missing key")
	}
	if x.GlobalDepth() < 2 || l.Splits == 0 {
		t.Errorf("no growth: depth %d, splits %d", x.GlobalDepth(), l.Splits)
	}
	if !x.Delete(9) || x.Search(9, trace.Nop{}) || !l.Delete(9) || l.Search(9, trace.Nop{}) {
		t.Error("delete failed")
	}
	y := NewExtendibleHash(1)
	_, _ = y.Insert(0)
	if _, err := y.Insert(1 << 12); err == nil {
		t.Error("keys sharing every low bit did not fail")
	}
}

const (
	employees = "id,name,dept\n1,Ama,10\n2,Kofi,20\n3,Esi,10\n4,Yaw,30\n5,Abena,20"
	depts     = "dept,title\n10,Sales\n20,Research\n40,Legal\n10,Support"
)

func TestJoinsAgree(t *testing.T) {
	l, _ := ParseTable(employees)
	r, _ := ParseTable(depts)
	norm := func(tb Table) []string {
		out := tb.Lines()[1:]
		sort.Strings(out)
		return out
	}
	want, st, err := NestedLoopJoin(l, r, "dept", "dept", trace.Nop{})
	if err != nil {
		t.Fatal(err)
	}
	if len(want.Rows) != 6 || st.Comparisons != 20 || st.InnerScans != 5 {
		t.Fatalf("nested loop: %d rows, %+v", len(want.Rows), st)
	}
	joins := map[string]func() (Table, JoinStats, error){
		"block": func() (Table, JoinStats, error) { return BlockNestedLoopJoin(l, r, "dept", "dept", 2, trace.Nop{}) },
		"index": func() (Table, JoinStats, error) { return IndexNestedLoopJoin(l, r, "dept", "dept", trace.Nop{}) },
		"hash":  func() (Table, JoinStats, error) { return HashJoin(l, r, "dept", "dept", trace.Nop{}) },
		"merge": func() (Table, JoinStats, error) { return SortMergeJoin(l, r, "dept", "dept", trace.Nop{}) },
	}
	for name, join := range joins {
		got, st, err := join()
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(norm(got), norm(want)) {
			t.Errorf("%s join: %v, want %v", name, norm(got), norm(want))
		}
		if name == "block" && st.InnerScans != 3 {
			t.Errorf("block join scanned the inner table %d times, want 3", st.InnerScans)
		}
		if name == "hash" && st.Comparisons >= 20 {
			t.Errorf("hash join compared %d times", st.Comparisons)
		}
	}
	if got := want.Columns; !slices.Equal(got, []string{"id", "name", "dept", "r.dept", "title"}) {
		t.Errorf("columns %v", got)
	}
	if _, _, err := HashJoin(l, r, "nope", "dept", trace.Nop{}); err == nil {
		t.Error("unknown column accepted")
	}
}

func TestExternalSort(t *testing.T) {
	in := []int{9, 3, 7, 1, 8, 2, 6, 4, 5, 0}
	var log trace.Log
	got, passes := ExternalSort(in, 3, 2, &log)
	if !slices.IsSorted(got) || len(got) != len(in) {
		t.Fatalf("got %v", got)
	}
	if len(passes) != 3 || len(passes[0].Runs) != 4 || len(passes[2].Runs) != 1 {
		t.Errorf("passes %v", passes)
	}
	if replayed := trace.Replay(slices.Clone(in), log.Events); !slices.Equal(replayed, got) {
		t.Errorf("replay %v", replayed)
	}
	if in[0] != 9 {
		t.Error("the input was changed")
	}
}

func TestQueryAndPlan(t *testing.T) {
	tb, _ := ParseTable("name,city,age\nAma,Accra,30\nKofi,Kumasi,41\nEsi,Accra,25\nYaw,Tamale,52\nAbena,Accra,35")
	stages, err := Query(tb, "select age > 26 | group city count | sort count desc | limit 1")
	if err != nil {
		t.Fatal(err)
	}
	last := stages[len(stages)-1].Table
	if len(last.Rows) != 1 || last.Rows[0][0] != "Accra" || last.Rows[0][1] != "2" {
		t.Errorf("got %v", last.Rows)
	}
	if _, err := Query(tb, "explode"); err == nil {
		t.Error("unknown operator accepted")
	}
	p, err := Plan(tb, "name = Esi", "name", 1, 4)
	if err != nil || p.Access != "index scan" || p.ActualRows != 1 {
		t.Errorf("selective plan %+v, %v", p.Access, err)
	}
	p, _ = Plan(tb, "age > 0", "age", 1, 4)
	if p.Access != "full scan" || p.ActualRows != 5 {
		t.Errorf("broad plan chose %s", p.Access)
	}
}

func TestTwoPhaseLockingDeadlock(t *testing.T) {
	ops, _ := ParseSchedule("r1(x) r2(y) w1(y) w2(x) c1 c2")
	run := TwoPhaseLocking(ops)
	if run.Deadlocks != 1 || !slices.Equal(run.Aborted, []string{"T2"}) || !slices.Equal(run.Committed, []string{"T1"}) {
		t.Errorf("deadlocks %d, aborted %v, committed %v", run.Deadlocks, run.Aborted, run.Committed)
	}
	ops, _ = ParseSchedule("T1 read x, T2 read x, T1 write y, T2 commit, T1 commit")
	if run := TwoPhaseLocking(ops); run.Deadlocks != 0 || len(run.Committed) != 2 {
		t.Errorf("shared reads: %+v", run)
	}
}

func TestPrecedence(t *testing.T) {
	ops, _ := ParseSchedule("r1(x) w2(x) w1(x) c1 c2") // the lost-update schedule
	g := Precedence(ops, trace.Nop{})
	if g.Serializable || len(g.Cycle) != 2 {
		t.Errorf("found no cycle: %+v", g)
	}
	ops, _ = ParseSchedule("r1(x) w1(x) r2(x) w2(x) r1(y)")
	g = Precedence(ops, trace.Nop{})
	if !g.Serializable || !slices.Equal(g.Order, []int{0, 1}) {
		t.Errorf("serial schedule: %+v", g)
	}
}

func TestTimestampOrdering(t *testing.T) {
	ops, _ := ParseSchedule("r1(x) w2(x) w1(x) c2 c1")
	if run := TimestampOrdering(ops, false); !slices.Equal(run.Aborted, []string{"T1"}) {
		t.Errorf("basic: aborted %v", run.Aborted)
	}
	ops, _ = ParseSchedule("w2(x) w1(x) c1 c2")
	if run := TimestampOrdering(ops, false); len(run.Aborted) != 0 {
		t.Errorf("in-order writes aborted %v", run.Aborted)
	}
	ops, _ = ParseSchedule("r1(y) w2(x) w1(x) c1 c2")
	if run := TimestampOrdering(ops, true); len(run.Aborted) != 0 {
		t.Errorf("Thomas write rule aborted %v", run.Aborted)
	}
}

func TestSnapshotIsolation(t *testing.T) {
	ops, err := ParseSchedule("T1 begin, T2 begin, T2 write x 5, T2 commit, T1 read x, T1 write x 7, T1 commit, T3 read x")
	if err != nil {
		t.Fatal(err)
	}
	run := SnapshotIsolation(ops)
	if !slices.Equal(run.Aborted, []string{"T1"}) || len(run.Versions["x"]) != 1 {
		t.Errorf("aborted %v, versions %v", run.Aborted, run.Versions)
	}
	if out := run.Steps[4].Outcome; !strings.Contains(out, "x = 0") {
		t.Errorf("T1 read outside its snapshot: %s", out)
	}
	if out := run.Steps[len(run.Steps)-1].Outcome; !strings.Contains(out, "x = 5") {
		t.Errorf("T3 read %s", out)
	}
}

func TestIsolationAnomalies(t *testing.T) {
	want := map[Level][]bool{
		ReadUncommitted: {true, true, true, true, true},
		ReadCommitted:   {false, true, true, true, true},
		RepeatableRead:  {false, false, false, false, true},
		Serializable:    {false, false, false, false, false},
	}
	for _, l := range Levels() {
		for i, a := range IsolationAnomalies(l) {
			if a.Name != AnomalyNames[i] || a.Occurs != want[l][i] {
				t.Errorf("%s, %s: occurs %v, want %v\n%s", l, a.Name, a.Occurs, want[l][i], strings.Join(a.Story, "\n"))
			}
		}
	}
	if l, err := ParseLevel("Read Committed"); err != nil || l != ReadCommitted {
		t.Errorf("ParseLevel = %v, %v", l, err)
	}
}

func TestParseSchedule(t *testing.T) {
	ops, err := ParseSchedule("r1(x) w2(y), T3 write z 4, T3 commit")
	if err != nil || len(ops) != 4 || ops[2].Value != 4 || ops[1].String() != "W2(y)" {
		t.Errorf("%v, %v", ops, err)
	}
	for _, bad := range []string{"", "q1(x)", "r1", "T1 jump", "T1 write"} {
		if _, err := ParseSchedule(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
	if _, err := ParseTable("a,b\n1"); err == nil {
		t.Error("short row accepted")
	}
}
