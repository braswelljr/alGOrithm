package databases

import (
	"fmt"
	"slices"
)

// TOStep is one operation under timestamp ordering and what became of it.
type TOStep struct {
	Txn, Op string
	Outcome string
	OK      bool
}

// TORun is the result of TimestampOrdering.
type TORun struct {
	Steps []TOStep
	// TS is each transaction's timestamp; RTS and WTS each item's largest
	// read and write timestamps at the end.
	TS       map[string]int
	RTS, WTS map[string]int
	Items    []string
	Aborted  []string
}

// TimestampOrdering runs a schedule under basic timestamp ordering, a
// concurrency control with no locks and no waiting: each transaction gets a
// timestamp when it starts (here, the order it first appears), and the
// schedule must be equivalent to running the transactions in timestamp
// order. Each item remembers the largest timestamps that read it (RTS) and
// wrote it (WTS). A read by T is too late if a younger transaction already
// wrote the item (TS(T) < WTS); a write is too late if a younger one already
// read it (TS(T) < RTS) or wrote it (TS(T) < WTS). Too late means T aborts.
//
// With thomas set, the Thomas write rule (Thomas, 1979) applies: a write
// that is too late only because of a younger write is skipped instead,
// since that younger value would overwrite it anyway. Simplified: an aborted
// transaction is not restarted; its later operations are skipped. Time:
// O(n) for n operations.
func TimestampOrdering(ops []TxOp, thomas bool) TORun {
	run := TORun{TS: map[string]int{}, RTS: map[string]int{}, WTS: map[string]int{}}
	for i, t := range txnOrder(ops) {
		run.TS[t] = i + 1
	}
	aborted := map[string]bool{}
	items := map[string]bool{}
	add := func(o TxOp, outcome string, ok bool) {
		run.Steps = append(run.Steps, TOStep{Txn: o.Txn, Op: o.String(), Outcome: outcome, OK: ok})
	}
	abort := func(o TxOp, why string) {
		aborted[o.Txn] = true
		run.Aborted = append(run.Aborted, o.Txn)
		add(o, "aborts: "+why, false)
	}
	for _, o := range ops {
		if aborted[o.Txn] {
			if o.Action != Begin {
				add(o, "skipped: "+o.Txn+" aborted", false)
			}
			continue
		}
		ts := run.TS[o.Txn]
		switch o.Action {
		case Begin:
		case Commit:
			add(o, "commits", true)
		case Abort:
			abort(o, "as the schedule says")
		case Read:
			items[o.Item] = true
			if w := run.WTS[o.Item]; ts < w {
				abort(o, fmt.Sprintf("TS %d < WTS(%s) %d, a younger transaction wrote it", ts, o.Item, w))
				continue
			}
			run.RTS[o.Item] = max(run.RTS[o.Item], ts)
			add(o, fmt.Sprintf("reads; RTS(%s) = %d", o.Item, run.RTS[o.Item]), true)
		case Write:
			items[o.Item] = true
			if rt := run.RTS[o.Item]; ts < rt {
				abort(o, fmt.Sprintf("TS %d < RTS(%s) %d, a younger transaction read it", ts, o.Item, rt))
				continue
			}
			if w := run.WTS[o.Item]; ts < w {
				if thomas {
					add(o, fmt.Sprintf("skipped by the Thomas write rule: TS %d < WTS(%s) %d", ts, o.Item, w), true)
					continue
				}
				abort(o, fmt.Sprintf("TS %d < WTS(%s) %d, a younger transaction wrote it", ts, o.Item, w))
				continue
			}
			run.WTS[o.Item] = ts
			add(o, fmt.Sprintf("writes; WTS(%s) = %d", o.Item, ts), true)
		}
	}
	for k := range items {
		run.Items = append(run.Items, k)
	}
	slices.Sort(run.Items)
	return run
}
