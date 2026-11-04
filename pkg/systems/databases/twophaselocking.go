package databases

import (
	"fmt"
	"slices"
	"strings"
)

// LockStep is one thing that happened under two-phase locking.
type LockStep struct {
	Txn     string
	Op      string // "R1(x)", "C2", ...
	Outcome string // "granted S", "waits for T2", "deadlock: aborted", ...
	// State is "running" (the operation ran), "blocked", "done" (committed)
	// or "aborted".
	State string
	// Waiting lists the transactions blocked after this step.
	Waiting []string
}

// LockRun is the result of TwoPhaseLocking.
type LockRun struct {
	Steps              []LockStep
	Committed, Aborted []string
	Deadlocks          int
}

type lock struct {
	exclusive bool
	holders   []string
}

// TwoPhaseLocking runs a schedule under strict two-phase locking, the
// protocol most lock-based databases use: a read takes a shared (S) lock and
// a write an exclusive (X) lock on the item, a shared lock upgrades to
// exclusive when its only holder writes, and every lock is held until the
// transaction commits or aborts. Growing then shrinking makes every schedule
// it allows conflict-serializable; holding until the end also stops others
// reading uncommitted data.
//
// An operation that cannot get its lock waits, and the transaction's later
// operations wait behind it, while the rest of the schedule goes on. Each
// time a transaction starts waiting, the wait-for graph (who waits for whom)
// is searched for a cycle; a cycle is a deadlock, broken by aborting its
// youngest transaction (the one that appeared last). A transaction whose
// operations run out without a commit commits at the end. Begin operations
// are ignored. Time: O(n^2 t) for n operations and t transactions, at
// worst.
func TwoPhaseLocking(ops []TxOp) LockRun {
	lm := newLockManager(ops)
	for len(lm.pending) > 0 {
		idx := slices.IndexFunc(lm.pending, func(o TxOp) bool { return len(lm.waitsFor[o.Txn]) == 0 })
		if idx < 0 { // everyone left waits on transactions with nothing left to do
			if !lm.commitIdle() { // cannot happen: a cycle is caught when it forms
				break
			}
			continue
		}
		lm.run1(idx)
	}
	for _, t := range lm.order {
		if !lm.finished[t] {
			lm.commit(t, "C"+strings.TrimPrefix(t, "T"), "commits at the end of its operations")
		}
	}
	return lm.run
}

// lockManager is the lock table and the transactions' progress.
type lockManager struct {
	run      LockRun
	order    []string
	age      map[string]int
	locks    map[string]*lock
	held     map[string][]string // txn -> items it locks
	waitsFor map[string][]string // txn -> the txns it waits for
	finished map[string]bool
	pending  []TxOp
}

func newLockManager(ops []TxOp) *lockManager {
	lm := &lockManager{order: txnOrder(ops), age: map[string]int{}, locks: map[string]*lock{},
		held: map[string][]string{}, waitsFor: map[string][]string{}, finished: map[string]bool{}}
	for i, t := range lm.order {
		lm.age[t] = i
	}
	for _, o := range ops {
		if o.Action != Begin {
			lm.pending = append(lm.pending, o)
		}
	}
	return lm
}

// step records what happened, with who is waiting afterward.
func (lm *lockManager) step(txn, op, outcome, state string) {
	var w []string
	for _, t := range lm.order {
		if len(lm.waitsFor[t]) > 0 {
			w = append(w, t)
		}
	}
	lm.run.Steps = append(lm.run.Steps, LockStep{Txn: txn, Op: op, Outcome: outcome, State: state, Waiting: w})
}

// run1 runs the pending operation at idx: a commit, an abort, or a lock.
func (lm *lockManager) run1(idx int) {
	o := lm.pending[idx]
	switch o.Action {
	case Commit:
		lm.commit(o.Txn, o.String(), "commits, releasing its locks")
	case Abort:
		lm.abort(o.Txn, o.String(), "aborts, releasing its locks")
	case Begin, Read, Write:
		lm.acquire(idx, o)
	}
}

// acquire grants o's lock, or makes its transaction wait and checks the
// wait-for graph for a deadlock.
func (lm *lockManager) acquire(idx int, o TxOp) {
	write := o.Action == Write
	l := lm.locks[o.Item]
	if l == nil {
		l = &lock{}
		lm.locks[o.Item] = l
	}
	if blockers := l.blockers(o.Txn, write); len(blockers) > 0 {
		lm.waitsFor[o.Txn] = blockers
		lm.step(o.Txn, o.String(), "waits for "+strings.Join(blockers, ", "), "blocked")
		lm.breakDeadlock()
		return
	}
	mode := l.grant(o.Txn, write)
	if !slices.Contains(lm.held[o.Txn], o.Item) {
		lm.held[o.Txn] = append(lm.held[o.Txn], o.Item)
	}
	lm.pending = slices.Delete(lm.pending, idx, idx+1)
	lm.step(o.Txn, o.String(), "granted "+mode, "running")
}

// blockers are the other holders a request conflicts with: any, for a write
// or an exclusive lock.
func (l *lock) blockers(txn string, write bool) []string {
	var out []string
	for _, h := range l.holders {
		if h != txn && (write || l.exclusive) {
			out = append(out, h)
		}
	}
	return out
}

// grant gives txn the lock and returns the mode it ends up with.
func (l *lock) grant(txn string, write bool) string {
	mode := "S"
	if write {
		mode = "X"
		if slices.Contains(l.holders, txn) && !l.exclusive {
			mode = "X (upgraded)"
		}
		l.exclusive = true
	}
	if !slices.Contains(l.holders, txn) {
		l.holders = append(l.holders, txn)
	}
	if l.exclusive && !write {
		mode = "X (already held)"
	}
	return mode
}

// breakDeadlock aborts the youngest transaction of a wait-for cycle, if any.
func (lm *lockManager) breakDeadlock() {
	cycle := waitCycle(lm.waitsFor, lm.order)
	if cycle == nil {
		return
	}
	victim := slices.MaxFunc(cycle, func(a, b string) int { return lm.age[a] - lm.age[b] })
	lm.run.Deadlocks++
	lm.abort(victim, "A"+strings.TrimPrefix(victim, "T"), fmt.Sprintf("deadlock %s: aborted as the youngest", strings.Join(cycle, " -> ")))
}

// commitIdle commits every unfinished transaction with no operations left,
// and reports whether there was one.
func (lm *lockManager) commitIdle() bool {
	progress := false
	for _, t := range lm.order {
		if !lm.finished[t] && !slices.ContainsFunc(lm.pending, func(o TxOp) bool { return o.Txn == t }) {
			lm.commit(t, "C"+strings.TrimPrefix(t, "T"), "commits at the end of its operations")
			progress = true
		}
	}
	return progress
}

func (lm *lockManager) commit(txn, op, why string) {
	lm.release(txn)
	lm.run.Committed = append(lm.run.Committed, txn)
	lm.step(txn, op, why, "done")
}

func (lm *lockManager) abort(txn, op, why string) {
	lm.release(txn)
	lm.run.Aborted = append(lm.run.Aborted, txn)
	lm.step(txn, op, why, "aborted")
}

// release drops txn's locks, operations and waits, and frees whoever waited
// only for it.
func (lm *lockManager) release(txn string) {
	for _, item := range lm.held[txn] {
		l := lm.locks[item]
		l.holders = slices.DeleteFunc(l.holders, func(h string) bool { return h == txn })
		if len(l.holders) == 0 {
			delete(lm.locks, item)
		}
	}
	delete(lm.held, txn)
	delete(lm.waitsFor, txn)
	lm.finished[txn] = true
	lm.pending = slices.DeleteFunc(lm.pending, func(o TxOp) bool { return o.Txn == txn })
	for t, ws := range lm.waitsFor {
		if ws = slices.DeleteFunc(ws, func(h string) bool { return h == txn }); len(ws) == 0 {
			delete(lm.waitsFor, t)
		} else {
			lm.waitsFor[t] = ws
		}
	}
}

// waitCycle returns the transactions of a cycle in the wait-for graph, or
// nil, searching from each transaction in order with a depth-first search.
func waitCycle(waitsFor map[string][]string, order []string) []string {
	state := map[string]int{} // 0 unseen, 1 on the path, 2 done
	var path []string
	var found []string
	var dfs func(t string) bool
	dfs = func(t string) bool {
		state[t] = 1
		path = append(path, t)
		for _, u := range waitsFor[t] {
			if state[u] == 1 {
				found = slices.Clone(path[slices.Index(path, u):])
				return true
			}
			if state[u] == 0 && dfs(u) {
				return true
			}
		}
		path = path[:len(path)-1]
		state[t] = 2
		return false
	}
	for _, t := range order {
		if state[t] == 0 && dfs(t) {
			return found
		}
	}
	return nil
}
