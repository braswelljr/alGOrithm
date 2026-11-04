package concurrency

import "fmt"

// The classic synchronization problems, simulated on the same seeded
// scheduler as Race: every thread is a small program of steps, a step that
// cannot proceed reports itself blocked, and the run replays exactly from its
// seed.

func checkThreads(n int) error {
	if n < 1 || n > MaxThreads {
		return fmt.Errorf("%w: 1 to %d threads", ErrInput, MaxThreads)
	}
	return nil
}

// RWResult is the outcome of ReadersWriters.
type RWResult struct {
	Timeline Timeline `json:"timeline"`
	// MaxWriterWait and MaxReaderWait are the longest any writer or reader
	// waited to get in, in ticks: under reader preference writers can starve.
	MaxWriterWait int `json:"maxWriterWait"`
	MaxReaderWait int `json:"maxReaderWait"`
	// MaxReaders is the most readers inside at once; writers are always alone.
	MaxReaders int `json:"maxReaders"`
	Violations int `json:"violations"` // ticks a writer shared the data: always 0
}

// ReadersWriters simulates readers and writers sharing data: any number of
// readers may read together, a writer needs it alone. With writersFirst, a
// waiting writer stops new readers from entering (writer preference);
// otherwise readers keep entering while any reader is inside, and a writer can
// wait a long time (reader preference, the first readers-writers problem).
// Threads 0..readers-1 are the readers; each thread enters rounds times.
func ReadersWriters(readers, writers, rounds int, writersFirst bool, seed int64) (RWResult, error) {
	if err := checkThreads(readers + writers); err != nil || readers < 1 || writers < 1 || rounds < 1 || (readers+writers)*rounds*6 > MaxTicks {
		return RWResult{}, fmt.Errorf("%w: at least one reader and one writer, %d threads in all", ErrInput, MaxThreads)
	}
	st := &rwState{writersFirst: writersFirst}
	res := RWResult{}
	ts := make([]thread, 0, readers+writers)
	for i := 0; i < readers+writers; i++ {
		ts = append(ts, &rwThread{s: st, writer: i >= readers, rounds: rounds, res: &res})
	}
	tl, _ := run(ts, seed, func() int { return st.readers + 10*st.writing })
	res.Timeline = tl
	return res, nil
}

type rwState struct {
	readers, writing, waitingWriters int
	writersFirst                     bool
	tick                             int
}

type rwThread struct {
	s                        *rwState
	writer                   bool
	rounds, pc, waited, done int
	res                      *RWResult
}

func (t *rwThread) step() (string, bool, bool) {
	s := t.s
	s.tick++
	if t.done == t.rounds {
		return "done", false, true
	}
	switch t.pc {
	case 0: // ask to enter
		t.pc, t.waited = 1, 0
		if t.writer {
			s.waitingWriters++
		}
		return "want in", false, false
	case 1: // wait until allowed
		can := s.writing == 0 && s.readers == 0
		if !t.writer {
			can = s.writing == 0 && (!s.writersFirst || s.waitingWriters == 0)
		}
		if !can {
			t.waited++
			return "wait", true, false
		}
		if t.writer {
			s.waitingWriters--
			s.writing = 1
			t.res.MaxWriterWait = max(t.res.MaxWriterWait, t.waited)
		} else {
			s.readers++
			t.res.MaxReaders = max(t.res.MaxReaders, s.readers)
			t.res.MaxReaderWait = max(t.res.MaxReaderWait, t.waited)
		}
		if s.writing > 0 && s.readers > 0 {
			t.res.Violations++
		}
		t.pc = 2
		if t.writer {
			return "start writing", false, false
		}
		return "start reading", false, false
	case 2:
		t.pc = 3
		if t.writer {
			return "write", false, false
		}
		return "read", false, false
	default: // leave
		if t.writer {
			s.writing = 0
		} else {
			s.readers--
		}
		t.pc, t.done = 0, t.done+1
		return "leave", false, t.done == t.rounds
	}
}

// BarberResult is the outcome of SleepingBarber.
type BarberResult struct {
	Timeline   Timeline `json:"timeline"`
	Served     int      `json:"served"`
	TurnedAway int      `json:"turnedAway"` // customers who found every chair taken
}

// SleepingBarber simulates Dijkstra's sleeping barber: a barber who sleeps
// when nobody waits, a waiting room of chairs, and customers who arrive, take
// a chair if one is free (waking the barber) and leave if none is. Thread 0 is
// the barber; the others are customers, each arriving once after a few ticks
// that depend on the seed.
func SleepingBarber(customers, chairs int, seed int64) (BarberResult, error) {
	if err := checkThreads(customers + 1); err != nil || customers < 1 || chairs < 0 {
		return BarberResult{}, fmt.Errorf("%w: 1 to %d customers and chairs >= 0", ErrInput, MaxThreads-1)
	}
	s := &shop{chairs: chairs, customers: customers}
	res := BarberResult{}
	ts := []thread{&barber{s: s, res: &res}}
	for i := 0; i < customers; i++ {
		ts = append(ts, &customer{s: s, id: i + 1, delay: int(seed+int64(i)*7)%5 + i, res: &res})
	}
	tl, _ := run(ts, seed, func() int { return len(s.waiting) })
	res.Timeline = tl
	return res, nil
}

type shop struct {
	chairs, customers, finished int
	waiting                     []int
	cutting                     int // the customer in the chair, or 0
	doneCut                     map[int]bool
}

type barber struct {
	s   *shop
	pc  int
	res *BarberResult
}

func (b *barber) step() (string, bool, bool) {
	s := b.s
	if s.doneCut == nil {
		s.doneCut = map[int]bool{}
	}
	switch b.pc {
	case 0:
		if len(s.waiting) == 0 {
			if s.finished == s.customers {
				return "close the shop", false, true
			}
			return "sleep", true, false
		}
		s.cutting, s.waiting = s.waiting[0], s.waiting[1:]
		b.pc = 1
		return fmt.Sprintf("call customer %d", s.cutting), false, false
	case 1:
		b.pc = 2
		return fmt.Sprintf("cut customer %d's hair", s.cutting), false, false
	default:
		s.doneCut[s.cutting] = true
		b.res.Served++
		b.pc = 0
		return fmt.Sprintf("finish customer %d", s.cutting), false, false
	}
}

type customer struct {
	s         *shop
	id, delay int
	pc        int
	res       *BarberResult
}

func (c *customer) step() (string, bool, bool) {
	s := c.s
	switch c.pc {
	case 0:
		if c.delay > 0 {
			c.delay--
			return "on the way", false, false
		}
		// The waiting room holds chairs customers; an idle barber takes one more
		// straight to his chair.
		busy := s.cutting != 0 && !s.doneCut[s.cutting]
		room := s.chairs
		if !busy {
			room++
		}
		if len(s.waiting) >= room {
			s.finished++
			c.res.TurnedAway++
			return "no free chair: leave", false, true
		}
		s.waiting = append(s.waiting, c.id)
		c.pc = 1
		return "sit and wait", false, false
	default:
		if !s.doneCut[c.id] {
			return "wait for the barber", true, false
		}
		s.finished++
		return "leave with a haircut", false, true
	}
}

// PetersonResult is the outcome of Peterson.
type PetersonResult struct {
	Timeline Timeline `json:"timeline"`
	// Violations counts the ticks both threads were in the critical section.
	Violations int `json:"violations"`
	Entries    int `json:"entries"`
}

// Peterson simulates Peterson's algorithm (1981), mutual exclusion for two
// threads with nothing but shared memory: each raises its flag, gives the
// turn to the other, and waits while the other's flag is up and it is the
// other's turn. Mutual exclusion holds on every schedule. With broken, the
// threads instead check the other's flag and then raise their own, two
// separate steps, and on some schedules both get in.
func Peterson(rounds int, broken bool, seed int64) (PetersonResult, error) {
	if rounds < 1 || rounds*12 > MaxTicks {
		return PetersonResult{}, fmt.Errorf("%w: at least one round", ErrInput)
	}
	s := &peterson{}
	res := PetersonResult{}
	ts := []thread{&petersonThread{s: s, me: 0, rounds: rounds, broken: broken, res: &res}, &petersonThread{s: s, me: 1, rounds: rounds, broken: broken, res: &res}}
	tl, _ := run(ts, seed, func() int { return s.inside })
	res.Timeline = tl
	return res, nil
}

type peterson struct {
	flag   [2]bool
	turn   int
	inside int
}

type petersonThread struct {
	s                *peterson
	me, rounds, done int
	pc               int
	broken           bool
	res              *PetersonResult
}

func (t *petersonThread) step() (string, bool, bool) {
	s, other := t.s, 1-t.me
	if t.done == t.rounds {
		return "done", false, true
	}
	if t.broken {
		switch t.pc {
		case 0: // check first...
			if s.flag[other] {
				return "wait: the other's flag is up", true, false
			}
			t.pc = 1
			return "see the other's flag down", false, false
		case 1: // ...then raise the flag: too late if the other did the same
			s.flag[t.me], t.pc = true, 2
			return "raise my flag", false, false
		}
	} else {
		switch t.pc {
		case 0:
			s.flag[t.me], t.pc = true, 1
			return "raise my flag", false, false
		case 1:
			s.turn, t.pc = other, 2
			return fmt.Sprintf("give the turn to %d", other), false, false
		}
	}
	switch t.pc {
	case 2:
		if !t.broken && s.flag[other] && s.turn == other {
			return "wait: the other wants in and has the turn", true, false
		}
		s.inside++
		t.res.Entries++
		if s.inside > 1 {
			t.res.Violations++
		}
		t.pc = 3
		return "enter the critical section", false, false
	case 3:
		if s.inside > 1 {
			t.res.Violations++
		}
		t.pc = 4
		return "work in the critical section", false, false
	default:
		s.inside--
		s.flag[t.me], t.pc = false, 0
		t.done++
		return "leave, lower my flag", false, t.done == t.rounds
	}
}

// BarrierResult is the outcome of Barrier.
type BarrierResult struct {
	Timeline Timeline `json:"timeline"`
	// Waited is the ticks threads spent waiting at the barrier, in all.
	Waited int `json:"waited"`
	Rounds int `json:"rounds"`
}

// Barrier simulates threads that work, then wait at a barrier until every
// thread has arrived, round after round: the fast threads wait for the slow
// one. Thread i works i%3 + 1 ticks a round.
func Barrier(threads, rounds int, seed int64) (BarrierResult, error) {
	if err := checkThreads(threads); err != nil || rounds < 1 || threads*rounds*6 > MaxTicks {
		return BarrierResult{}, fmt.Errorf("%w: 1 to %d threads, at least one round", ErrInput, MaxThreads)
	}
	b := &barrier{n: threads, arrived: map[int]int{}}
	res := BarrierResult{Rounds: rounds}
	ts := make([]thread, threads)
	for i := range ts {
		ts[i] = &barrierThread{b: b, work: i%3 + 1, rounds: rounds, res: &res}
	}
	tl, _ := run(ts, seed, func() int { return b.arrived[b.round] })
	res.Timeline = tl
	return res, nil
}

type barrier struct {
	n, round int
	arrived  map[int]int // round -> threads arrived
}

type barrierThread struct {
	b                         *barrier
	work, left, round, rounds int
	waiting                   bool
	res                       *BarrierResult
}

func (t *barrierThread) step() (string, bool, bool) {
	b := t.b
	if t.round == t.rounds {
		return "done", false, true
	}
	if t.waiting {
		if b.arrived[t.round] < b.n {
			t.res.Waited++
			return fmt.Sprintf("wait at barrier %d", t.round+1), true, false
		}
		t.waiting, t.left = false, 0
		t.round++
		return fmt.Sprintf("pass barrier %d", t.round), false, t.round == t.rounds
	}
	if t.left < t.work {
		t.left++
		return fmt.Sprintf("work (round %d)", t.round+1), false, false
	}
	b.arrived[t.round]++
	b.round = t.round
	t.waiting = true
	return fmt.Sprintf("arrive at barrier %d", t.round+1), false, false
}

// SemaphoreResult is the outcome of SemaphoreSim.
type SemaphoreResult struct {
	Timeline  Timeline `json:"timeline"`
	MaxInside int      `json:"maxInside"` // never more than the slots
	Waited    int      `json:"waited"`
}

// SemaphoreSim simulates threads sharing a counting semaphore of slots: each
// waits (P) until a slot is free, works two ticks, and signals (V) to free it,
// rounds times. At most slots threads are ever inside.
func SemaphoreSim(threads, slots, rounds int, seed int64) (SemaphoreResult, error) {
	if err := checkThreads(threads); err != nil || slots < 1 || rounds < 1 || threads*rounds*5 > MaxTicks {
		return SemaphoreResult{}, fmt.Errorf("%w: 1 to %d threads, at least one slot and round", ErrInput, MaxThreads)
	}
	sem := &semaphore{free: slots}
	res := SemaphoreResult{}
	ts := make([]thread, threads)
	for i := range ts {
		ts[i] = &semThread{s: sem, rounds: rounds, res: &res}
	}
	tl, _ := run(ts, seed, func() int { return slots - sem.free })
	res.Timeline = tl
	return res, nil
}

type semaphore struct{ free, inside int }

type semThread struct {
	s                *semaphore
	pc, rounds, done int
	res              *SemaphoreResult
}

func (t *semThread) step() (string, bool, bool) {
	s := t.s
	if t.done == t.rounds {
		return "done", false, true
	}
	switch t.pc {
	case 0:
		if s.free == 0 {
			t.res.Waited++
			return "wait (P): no free slot", true, false
		}
		s.free--
		s.inside++
		t.res.MaxInside = max(t.res.MaxInside, s.inside)
		t.pc = 1
		return "acquire a slot (P)", false, false
	case 1, 2:
		t.pc++
		return "work", false, false
	default:
		s.free++
		s.inside--
		t.pc, t.done = 0, t.done+1
		return "release the slot (V)", false, t.done == t.rounds
	}
}
