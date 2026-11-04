package synchronization

// Lock is the instruction a spinlock is built on.
type Lock string

// The locks Spinlock simulates.
const (
	// TestAndSet atomically sets the lock word to 1 and returns its old value.
	TestAndSet Lock = "test-and-set"
	// CompareAndSwap atomically writes 1 only if the word is still 0.
	CompareAndSwap Lock = "compare-and-swap"
	// NonAtomic reads the word, then writes it in a separate step: two
	// threads can both read 0 and both get in.
	NonAtomic Lock = "non-atomic"
)

type spinThread struct {
	kind              Lock
	word, inside      *int
	spins             *int
	rounds, csLen, pc int
	csLeft            int
}

func (t *spinThread) step() (string, string, bool) {
	switch t.pc {
	case 0:
		switch t.kind {
		case TestAndSet:
			old := *t.word
			*t.word = 1
			if old == 1 {
				*t.spins++
				return "test-and-set: was 1, spin", "blocked", false
			}
			t.enter()
			return "test-and-set: was 0, got the lock", "running", false
		case CompareAndSwap:
			if *t.word != 0 {
				*t.spins++
				return "compare-and-swap(0, 1) failed, spin", "blocked", false
			}
			*t.word = 1
			t.enter()
			return "compare-and-swap(0, 1) succeeded", "running", false
		default:
			if *t.word != 0 {
				*t.spins++
				return "reads 1, spin", "blocked", false
			}
			t.pc = 1
			return "reads 0", "running", false
		}
	case 1: // NonAtomic's separate write
		*t.word = 1
		t.enter()
		return "writes 1, enters", "running", false
	default:
		t.csLeft--
		if t.csLeft > 0 {
			return "in the critical section", "running", false
		}
		*t.inside--
		*t.word = 0
		t.rounds--
		t.pc = 0
		return "releases the lock", "running", t.rounds == 0
	}
}

func (t *spinThread) enter() {
	*t.inside++
	t.pc, t.csLeft = 2, t.csLen
}

// Spinlock runs the given number of threads that each take a spinlock rounds times and
// hold it for csLen ticks. A waiting thread spins, retrying the instruction,
// which burns its CPU time (Counts["spins"]). With TestAndSet or
// CompareAndSwap the check and the write are one atomic instruction, so
// only one thread is ever inside; with NonAtomic they are two steps, and the
// scheduler can run another thread between them.
//
// Time: O(threads * rounds * csLen) ticks plus spinning. Space: O(threads).
func Spinlock(kind Lock, threads, rounds, csLen int, seed int64) (Result, error) {
	if kind != TestAndSet && kind != CompareAndSwap && kind != NonAtomic {
		return Result{}, inputErr("unknown lock %q", kind)
	}
	if threads < 2 || threads > 8 || rounds < 1 || rounds > 10 || csLen < 1 || csLen > 10 {
		return Result{}, inputErr("2-8 threads, 1-10 rounds, critical section 1-10 ticks")
	}
	word, inside, spins := 0, 0, 0
	ts := make([]thread, threads)
	for i := range ts {
		ts[i] = &spinThread{kind: kind, word: &word, inside: &inside, spins: &spins, rounds: rounds, csLen: csLen}
	}
	res := schedule(ts, names("thread", threads), seed, 5000, func() int { return inside }, 1)
	res.Counts = map[string]int{"spins": spins}
	return res, nil
}
