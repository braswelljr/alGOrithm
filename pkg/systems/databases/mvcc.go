package databases

import (
	"fmt"
	"slices"
)

// Version is one committed value of an item under MVCC.
type Version struct {
	Value  int
	Txn    string
	Commit int // the commit timestamp; versions of an item are in its order
}

// MVCCStep is one operation under snapshot isolation and what it did.
type MVCCStep struct {
	Txn, Op, Outcome string
}

// MVCCRun is the result of SnapshotIsolation.
type MVCCRun struct {
	Steps              []MVCCStep
	Versions           map[string][]Version
	Items              []string
	Committed, Aborted []string
}

type mvccTxn struct {
	snapshot int // the last commit timestamp it can see
	writes   map[string]int
	order    []string
	done     bool
}

// SnapshotIsolation runs a schedule under multiversion concurrency control
// with snapshot isolation, as PostgreSQL, Oracle and many others do. A write
// never overwrites: committing installs a new version of each item written,
// stamped with the commit's timestamp. A transaction reads the snapshot of
// the database as of when it began: for each item, the newest version
// committed before then (or its own write), so readers never wait for writers
// nor writers for readers. Two transactions that write the same item cannot
// both commit: the first committer wins, and the second aborts when it tries
// to commit.
//
// A transaction begins at its "begin" or at its first operation. Items start
// at 0 and writes carry their value ("T1 write x 5"). Snapshot isolation
// still allows write skew, which the isolation-anomalies entry shows.
// Simplified: no garbage collection of old versions. Time: O(n v) for n
// operations and v versions per item.
func SnapshotIsolation(ops []TxOp) MVCCRun {
	run := MVCCRun{Versions: map[string][]Version{}}
	clock := 0
	txns := map[string]*mvccTxn{}
	add := func(o TxOp, outcome string) {
		op := o.String()
		if o.Action == Write {
			op = fmt.Sprintf("W%s(%s=%d)", o.Txn[1:], o.Item, o.Value)
		}
		run.Steps = append(run.Steps, MVCCStep{Txn: o.Txn, Op: op, Outcome: outcome})
	}
	for _, o := range ops {
		t, ok := txns[o.Txn]
		if !ok {
			t = &mvccTxn{snapshot: clock, writes: map[string]int{}}
			txns[o.Txn] = t
			if o.Action != Begin {
				add(TxOp{Txn: o.Txn, Action: Begin}, fmt.Sprintf("begins; sees commits up to %d", clock))
			}
		}
		if t.done {
			add(o, "skipped: "+o.Txn+" has finished")
			continue
		}
		switch o.Action {
		case Begin:
			if ok {
				add(o, "already begun")
			} else {
				add(o, fmt.Sprintf("begins; sees commits up to %d", clock))
			}
		case Read:
			if v, own := t.writes[o.Item]; own {
				add(o, fmt.Sprintf("reads %s = %d, its own write", o.Item, v))
				continue
			}
			v, by := 0, "the initial value"
			for _, ver := range run.Versions[o.Item] {
				if ver.Commit <= t.snapshot {
					v, by = ver.Value, fmt.Sprintf("the version %s committed at %d", ver.Txn, ver.Commit)
				}
			}
			add(o, fmt.Sprintf("reads %s = %d, %s", o.Item, v, by))
		case Write:
			if _, again := t.writes[o.Item]; !again {
				t.order = append(t.order, o.Item)
			}
			t.writes[o.Item] = o.Value
			add(o, "writes a private version, seen by no one else until it commits")
		case Abort:
			t.done = true
			run.Aborted = append(run.Aborted, o.Txn)
			add(o, "aborts; its versions are thrown away")
		case Commit:
			t.done = true
			conflict := ""
			for _, item := range t.order {
				vs := run.Versions[item]
				if len(vs) > 0 && vs[len(vs)-1].Commit > t.snapshot {
					conflict = fmt.Sprintf("%s committed %s at %d, after %s began", vs[len(vs)-1].Txn, item, vs[len(vs)-1].Commit, o.Txn)
					break
				}
			}
			if conflict != "" {
				run.Aborted = append(run.Aborted, o.Txn)
				add(o, "aborts: write-write conflict, the first committer wins ("+conflict+")")
				continue
			}
			clock++
			for _, item := range t.order {
				run.Versions[item] = append(run.Versions[item], Version{Value: t.writes[item], Txn: o.Txn, Commit: clock})
			}
			run.Committed = append(run.Committed, o.Txn)
			add(o, fmt.Sprintf("commits at %d", clock))
		}
	}
	for k := range run.Versions {
		run.Items = append(run.Items, k)
	}
	slices.Sort(run.Items)
	return run
}
