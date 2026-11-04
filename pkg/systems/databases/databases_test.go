package databases

import (
	"errors"
	"math/rand"
	"reflect"
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

func TestRelationalOperators(t *testing.T) {
	tb, err := ParseTable("name,city,age\nAma,Accra,30\nKofi,Kumasi,41\nEsi,Accra,25\nYaw,Tamale,52\nAbena,Accra,35\nAma,Accra,99")
	if err != nil {
		t.Fatal(err)
	}
	t.Run("every comparison", func(t *testing.T) {
		want := map[string][]string{
			"age = 30":       {"Ama"},
			"age != 30":      {"Kofi", "Esi", "Yaw", "Abena", "Ama"},
			"age < 30":       {"Esi"},
			"age <= 30":      {"Ama", "Esi"},
			"age > 41":       {"Yaw", "Ama"},
			"age>=41":        {"Kofi", "Yaw", "Ama"},
			"city = Accra":   {"Ama", "Esi", "Abena", "Ama"},
			"name < B":       {"Ama", "Abena", "Ama"},
			" city != Accra": {"Kofi", "Yaw"},
		}
		for cond, names := range want {
			p, err := ParsePredicate(cond)
			if err != nil {
				t.Fatalf("%q: %v", cond, err)
			}
			got, err := Select(tb, p)
			if err != nil {
				t.Fatal(err)
			}
			var gotNames []string
			for _, r := range got.Rows {
				gotNames = append(gotNames, r[0])
			}
			if !slices.Equal(gotNames, names) {
				t.Errorf("%q: %v, want %v", cond, gotNames, names)
			}
		}
		for _, bad := range []string{"age", "= 3", "age >", ""} {
			if _, err := ParsePredicate(bad); !errors.Is(err, ErrInput) {
				t.Errorf("ParsePredicate(%q): %v", bad, err)
			}
		}
		if _, err := Select(tb, Predicate{"height", ">", "1"}); !errors.Is(err, ErrInput) {
			t.Errorf("an unknown column: %v", err)
		}
	})
	t.Run("project drops duplicate rows", func(t *testing.T) {
		got, err := Project(tb, []string{"city", "NAME"})
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(got.Columns, []string{"city", "name"}) || len(got.Rows) != 5 || got.Rows[0][1] != "Ama" {
			t.Errorf("project %v %v", got.Columns, got.Rows)
		}
		if _, err := Project(tb, []string{"zip"}); !errors.Is(err, ErrInput) {
			t.Errorf("an unknown column: %v", err)
		}
	})
	t.Run("every aggregate", func(t *testing.T) {
		want := map[string][][]string{
			"count": {{"Accra", "4"}, {"Kumasi", "1"}, {"Tamale", "1"}},
			"sum":   {{"Accra", "189"}, {"Kumasi", "41"}, {"Tamale", "52"}},
			"avg":   {{"Accra", "47"}, {"Kumasi", "41"}, {"Tamale", "52"}}, // 189 / 4 rounds down
			"min":   {{"Accra", "25"}, {"Kumasi", "41"}, {"Tamale", "52"}},
			"max":   {{"Accra", "99"}, {"Kumasi", "41"}, {"Tamale", "52"}},
		}
		for agg, rows := range want {
			got, err := GroupBy(tb, "city", agg, "age")
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got.Rows, rows) {
				t.Errorf("%s: %v, want %v", agg, got.Rows, rows)
			}
			name := agg + "(age)"
			if agg == "count" {
				name = "count"
			}
			if got.Columns[1] != name {
				t.Errorf("%s column %q", agg, got.Columns[1])
			}
		}
		for name, f := range map[string]func() error{
			"unknown aggregate": func() error { _, err := GroupBy(tb, "city", "median", "age"); return err },
			"unknown group":     func() error { _, err := GroupBy(tb, "zip", "count", ""); return err },
			"unknown of":        func() error { _, err := GroupBy(tb, "city", "sum", "zip"); return err },
			"sum of text":       func() error { _, err := GroupBy(tb, "city", "sum", "name"); return err },
		} {
			if err := f(); !errors.Is(err, ErrInput) {
				t.Errorf("%s: %v", name, err)
			}
		}
	})
	t.Run("a whole query", func(t *testing.T) {
		stages, err := Query(tb, "where city = Accra | project name, age | order age desc | limit 2 |")
		if err != nil {
			t.Fatal(err)
		}
		if len(stages) != 4 {
			t.Fatalf("%d stages", len(stages))
		}
		last := stages[3].Table
		if !reflect.DeepEqual(last.Rows, [][]string{{"Ama", "99"}, {"Abena", "35"}}) {
			t.Errorf("rows %v", last.Rows)
		}
		if stages[0].Op != "where city = Accra" || len(stages[0].Table.Rows) != 4 {
			t.Errorf("first stage %q with %d rows", stages[0].Op, len(stages[0].Table.Rows))
		}
		asc, _ := Query(tb, "sort age | limit 100")
		if asc[1].Table.Rows[0][2] != "25" || len(asc[1].Table.Rows) != 6 {
			t.Errorf("ascending sort, limit past the end: %v", asc[1].Table.Rows)
		}
		sum, _ := Query(tb, "group city sum age")
		if sum[0].Table.Rows[0][1] != "189" {
			t.Errorf("group sum: %v", sum[0].Table.Rows)
		}
		for _, bad := range []string{"", " | ", "select age", "group city", "group a b c d", "sort", "sort zip", "limit -1", "limit x", "project zip", "drop name"} {
			if _, err := Query(tb, bad); !errors.Is(err, ErrInput) {
				t.Errorf("Query(%q): %v", bad, err)
			}
		}
	})
	if _, err := ParseTable("\n\n"); !errors.Is(err, ErrInput) {
		t.Errorf("a table with no header: %v", err)
	}
}

func TestJoinColumns(t *testing.T) {
	l, _ := ParseTable(employees)
	r, _ := ParseTable(depts)
	cases := []struct {
		on, lc, rc string
		ok         bool
	}{
		{"id = dept_id", "id", "dept_id", true},
		{" dept ", "dept", "dept", true},
		{"", "dept", "dept", true},
	}
	for _, c := range cases {
		lc, rc, ok := JoinColumns(l, r, c.on)
		if lc != c.lc || rc != c.rc || ok != c.ok {
			t.Errorf("JoinColumns(%q) = %q %q %v", c.on, lc, rc, ok)
		}
	}
	other, _ := ParseTable("x,y\n1,2")
	if _, _, ok := JoinColumns(l, other, ""); ok {
		t.Error("tables with no column in common found one")
	}
}

func TestStructureSnapshots(t *testing.T) {
	t.Run("B+ tree", func(t *testing.T) {
		tr := NewBPlusTree(2) // held to 3
		if tr.Order() != 3 || tr.Height() != 1 || len(tr.Snapshot().Nodes) != 0 {
			t.Errorf("empty: order %d height %d", tr.Order(), tr.Height())
		}
		for k := 1; k <= 20; k++ {
			tr.Insert(k, k*k)
		}
		checkBPlus(t, tr)
		s := tr.Snapshot()
		var leaves, nexts int
		var walked []string
		for _, n := range s.Nodes {
			if n.Role == "leaf" {
				leaves++
				walked = append(walked, n.Label)
			}
		}
		for _, e := range s.Edges {
			if e.Label == "next" {
				nexts++
			}
		}
		if s.Nodes[0].Role != "root" || leaves < 2 || nexts != leaves-1 || tr.Height() < 3 {
			t.Errorf("root %q, %d leaves, %d next edges, height %d", s.Nodes[0].Role, leaves, nexts, tr.Height())
		}
		var keys []string
		for _, l := range walked {
			keys = append(keys, strings.Fields(l)...)
		}
		if len(keys) != 20 || keys[0] != "1" || keys[19] != "20" {
			t.Errorf("the leaves hold %v", keys)
		}
		if len(s.Edges)-nexts != len(s.Nodes)-1 {
			t.Errorf("%d parent edges for %d nodes", len(s.Edges)-nexts, len(s.Nodes))
		}
	})
	t.Run("buffer pool", func(t *testing.T) {
		b, _ := NewBufferPool(3, Clock, 1)
		if b.Unpin(1) {
			t.Error("unpinned a page not in memory")
		}
		_, _ = b.Fetch(1, true, true, trace.Nop{})
		_, _ = b.Fetch(2, false, false, trace.Nop{})
		labels := []string{}
		for _, n := range b.Snapshot().Nodes {
			labels = append(labels, n.Label)
		}
		if want := []string{"page 1 (pinned 1, ref, dirty)", "page 2 (ref)", "empty"}; !slices.Equal(labels, want) {
			t.Errorf("labels %q, want %q", labels, want)
		}
		if b.Snapshot().Nodes[0].Role != "hand" {
			t.Error("the clock hand is not shown")
		}
		if !b.Unpin(1) || b.Unpin(1) {
			t.Error("one pin, one unpin")
		}
		l, _ := NewBufferPool(1, LRUK, 2)
		_, _ = l.Fetch(5, false, false, trace.Nop{})
		if n := l.Snapshot().Nodes[0]; n.Label != "page 5" || n.Role != "" {
			t.Errorf("LRU-K frame %+v", n)
		}
	})
	t.Run("extendible hash", func(t *testing.T) {
		h := NewExtendibleHash(1)
		if s := h.Snapshot(); len(s.Nodes) != 2 || s.Nodes[0].Label != "dir" || s.Nodes[1].Label != "d=0: (empty)" {
			t.Errorf("empty %+v", s)
		}
		for _, k := range []int{0, 1, 2, 3} {
			if _, err := h.Insert(k); err != nil {
				t.Fatal(err)
			}
		}
		s := h.Snapshot()
		dirs := 1 << h.GlobalDepth()
		if len(s.Edges) != dirs || s.Nodes[0].Role != "global depth 2" || s.Nodes[1].Label != "dir 01" {
			t.Errorf("%d edges for %d entries, %+v", len(s.Edges), dirs, s.Nodes[:2])
		}
		keys := 0
		for _, n := range s.Nodes[dirs:] {
			keys += len(strings.Fields(n.Label[strings.Index(n.Label, ":")+1:]))
		}
		if keys != 4 {
			t.Errorf("buckets hold %d keys, want 4", keys)
		}
	})
	t.Run("linear hash", func(t *testing.T) {
		h := NewLinearHash(1, 1)
		for _, k := range []int{1, 2, 3, 4, 5} {
			h.Insert(k)
		}
		s := h.Snapshot()
		next := 0
		for _, n := range s.Nodes {
			if n.Role == "next" {
				next++
			}
		}
		if next != 1 && h.next < len(s.Nodes) {
			t.Errorf("%d nodes marked next", next)
		}
		if !h.Delete(3) || h.Delete(3) || h.Search(3, trace.Nop{}) {
			t.Error("delete")
		}
	})
	t.Run("LSM tree", func(t *testing.T) {
		l := NewLSMTree(2, 10)
		l.Put(1, 10)
		l.Put(2, 20) // flushes run 0
		l.Delete(1)  // stays in the memtable
		s := l.Snapshot()
		if len(s.Nodes) != 2 || s.Nodes[0].Label != "memtable: 1=x" || s.Nodes[0].Role != "head" || s.Nodes[1].Label != "run 0: 1=10 2=20" {
			t.Errorf("%+v", s.Nodes)
		}
		if len(s.Edges) != 1 || s.Edges[0].Label != "older" {
			t.Errorf("edges %+v", s.Edges)
		}
	})
	t.Run("slotted page", func(t *testing.T) {
		p := NewSlottedPage(64)
		a, _ := p.Insert("alpha")
		_, _ = p.Insert("bravo")
		p.Delete(a)
		if p.Delete(a) || p.Delete(99) {
			t.Error("deleted a slot twice")
		}
		var labels []string
		for _, n := range p.Snapshot().Nodes {
			labels = append(labels, n.Label)
		}
		joined := strings.Join(labels, "|")
		if labels[0] != "header: 2 slots" || !strings.Contains(joined, "slot 0: free") || !strings.Contains(joined, "slot 1: @") || !strings.Contains(joined, "free ") {
			t.Errorf("labels %q", labels)
		}
	})
}

func TestWALLog(t *testing.T) {
	rec, err := RecoverWAL("T1 write x 5; T1 read x; T2 write x 8\nT2 abort, flush x, flush q, crash")
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	for _, r := range rec.Log {
		lines = append(lines, r.String())
	}
	want := []string{"1: T1 begin", "2: T1 update x 0 -> 5", "3: T2 begin", "4: T2 update x 5 -> 8", "5: T2 undo x back to 5", "6: T2 abort"}
	if !slices.Equal(lines, want) {
		t.Errorf("log\n%s\nwant\n%s", strings.Join(lines, "\n"), strings.Join(want, "\n"))
	}
	if !slices.Equal(rec.Losers, []string{"T1"}) || rec.Disk["x"] != 5 || rec.Recovered["x"] != 0 {
		t.Errorf("losers %v, disk %v, recovered %v", rec.Losers, rec.Disk, rec.Recovered)
	}
	if !slices.Equal(rec.Items(), []string{"x"}) {
		t.Errorf("items %v", rec.Items())
	}
	for _, bad := range []string{"flush", "T1 commit, T1 write x 1", "T1 fly"} {
		if _, err := RecoverWAL(bad); !errors.Is(err, ErrInput) {
			t.Errorf("RecoverWAL(%q): %v", bad, err)
		}
	}
}

func TestTwoPhaseLockingWaitsAndAborts(t *testing.T) {
	// T1 never commits: T2 waits on it until T1 runs out of operations and
	// commits at the end, releasing its lock.
	ops, _ := ParseSchedule("w1(x) r2(x) c2")
	run := TwoPhaseLocking(ops)
	if !slices.Equal(run.Committed, []string{"T1", "T2"}) || run.Deadlocks != 0 {
		t.Errorf("committed %v, deadlocks %d", run.Committed, run.Deadlocks)
	}
	waited := false
	for _, s := range run.Steps {
		if slices.Contains(s.Waiting, "T2") {
			waited = true
		}
	}
	if !waited {
		t.Error("T2 never waited for T1's lock")
	}
	ops, _ = ParseSchedule("w1(x) a1 r2(x) c2 b3")
	run = TwoPhaseLocking(ops)
	if !slices.Equal(run.Aborted, []string{"T1"}) || !slices.Equal(run.Committed, []string{"T2", "T3"}) {
		t.Errorf("aborted %v, committed %v", run.Aborted, run.Committed)
	}
}

func TestNames(t *testing.T) {
	if Level(9).String() != "level(9)" || Action(9).String() != "?" || Write.String() != "W" {
		t.Error("names")
	}
	for _, s := range []string{"2", " REPEATABLE READ "} {
		if l, err := ParseLevel(s); err != nil || l != RepeatableRead {
			t.Errorf("ParseLevel(%q) = %v %v", s, l, err)
		}
	}
	if _, err := ParseLevel("snapshot"); !errors.Is(err, ErrInput) {
		t.Errorf("an unknown level: %v", err)
	}
}
