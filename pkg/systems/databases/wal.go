package databases

import (
	"fmt"
	"slices"
	"strings"
)

// LogRecord is one record of a write-ahead log.
type LogRecord struct {
	LSN  int    // its position in the log, from 1
	Kind string // "begin", "update", "clr" (an undo, written while aborting), "commit", "abort"
	Txn  string
	Item string
	Old  int
	New  int
}

// String writes the record the way a textbook lists a log.
func (r LogRecord) String() string {
	switch r.Kind {
	case "update":
		return fmt.Sprintf("%d: %s update %s %d -> %d", r.LSN, r.Txn, r.Item, r.Old, r.New)
	case "clr":
		return fmt.Sprintf("%d: %s undo %s back to %d", r.LSN, r.Txn, r.Item, r.New)
	}
	return fmt.Sprintf("%d: %s %s", r.LSN, r.Txn, r.Kind)
}

// Recovery is what a crash and the recovery after it did.
type Recovery struct {
	Log []LogRecord
	// Disk is the items on disk at the crash; Memory the items in the
	// buffer, lost in the crash; Recovered the items after recovery.
	Disk, Memory, Recovered map[string]int
	Winners, Losers         []string // committed or aborted; unfinished
	Redone, Undone          int
	Steps                   []string
}

// RecoverWAL runs a script against a database with a write-ahead log, crashes
// it, and recovers, the way ARIES (Mohan and others, 1992) does in three
// passes: analysis finds the transactions that never finished (the losers),
// redo repeats history by replaying every update from the log onto the disk,
// and undo rolls the losers back, newest change first.
//
// The script is one operation per comma-separated part: "T1 begin",
// "T1 write x 5", "T1 commit", "T1 abort", "flush x" (write x's page from the
// buffer to disk) and "crash" (stop; without one, the crash comes after the
// last operation). Every item starts at 0. Writes change only the buffer, so
// a crash loses everything not flushed, which redo restores from the log;
// flushed changes of unfinished transactions are on disk, which undo
// removes. An abort undoes its changes at once, logging each undo (a CLR) so
// redo replays it too.
//
// Simplified: no checkpoints, pages are single items, and the script must
// lock as strict two-phase locking would (no write to an item another
// unfinished transaction wrote). Time: O(n) for n operations.
func RecoverWAL(script string) (Recovery, error) {
	w := &walRun{rec: Recovery{Disk: map[string]int{}, Memory: map[string]int{}}, active: map[string]bool{}, finished: map[string]bool{}}
	for _, part := range strings.FieldsFunc(script, func(r rune) bool { return r == ',' || r == ';' || r == '\n' }) {
		crashed, err := w.do(part)
		if err != nil {
			return Recovery{}, err
		}
		if crashed {
			break
		}
	}
	w.rec.Steps = append(w.rec.Steps, "crash: the buffer is lost")
	w.analysis()
	w.rec.Recovered = w.undo(w.redo())
	return w.rec, nil
}

// walRun is a database writing ahead to its log until it crashes: the log,
// the buffer (Memory), the disk, and which transactions are running or done.
type walRun struct {
	rec              Recovery
	active, finished map[string]bool
}

func (w *walRun) log(r LogRecord) {
	r.LSN = len(w.rec.Log) + 1
	w.rec.Log = append(w.rec.Log, r)
}

// value is an item as a transaction sees it: in the buffer, else on disk.
func (w *walRun) value(item string) int {
	if v, ok := w.rec.Memory[item]; ok {
		return v
	}
	return w.rec.Disk[item]
}

// do runs one part of the script and reports whether it was the crash.
func (w *walRun) do(part string) (crashed bool, err error) {
	f := strings.Fields(part)
	if len(f) == 0 {
		return false, nil
	}
	switch strings.ToLower(f[0]) {
	case "crash":
		return true, nil
	case "flush":
		if len(f) != 2 {
			return false, fmt.Errorf("%w: write flush ITEM, not %q", ErrInput, part)
		}
		if v, ok := w.rec.Memory[f[1]]; ok {
			w.rec.Disk[f[1]] = v
			w.rec.Steps = append(w.rec.Steps, fmt.Sprintf("flush %s = %d to disk", f[1], v))
		}
		return false, nil
	}
	if !isTxn(f[0]) {
		return false, fmt.Errorf("%w: cannot read %q (write T1 begin, T1 write x 5, T1 commit, T1 abort, flush x or crash)", ErrInput, part)
	}
	op, err := parseLong(f)
	if err != nil {
		return false, err
	}
	if w.finished[op.Txn] {
		return false, fmt.Errorf("%w: %s has already finished", ErrInput, op.Txn)
	}
	w.apply(op)
	return false, nil
}

// apply runs a transaction's operation, logging it first; a transaction's
// first operation logs its begin.
func (w *walRun) apply(op TxOp) {
	if !w.active[op.Txn] {
		w.active[op.Txn] = true
		w.log(LogRecord{Kind: "begin", Txn: op.Txn})
	}
	switch op.Action {
	case Begin:
	case Write:
		w.log(LogRecord{Kind: "update", Txn: op.Txn, Item: op.Item, Old: w.value(op.Item), New: op.Value})
		w.rec.Memory[op.Item] = op.Value
	case Read:
		w.rec.Steps = append(w.rec.Steps, fmt.Sprintf("%s reads %s = %d", op.Txn, op.Item, w.value(op.Item)))
	case Commit:
		w.log(LogRecord{Kind: "commit", Txn: op.Txn})
		delete(w.active, op.Txn)
		w.finished[op.Txn] = true
	case Abort:
		for i := len(w.rec.Log) - 1; i >= 0; i-- {
			if l := w.rec.Log[i]; l.Txn == op.Txn && l.Kind == "update" {
				w.log(LogRecord{Kind: "clr", Txn: op.Txn, Item: l.Item, New: l.Old})
				w.rec.Memory[l.Item] = l.Old
			}
		}
		w.log(LogRecord{Kind: "abort", Txn: op.Txn})
		delete(w.active, op.Txn)
		w.finished[op.Txn] = true
	}
}

// analysis finds the winners (finished before the crash) and the losers.
func (w *walRun) analysis() {
	done := map[string]bool{}
	for _, l := range w.rec.Log {
		if l.Kind == "commit" || l.Kind == "abort" {
			done[l.Txn] = true
		}
	}
	for _, l := range w.rec.Log {
		switch {
		case l.Kind != "begin":
		case done[l.Txn]:
			w.rec.Winners = append(w.rec.Winners, l.Txn)
		default:
			w.rec.Losers = append(w.rec.Losers, l.Txn)
		}
	}
	w.rec.Steps = append(w.rec.Steps, fmt.Sprintf("analysis: finished %v, losers %v", w.rec.Winners, w.rec.Losers))
}

// redo repeats history onto a copy of the disk, every update and
// compensation in log order.
func (w *walRun) redo() map[string]int {
	state := map[string]int{}
	for k, v := range w.rec.Disk {
		state[k] = v
	}
	for _, l := range w.rec.Log {
		if l.Kind == "update" || l.Kind == "clr" {
			state[l.Item] = l.New
			w.rec.Redone++
			w.rec.Steps = append(w.rec.Steps, fmt.Sprintf("redo %d: %s = %d", l.LSN, l.Item, l.New))
		}
	}
	return state
}

// undo rolls the losers' updates back, newest first.
func (w *walRun) undo(state map[string]int) map[string]int {
	for i := len(w.rec.Log) - 1; i >= 0; i-- {
		l := w.rec.Log[i]
		if l.Kind == "update" && slices.Contains(w.rec.Losers, l.Txn) {
			state[l.Item] = l.Old
			w.rec.Undone++
			w.rec.Steps = append(w.rec.Steps, fmt.Sprintf("undo %d: %s = %d (%s never committed)", l.LSN, l.Item, l.Old, l.Txn))
		}
	}
	return state
}

// Items lists every item the recovery touched, sorted.
func (r Recovery) Items() []string {
	set := map[string]bool{}
	for _, m := range []map[string]int{r.Disk, r.Memory, r.Recovered} {
		for k := range m {
			set[k] = true
		}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}
