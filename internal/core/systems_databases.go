package core

import (
	"fmt"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/systems/databases"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// databasesEntries is every entry of the databases category.
func databasesEntries() []Entry {
	return databaseEntries()
}

// The example data the database entries share.
const (
	dbEmployees = "id,name,dept\n1,Ama,10\n2,Kofi,20\n3,Esi,10\n4,Yaw,30\n5,Abena,20"
	dbDepts     = "dept,title\n10,Sales\n20,Research\n40,Legal\n10,Support"
	dbPeople    = "name,city,age\nAma,Accra,30\nKofi,Kumasi,41\nEsi,Accra,25\nYaw,Tamale,52\nAbena,Accra,35\nKwame,Kumasi,28"
	dbMaxText   = 20_000
)

func databasesEntryFile(name string) string { return "systems/databases/" + name + ".go" }

var databasesEntryIdx = []struct {
	id, name, file, fn, summary, time, space, fill string
	ops                                            []string
	options                                        []Option
	build                                          func(in Input) playground
	example                                        Input
}{
	{"b-plus-tree", "B+ Tree Index", databasesEntryFile("bplustree"), "BPlusTree",
		"The index on almost every database column: wide inner nodes that only guide the search, every key in the leaves, and the leaves linked for range scans.",
		"O(log_m n) search, insert and delete; O(log_m n + k) for a range of k keys", "O(n)", "insert",
		[]string{"insert KEY", "delete KEY", "search KEY", "range LO HI"}, []Option{{"order", "order m (most children)", 3, 3, 16}},
		func(in Input) playground { return &bplusP{t: databases.NewBPlusTree(in.opt("order", 3))} },
		Input{Values: []int{10, 20, 5}, Ops: ParseOps("insert 6, insert 12, insert 30, insert 7, insert 17, search 12, range 6 17, delete 6, delete 7"), Options: map[string]int{"order": 3}}},
	{"lsm-tree", "LSM Tree", databasesEntryFile("lsm"), "LSMTree",
		"Writes into a memtable flushed as sorted runs, reads newest first with a Bloom filter per run, and compaction merging runs: the engine of LevelDB and RocksDB.",
		"O(1) amortized writes; O(r log n) reads over r runs, most skipped by Bloom filters", "O(n)", "put",
		[]string{"put KEY [VALUE]", "delete KEY", "get KEY", "flush", "compact"},
		[]Option{{"memtable", "keys per memtable", 3, 1, 100}, {"runs", "runs before compaction", 3, 1, 20}},
		func(in Input) playground {
			return &lsmP{t: databases.NewLSMTree(in.opt("memtable", 3), in.opt("runs", 3))}
		},
		Input{Values: []int{1, 2, 3, 4, 5}, Ops: ParseOps("put 2 20, delete 3, get 2, get 3, get 9, put 6, put 7, compact, get 1"), Options: map[string]int{"memtable": 3, "runs": 3}}},
	{"slotted-page", "Slotted Page", databasesEntryFile("slottedpage"), "SlottedPage",
		"How a page holds variable-length records: a slot directory from the front, records from the back, holes left by deletes, and compaction.",
		"O(s) insert for s slots, O(1) get and delete, O(n) compaction", "O(page size)", "insert",
		[]string{"insert TEXT", "delete SLOT", "get SLOT", "compact"}, []Option{{"size", "page size in bytes", 64, 16, 4096}},
		func(in Input) playground { return &pageP{p: databases.NewSlottedPage(in.opt("size", 64))} },
		Input{Ops: ParseOps("insert alice, insert bob, insert carol, delete 1, insert dave, delete 0, compact, get 2"), Options: map[string]int{"size": 48}}},
	{"buffer-pool-clock", "Buffer Pool (Clock)", databasesEntryFile("bufferpool"), "BufferPool",
		"Disk pages cached in a fixed number of frames: hits, misses, pins, dirty write-backs, and a clock hand giving each page a second chance.",
		"O(n) per miss for n frames, O(n) per hit to find the frame", "O(n)", "read",
		[]string{"read PAGE", "write PAGE", "pin PAGE", "unpin PAGE"}, []Option{{"frames", "frames", 3, 1, 64}},
		func(in Input) playground { return newPoolP(in.opt("frames", 3), databases.Clock, 1) },
		Input{Ops: ParseOps("read 1, read 2, write 3, read 1, pin 2, read 4, read 5, unpin 2, read 1, read 6"), Options: map[string]int{"frames": 3}}},
	{"buffer-pool-lru-k", "Buffer Pool (LRU-K)", databasesEntryFile("bufferpool"), "BufferPool",
		"The same buffer pool evicting the page whose K-th most recent use is oldest, so one big scan cannot push out the pages used again and again.",
		"O(n) per miss for n frames", "O(n + pages seen)", "read",
		[]string{"read PAGE", "write PAGE", "pin PAGE", "unpin PAGE"}, []Option{{"frames", "frames", 3, 1, 64}, {"k", "uses remembered (K)", 2, 1, 8}},
		func(in Input) playground { return newPoolP(in.opt("frames", 3), databases.LRUK, in.opt("k", 2)) },
		Input{Ops: ParseOps("read 1, read 1, read 2, read 2, read 3, read 4, read 5, read 1, read 2"), Options: map[string]int{"frames": 3, "k": 2}}},
	{"extendible-hashing", "Extendible Hashing", databasesEntryFile("extendiblehash"), "ExtendibleHash",
		"A directory of 2^g pointers to buckets read by a key's low bits: a full bucket splits on one more bit, and the directory doubles only when it must.",
		"O(1) search, O(1) amortized insert", "O(n + 2^g)", "insert",
		[]string{"insert KEY", "delete KEY", "search KEY"}, []Option{{"capacity", "keys per bucket", 2, 1, 16}},
		func(in Input) playground { return &extHashP{h: databases.NewExtendibleHash(in.opt("capacity", 2))} },
		Input{Ops: ParseOps("insert 4, insert 8, insert 1, insert 12, insert 5, insert 3, insert 16, search 12, delete 8"), Options: map[string]int{"capacity": 2}}},
	{"linear-hashing", "Linear Hashing", databasesEntryFile("linearhash"), "LinearHash",
		"A hash table that grows one bucket at a time in a fixed order, with overflow chains until each bucket's turn to split comes.",
		"O(1) amortized search and insert", "O(n)", "insert",
		[]string{"insert KEY", "delete KEY", "search KEY"}, []Option{{"buckets", "starting buckets", 2, 1, 16}, {"capacity", "keys per bucket", 2, 1, 16}},
		func(in Input) playground {
			return &linHashP{h: databases.NewLinearHash(in.opt("buckets", 2), in.opt("capacity", 2))}
		},
		Input{Ops: ParseOps("insert 4, insert 8, insert 1, insert 12, insert 5, insert 3, insert 16, insert 7, search 12, delete 8"), Options: map[string]int{"buckets": 2, "capacity": 2}}},
}

func databasesEntryJoin(id, name, fn, summary, time string, run func(l, r databases.Table, lc, rc string, in Input, rec trace.Recorder) (databases.Table, databases.JoinStats, error)) Entry {
	e := entry(id, name, databasesEntryFile("joins"), fn)
	e.Summary, e.Time, e.Space, e.Input, e.Traced, e.Params = summary, time, "O(k) for k joined rows", TwoTexts, true, []string{"key"}
	e.Limits = fmt.Sprintf("at most %d characters per table", dbMaxText)
	e.Example = &Input{Text: dbEmployees, Other: dbDepts, Key: "dept"}
	e.Run = func(in Input, r trace.Recorder) (Output, error) {
		l, rt, lc, rc, err := joinTables(in)
		if err != nil {
			return Output{}, err
		}
		res, st, err := run(l, rt, lc, rc, in, r)
		if err != nil {
			return Output{}, wrap(err)
		}
		return Output{Lines: res.Lines(), Summary: map[string]float64{
			"rows": float64(len(res.Rows)), "comparisons": float64(st.Comparisons), "inner scans": float64(st.InnerScans), "probes": float64(st.Probes),
		}}, nil
	}
	return e
}

var databasesEntryNl = databasesEntryJoin("nested-loop-join", "Nested-Loop Join", "NestedLoopJoin",
	"Join two tables by comparing every row of one with every row of the other: simple, and slow on anything big.", "O(n * m)",
	func(l, r databases.Table, lc, rc string, _ Input, rec trace.Recorder) (databases.Table, databases.JoinStats, error) {
		return databases.NestedLoopJoin(l, r, lc, rc, rec)
	})

var databasesEntryInl = databasesEntryJoin("index-nested-loop-join", "Index Nested-Loop Join", "IndexNestedLoopJoin",
	"For each outer row, look the key up in an index on the inner table instead of scanning it.", "O(n log m + k)",
	func(l, r databases.Table, lc, rc string, _ Input, rec trace.Recorder) (databases.Table, databases.JoinStats, error) {
		return databases.IndexNestedLoopJoin(l, r, lc, rc, rec)
	})

var databasesEntryHj = databasesEntryJoin("hash-join", "Hash Join", "HashJoin",
	"Build a hash table on one table's join column, then probe it with each row of the other: the usual plan for big equi-joins.", "O(n + m + k) expected",
	func(l, r databases.Table, lc, rc string, _ Input, rec trace.Recorder) (databases.Table, databases.JoinStats, error) {
		return databases.HashJoin(l, r, lc, rc, rec)
	})

var databasesEntrySmj = databasesEntryJoin("sort-merge-join", "Sort-Merge Join", "SortMergeJoin",
	"Sort both tables on the join column, then walk them together like a merge, pairing runs of equal keys.", "O(n log n + m log m + k)",
	func(l, r databases.Table, lc, rc string, _ Input, rec trace.Recorder) (databases.Table, databases.JoinStats, error) {
		return databases.SortMergeJoin(l, r, lc, rc, rec)
	})

func writeAheadLogEntry() Entry {
	wal := entry("write-ahead-log", "Write-Ahead Log and Recovery", databasesEntryFile("wal"), "RecoverWAL")
	wal.Summary = "Log every change before making it, crash, then recover ARIES-style: analysis finds the losers, redo repeats history, undo rolls the losers back."
	wal.Time, wal.Space, wal.Input = "O(n)", "O(n)", Text
	wal.Limits = fmt.Sprintf("at most %d characters", dbMaxText)
	wal.Example = &Input{Text: "T1 write x 5, T2 write y 7, T1 commit, flush y, T3 write z 9, T3 abort, T2 write x 8, crash"}
	wal.Call = func(in Input) (string, string) { return "", fmt.Sprintf("databases.RecoverWAL(%q)", in.Text) }
	wal.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := textLimit(in.Text); err != nil {
			return Output{}, err
		}
		rec, err := databases.RecoverWAL(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		lines := []string{"the log:"}
		for _, l := range rec.Log {
			lines = append(lines, "  "+l.String())
		}
		lines = append(lines, "what happened:")
		for _, s := range rec.Steps {
			lines = append(lines, "  "+s)
		}
		items := rec.Items()
		tb := &TableOutput{Rows: len(items), Cols: 3, RowLabels: items, ColLabels: []string{"buffer at crash", "disk at crash", "recovered"}}
		for _, it := range items {
			mem, ok := rec.Memory[it]
			if !ok {
				mem = rec.Disk[it]
			}
			tb.Cells = append(tb.Cells, mem, rec.Disk[it], rec.Recovered[it])
		}
		return Output{Lines: lines, Table: tb, Summary: map[string]float64{"redone": float64(rec.Redone), "undone": float64(rec.Undone), "losers": float64(len(rec.Losers))}}, nil
	}
	return wal
}

func blockNestedLoopJoinEntry() Entry {
	bnl := databasesEntryJoin("block-nested-loop-join", "Block Nested-Loop Join", "BlockNestedLoopJoin",
		"A nested-loop join reading the outer table a block at a time, so the inner table is scanned once per block instead of once per row.", "O(n * m), with n / block inner scans",
		func(l, r databases.Table, lc, rc string, in Input, rec trace.Recorder) (databases.Table, databases.JoinStats, error) {
			return databases.BlockNestedLoopJoin(l, r, lc, rc, in.opt("block", 2), rec)
		})
	bnl.Options = []Option{{"block", "outer rows per block", 2, 1, 1000}}
	bnl.Example = &Input{Text: dbEmployees, Other: dbDepts, Key: "dept", Options: map[string]int{"block": 2}}
	return bnl
}

func externalMergeSortEntry() Entry {
	ext := entry("external-merge-sort", "External Merge Sort", databasesEntryFile("externalsort"), "ExternalSort")
	ext.Summary = "Sort more data than fits in memory: sorted runs the size of memory, then k-way merges with a heap, pass after pass, until one run is left."
	ext.Time, ext.Space, ext.Input, ext.Traced = "O(N log N), 1 + ceil(log_k(N / M)) passes", "O(N)", Ints, true
	ext.Options = []Option{{"memory", "values that fit in memory (M)", 3, 1, 10_000}, {"ways", "runs merged at once (k)", 2, 2, 64}}
	ext.Limits = "at most 10000 values"
	ext.Example = &Input{Values: []int{29, 3, 17, 41, 8, 12, 35, 1, 22, 14, 6, 30}, Options: map[string]int{"memory": 3, "ways": 2}}
	ext.Run = func(in Input, r trace.Recorder) (Output, error) {
		if err := values(in, "sort", 10_000); err != nil {
			return Output{}, err
		}
		sorted, passes := databases.ExternalSort(in.Values, in.opt("memory", 3), in.opt("ways", 2), r)
		var lines []string
		for i, p := range passes {
			runs := make([]string, len(p.Runs))
			for j, run := range p.Runs {
				runs[j] = "[" + joinInts(run) + "]"
			}
			lines = append(lines, fmt.Sprintf("pass %d: %d runs %s", i, len(p.Runs), strings.Join(runs, " ")))
		}
		return Output{Initial: append([]int(nil), in.Values...), Values: sorted, Lines: lines, Summary: map[string]float64{"passes": float64(len(passes))}}, nil
	}
	return ext
}

func queryPlannerEntry() Entry {
	plan := entry("query-planner", "Query Planner (index or full scan)", databasesEntryFile("planner"), "Plan")
	plan.Summary = "A cost-based optimizer in miniature: estimate how many rows a condition matches, price an index scan and a full scan in page reads, and pick the cheaper."
	plan.Time, plan.Space, plan.Input, plan.Params = "O(n)", "O(n)", TwoTexts, []string{"key"}
	plan.Options = []Option{{"rows-per-page", "rows per page", 2, 1, 1000}, {"fanout", "index fanout", 4, 2, 1000}}
	plan.Limits = fmt.Sprintf("at most %d characters", dbMaxText)
	plan.Example = &Input{Text: dbPeople, Other: "name = Esi", Key: "name", Options: map[string]int{"rows-per-page": 1, "fanout": 4}}
	plan.Call = func(in Input) (string, string) {
		return tableSetup(in.Text), fmt.Sprintf("databases.Plan(t, %q, %q, %d, %d)", in.Other, in.Key, in.opt("rows-per-page", 2), in.opt("fanout", 4))
	}
	plan.Run = func(in Input, _ trace.Recorder) (Output, error) {
		t, err := oneTable(in.Text)
		if err != nil {
			return Output{}, err
		}
		q, err := databases.Plan(t, in.Other, in.Key, in.opt("rows-per-page", 2), in.opt("fanout", 4))
		if err != nil {
			return Output{}, wrap(err)
		}
		lines := append(append([]string(nil), q.Explain...), "result:")
		lines = append(lines, q.Result.Lines()...)
		return Output{Lines: lines, Text: q.Access, Summary: map[string]float64{
			"estimated rows": q.EstimatedRows, "actual rows": float64(q.ActualRows), "full scan cost": q.FullScanCost, "index scan cost": q.IndexScanCost,
		}}, nil
	}
	return plan
}

func relationalAlgebraEntry() Entry {
	alg := entry("relational-algebra", "Relational Algebra (select, project, group by)", databasesEntryFile("algebra"), "Query")
	alg.Summary = "Run a pipeline of relational operators over a table: select rows, project columns, group and aggregate, sort and limit, seeing the table after each."
	alg.Time, alg.Space, alg.Input = "O(n log n) per sort, O(n) per other stage", "O(n)", TwoTexts
	alg.Limits = fmt.Sprintf("at most %d characters", dbMaxText)
	alg.Example = &Input{Text: dbPeople, Other: "select age > 26 | group city count | sort count desc"}
	alg.Call = func(in Input) (string, string) {
		return tableSetup(in.Text), fmt.Sprintf("databases.Query(t, %q)", in.Other)
	}
	alg.Run = func(in Input, _ trace.Recorder) (Output, error) {
		t, err := oneTable(in.Text)
		if err != nil {
			return Output{}, err
		}
		stages, err := databases.Query(t, in.Other)
		if err != nil {
			return Output{}, wrap(err)
		}
		var lines []string
		for _, s := range stages {
			lines = append(lines, "-- "+s.Op)
			lines = append(lines, s.Table.Lines()...)
		}
		return Output{Lines: lines, Summary: map[string]float64{"rows": float64(len(stages[len(stages)-1].Table.Rows))}}, nil
	}
	return alg
}

func databaseEntries() []Entry {
	return concat(
		databaseIndexEntries(),
		[]Entry{
			writeAheadLogEntry(),
			databasesEntryNl,
			blockNestedLoopJoinEntry(),
			databasesEntryInl,
			databasesEntryHj,
			databasesEntrySmj,
			externalMergeSortEntry(),
			queryPlannerEntry(),
			relationalAlgebraEntry(),
		},
		transactionEntries(),
	)
}

func transactionEntryScheduleSetup(text string) string {
	return fmt.Sprintf("ops, err := databases.ParseSchedule(%q)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", text)
}

var transactionEntryFile = databasesEntryFile

func transactionEntrySchedule(id, name, fname, fn, summary, time, example string) Entry {
	e := entry(id, name, transactionEntryFile(fname), fn)
	e.Summary, e.Time, e.Space, e.Input = summary, time, "O(n)", Text
	e.Limits = fmt.Sprintf("at most %d characters", dbMaxText)
	e.Example = &Input{Text: example}
	return e
}

func twoPhaseLockingEntry() Entry {
	tpl := transactionEntrySchedule("two-phase-locking", "Strict Two-Phase Locking", "twophaselocking", "TwoPhaseLocking",
		"Shared locks to read, exclusive to write, all held to the end: blocked operations wait, and a cycle in the wait-for graph is a deadlock broken by aborting the youngest.",
		"O(n^2 t) at worst", "r1(x) r2(y) w1(y) w2(x) c1 c2 r3(x) w3(x) c3")
	tpl.Call = func(in Input) (string, string) {
		return transactionEntryScheduleSetup(in.Text), "databases.TwoPhaseLocking(ops)"
	}
	tpl.Run = func(in Input, _ trace.Recorder) (Output, error) {
		ops, err := parseSchedule(in.Text)
		if err != nil {
			return Output{}, err
		}
		run := databases.TwoPhaseLocking(ops)
		var lines []string
		var spans []Span
		for i, s := range run.Steps {
			lines = append(lines, fmt.Sprintf("%-6s %s", s.Op, s.Outcome))
			spans = append(spans, Span{Lane: s.Txn, Label: s.Op, Start: i, End: i + 1, State: s.State})
			for _, w := range s.Waiting {
				if w != s.Txn {
					spans = append(spans, Span{Lane: w, Label: "waiting", Start: i, End: i + 1, State: "blocked"})
				}
			}
		}
		return Output{Lines: lines, Timeline: spans, Summary: map[string]float64{
			"committed": float64(len(run.Committed)), "aborted": float64(len(run.Aborted)), "deadlocks": float64(run.Deadlocks),
		}}, nil
	}
	return tpl
}

func conflictSerializabilityEntry() Entry {
	ser := transactionEntrySchedule("conflict-serializability", "Conflict Serializability (precedence graph)", "serializability", "Precedence",
		"Draw an edge for every pair of conflicting operations; the schedule equals some serial order exactly when the graph has no cycle.",
		"O(n^2)", "r1(x) r2(y) w2(x) r3(y) w1(y) w3(z) c1 c2 c3")
	ser.Traced = true
	ser.Run = func(in Input, r trace.Recorder) (Output, error) {
		ops, err := parseSchedule(in.Text)
		if err != nil {
			return Output{}, err
		}
		g := databases.Precedence(ops, r)
		graph := &GraphOutput{Directed: true, Nodes: g.Txns, Edges: []GraphEdge{}}
		var lines []string
		for _, e := range g.Edges {
			graph.Edges = append(graph.Edges, GraphEdge{From: e.From, To: e.To, Weight: 1})
			lines = append(lines, fmt.Sprintf("%s -> %s: %s", g.Txns[e.From], g.Txns[e.To], e.Why))
		}
		names := func(idx []int) string {
			parts := make([]string, len(idx))
			for i, k := range idx {
				parts[i] = g.Txns[k]
			}
			return strings.Join(parts, ", ")
		}
		if g.Serializable {
			lines = append(lines, "conflict-serializable, equivalent to the serial order "+names(g.Order))
		} else {
			lines = append(lines, "not conflict-serializable: the cycle "+names(g.Cycle))
		}
		ok := g.Serializable
		return Output{Graph: graph, Lines: lines, Bool: &ok}, nil
	}
	return ser
}

func timestampOrderingEntry() Entry {
	to := transactionEntrySchedule("timestamp-ordering", "Timestamp Ordering", "timestamp", "TimestampOrdering",
		"No locks: every transaction gets a timestamp, and an operation that arrives too late for timestamp order aborts its transaction, unless the Thomas write rule can skip it.",
		"O(n)", "r1(x) w2(x) r2(y) w1(x) r3(y) w3(y) c1 c2 c3")
	to.Options = []Option{{"thomas", "Thomas write rule (1 on)", 0, 0, 1}}
	to.Call = func(in Input) (string, string) {
		return transactionEntryScheduleSetup(in.Text), fmt.Sprintf("databases.TimestampOrdering(ops, %v)", in.opt("thomas", 0) == 1)
	}
	to.Run = func(in Input, _ trace.Recorder) (Output, error) {
		ops, err := parseSchedule(in.Text)
		if err != nil {
			return Output{}, err
		}
		run := databases.TimestampOrdering(ops, in.opt("thomas", 0) == 1)
		lines := make([]string, 0, len(run.Steps))
		for _, s := range run.Steps {
			lines = append(lines, fmt.Sprintf("%-6s TS %d  %s", s.Op, run.TS[s.Txn], s.Outcome))
		}
		tb := &TableOutput{Rows: len(run.Items), Cols: 2, RowLabels: run.Items, ColLabels: []string{"RTS", "WTS"}}
		for _, it := range run.Items {
			tb.Cells = append(tb.Cells, run.RTS[it], run.WTS[it])
		}
		return Output{Lines: lines, Table: tb, Summary: map[string]float64{"aborted": float64(len(run.Aborted))}}, nil
	}
	return to
}

func mvccSnapshotIsolationEntry() Entry {
	mv := transactionEntrySchedule("mvcc-snapshot-isolation", "MVCC (snapshot isolation)", "mvcc", "SnapshotIsolation",
		"Writes make new versions instead of overwriting, each transaction reads its own snapshot, and of two writers of one item the first to commit wins.",
		"O(n v) for v versions per item", "T1 begin, T2 write x 5, T2 commit, T1 read x, T3 read x, T1 write x 7, T3 write y 1, T1 commit, T3 commit, T4 read x")
	mv.Call = func(in Input) (string, string) {
		return transactionEntryScheduleSetup(in.Text), "databases.SnapshotIsolation(ops)"
	}
	mv.Run = func(in Input, _ trace.Recorder) (Output, error) {
		ops, err := parseSchedule(in.Text)
		if err != nil {
			return Output{}, err
		}
		run := databases.SnapshotIsolation(ops)
		lines := make([]string, 0, len(run.Steps))
		for _, s := range run.Steps {
			lines = append(lines, fmt.Sprintf("%-10s %s", s.Op, s.Outcome))
		}
		tb := &TableOutput{Cols: 2, ColLabels: []string{"value", "commit"}}
		for _, it := range run.Items {
			for i, v := range run.Versions[it] {
				tb.Rows++
				tb.RowLabels = append(tb.RowLabels, fmt.Sprintf("%s v%d (%s)", it, i+1, v.Txn))
				tb.Cells = append(tb.Cells, v.Value, v.Commit)
			}
		}
		out := Output{Lines: lines, Summary: map[string]float64{"committed": float64(len(run.Committed)), "aborted": float64(len(run.Aborted))}}
		if tb.Rows > 0 {
			out.Table = tb
		}
		return out, nil
	}
	return mv
}

func isolationAnomaliesEntry() Entry {
	iso := entry("isolation-anomalies", "Isolation Levels and Their Anomalies", transactionEntryFile("anomalies"), "IsolationAnomalies")
	iso.Summary = "Dirty reads, non-repeatable reads, phantoms, lost updates and write skew, each run at all four isolation levels to see which ones each level stops."
	iso.Time, iso.Space, iso.Input = "O(1)", "O(1)", Simulation
	iso.Options = []Option{{"level", "level for the story (0 read uncommitted ... 3 serializable)", 1, 0, 3}}
	iso.Example = &Input{Options: map[string]int{"level": 2}}
	iso.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("databases.IsolationAnomalies(databases.Level(%d))", in.opt("level", 1))
	}
	iso.Run = func(in Input, _ trace.Recorder) (Output, error) {
		chosen := databases.Level(in.opt("level", 1))
		levels := databases.Levels()
		if chosen < 0 || int(chosen) >= len(levels) {
			return Output{}, inputError("the level is 0 (read uncommitted) to 3 (serializable)")
		}
		tb := &TableOutput{Rows: len(databases.AnomalyNames), Cols: len(levels), RowLabels: databases.AnomalyNames}
		for _, l := range levels {
			tb.ColLabels = append(tb.ColLabels, l.String())
		}
		tb.Cells = make([]int, tb.Rows*tb.Cols)
		var lines []string
		for c, l := range levels {
			for r, a := range databases.IsolationAnomalies(l) {
				if a.Occurs {
					tb.Cells[r*tb.Cols+c] = 1
				}
				if l == chosen {
					lines = append(lines, "-- "+a.Name+" at "+l.String())
					lines = append(lines, a.Story...)
				}
			}
		}
		return Output{Table: tb, Lines: lines}, nil
	}
	return iso
}

// transactionEntries are the concurrency-control entries: schedules in,
// what each protocol did with them out.
func transactionEntries() []Entry {
	return []Entry{
		twoPhaseLockingEntry(),
		conflictSerializabilityEntry(),
		timestampOrderingEntry(),
		mvccSnapshotIsolationEntry(),
		isolationAnomaliesEntry(),
	}
}

func textLimit(s string) error {
	if strings.TrimSpace(s) == "" {
		return inputError("give the text to run")
	}
	if len(s) > dbMaxText {
		return inputError(fmt.Sprintf("at most %d characters", dbMaxText))
	}
	return nil
}

func oneTable(text string) (databases.Table, error) {
	if err := textLimit(text); err != nil {
		return databases.Table{}, err
	}
	t, err := databases.ParseTable(text)
	return t, wrap(err)
}

func joinTables(in Input) (l, r databases.Table, lcol, rcol string, err error) {
	if l, err = oneTable(in.Text); err != nil {
		return
	}
	if r, err = oneTable(in.Other); err != nil {
		return
	}
	lc, rc, ok := databases.JoinColumns(l, r, in.Key)
	if !ok {
		err = inputError("the tables share no column name; give the join condition as --key left=right")
		return
	}
	return l, r, lc, rc, nil
}

func parseSchedule(text string) ([]databases.TxOp, error) {
	if err := textLimit(text); err != nil {
		return nil, err
	}
	ops, err := databases.ParseSchedule(text)
	return ops, wrap(err)
}

func tableSetup(text string) string {
	return fmt.Sprintf("t, err := databases.ParseTable(%q)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", text)
}

func newPoolP(frames int, policy string, k int) playground { //nolint:ireturn // it picks the clock or the LRU-K pool; both are playgrounds
	b, err := databases.NewBufferPool(frames, policy, k)
	if err != nil {
		panic(err) // the policies are constants
	}
	return &poolP{b: b}
}

// databaseIndexEntries is the storage and index structures, operated on step
// by step.
func databaseIndexEntries() []Entry {
	var out []Entry
	for _, s := range databasesEntryIdx {
		e := structureEntry(s.id, s.name, s.file, s.fn, s.summary, s.time, s.space, s.fill, s.ops, s.build)
		ex := s.example // a copy: before Go 1.22 the loop variable is shared
		e.Options, e.Example = s.options, &ex
		out = append(out, e)
	}
	return out
}
