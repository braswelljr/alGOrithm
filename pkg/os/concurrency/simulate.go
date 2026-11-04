package concurrency

import (
	"errors"
	"fmt"
	"math/rand"
)

// MaxTicks bounds a simulation; MaxThreads its threads.
const (
	MaxTicks   = 20_000
	MaxThreads = 16
)

// ErrInput is returned for a simulation with too many threads or too much
// work.
var ErrInput = errors.New("concurrency: invalid simulation")

// thread is one simulated thread: step runs its next micro-step if it can,
// returning what it did and whether it is blocked (could not proceed) or done.
type thread interface {
	step() (what string, blocked, done bool)
}

// run is the scheduler: each tick it picks a random thread that is not done
// and lets it try one step. It stops when every thread is done, or reports a
// deadlock when every remaining thread is blocked for a full round.
func run(threads []thread, seed int64, shared func() int) (Timeline, bool) {
	r := rand.New(rand.NewSource(seed))
	done := make([]bool, len(threads))
	blocked := make([]bool, len(threads))
	var tl Timeline
	for tick := 0; tick < MaxTicks; tick++ {
		var live []int
		allBlocked := true
		for i := range threads {
			if !done[i] {
				live = append(live, i)
				allBlocked = allBlocked && blocked[i]
			}
		}
		if len(live) == 0 {
			return tl, false
		}
		if allBlocked {
			return tl, true // deadlock: nobody can make progress
		}
		i := live[r.Intn(len(live))]
		what, b, d := threads[i].step()
		state := "running"
		switch {
		case d:
			state, done[i] = "done", true
		case b:
			state = "blocked"
		}
		if !b {
			for j := range blocked { // progress may unblock anyone
				blocked[j] = false
			}
		}
		blocked[i] = b
		tl = append(tl, Action{Tick: tick, Thread: i, What: what, State: state, Shared: shared()})
	}
	return tl, false
}

// RaceResult is the outcome of Race.
type RaceResult struct {
	Timeline Timeline `json:"timeline"`
	Final    int      `json:"final"`    // the counter at the end
	Expected int      `json:"expected"` // threads * increments
	Lost     int      `json:"lost"`     // updates overwritten by another thread
}

// Race simulates threads each adding 1 to a shared counter increments times.
// Each increment is three steps, read, add and write, so without a lock two
// threads can read the same value and one update is lost. With locked, a
// mutex around the three steps makes every increment count.
//
// Time: O(threads * increments). Space: the same, for the timeline.
func Race(threads, increments int, locked bool, seed int64) (RaceResult, error) {
	if threads < 1 || threads > MaxThreads || increments < 1 || threads*increments*5 > MaxTicks {
		return RaceResult{}, fmt.Errorf("%w: 1 to %d threads and at most %d steps", ErrInput, MaxThreads, MaxTicks)
	}
	counter, owner := 0, -1
	ts := make([]thread, threads)
	for i := range ts {
		ts[i] = &raceThread{id: i, left: increments, locked: locked, counter: &counter, owner: &owner}
	}
	tl, _ := run(ts, seed, func() int { return counter })
	exp := threads * increments
	return RaceResult{Timeline: tl, Final: counter, Expected: exp, Lost: exp - counter}, nil
}

type raceThread struct {
	id, left, pc, local int
	locked              bool
	counter, owner      *int
}

func (t *raceThread) step() (string, bool, bool) {
	if t.left == 0 {
		return "done", false, true
	}
	switch t.pc {
	case 0: // lock
		if !t.locked {
			t.pc = 1
			return t.step()
		}
		if *t.owner >= 0 {
			return "wait lock", true, false
		}
		*t.owner, t.pc = t.id, 1
		return "lock", false, false
	case 1:
		t.local, t.pc = *t.counter, 2
		return fmt.Sprintf("read %d", t.local), false, false
	case 2:
		t.local, t.pc = t.local+1, 3
		return fmt.Sprintf("add 1 = %d", t.local), false, false
	case 3:
		*t.counter, t.pc = t.local, 4
		return fmt.Sprintf("write %d", t.local), false, false
	default: // unlock
		t.pc, t.left = 0, t.left-1
		if t.locked {
			*t.owner = -1
			return "unlock", false, t.left == 0
		}
		return "finish increment", false, t.left == 0
	}
}

// PCResult is the outcome of ProducerConsumer.
type PCResult struct {
	Timeline Timeline `json:"timeline"`
	Produced int      `json:"produced"`
	Consumed int      `json:"consumed"`
	// MaxBuffered is the fullest the buffer got; it never exceeds the capacity.
	MaxBuffered int `json:"maxBuffered"`
}

// ProducerConsumer simulates producers making items into a bounded buffer of
// capacity and consumers taking them, items in all: a producer waits while the
// buffer is full, a consumer while it is empty. Threads 0..producers-1 are the
// producers.
func ProducerConsumer(producers, consumers, capacity, items int, seed int64) (PCResult, error) {
	if producers < 1 || consumers < 1 || producers+consumers > MaxThreads || capacity < 1 || items < 1 || items*6 > MaxTicks {
		return PCResult{}, fmt.Errorf("%w: producers and consumers 1 to %d in all, capacity and items at least 1", ErrInput, MaxThreads)
	}
	var (
		buffer              []int
		made, took, maxSeen int
	)
	var ts []thread
	for p := 0; p < producers; p++ {
		ts = append(ts, &pcThread{producer: true, buffer: &buffer, capacity: capacity, items: items, made: &made, took: &took, maxSeen: &maxSeen})
	}
	for c := 0; c < consumers; c++ {
		ts = append(ts, &pcThread{buffer: &buffer, capacity: capacity, items: items, made: &made, took: &took, maxSeen: &maxSeen})
	}
	tl, _ := run(ts, seed, func() int { return len(buffer) })
	return PCResult{Timeline: tl, Produced: made, Consumed: took, MaxBuffered: maxSeen}, nil
}

type pcThread struct {
	producer            bool
	buffer              *[]int
	capacity, items     int
	made, took, maxSeen *int
	holding             int
	hasItem             bool
}

func (t *pcThread) step() (string, bool, bool) {
	if t.producer {
		switch {
		case !t.hasItem && *t.made >= t.items:
			return "done", false, true
		case !t.hasItem:
			*t.made++
			t.holding, t.hasItem = *t.made, true
			return fmt.Sprintf("produce item %d", t.holding), false, false
		case len(*t.buffer) == t.capacity:
			return "wait: buffer full", true, false
		default:
			*t.buffer = append(*t.buffer, t.holding)
			t.hasItem = false
			*t.maxSeen = max(*t.maxSeen, len(*t.buffer))
			return fmt.Sprintf("put item %d", t.holding), false, false
		}
	}
	switch {
	case t.hasItem:
		t.hasItem = false
		return fmt.Sprintf("consume item %d", t.holding), false, false
	case len(*t.buffer) == 0 && *t.took+len(*t.buffer) >= t.items:
		return "done", false, true
	case len(*t.buffer) == 0:
		return "wait: buffer empty", true, false
	default:
		t.holding = (*t.buffer)[0]
		*t.buffer = (*t.buffer)[1:]
		t.hasItem = true
		*t.took++
		return fmt.Sprintf("take item %d", t.holding), false, false
	}
}

// DPResult is the outcome of DiningPhilosophers.
type DPResult struct {
	Timeline   Timeline `json:"timeline"`
	Deadlocked bool     `json:"deadlocked"`
	Meals      []int    `json:"meals"` // meals each philosopher finished
}

// DiningPhilosophers simulates n philosophers around a table with a fork
// between each pair, each wanting meals meals; eating takes both neighboring
// forks. Naively everyone picks up the left fork first, and if all do at once
// nobody can take a right fork: deadlock (the run stops and reports it).
// With ordered, each picks up the lower-numbered fork first, which breaks the
// cycle so the table can never deadlock.
func DiningPhilosophers(n, meals int, ordered bool, seed int64) (DPResult, error) {
	if n < 2 || n > MaxThreads || meals < 1 || n*meals*6 > MaxTicks {
		return DPResult{}, fmt.Errorf("%w: 2 to %d philosophers, at least 1 meal", ErrInput, MaxThreads)
	}
	forks := make([]int, n) // owner, -1 free
	for i := range forks {
		forks[i] = -1
	}
	res := DPResult{Meals: make([]int, n)}
	ts := make([]thread, n)
	for i := range ts {
		first, second := i, (i+1)%n // left, right
		if ordered && second < first {
			first, second = second, first
		}
		ts[i] = &philosopher{id: i, first: first, second: second, forks: forks, meals: &res.Meals[i], want: meals}
	}
	inUse := func() int {
		c := 0
		for _, f := range forks {
			if f >= 0 {
				c++
			}
		}
		return c
	}
	res.Timeline, res.Deadlocked = run(ts, seed, inUse)
	return res, nil
}

type philosopher struct {
	id, first, second, pc, want int
	forks                       []int
	meals                       *int
}

func (p *philosopher) step() (string, bool, bool) {
	if *p.meals == p.want {
		return "done", false, true
	}
	switch p.pc {
	case 0:
		p.pc = 1
		return "think", false, false
	case 1:
		if p.forks[p.first] >= 0 {
			return fmt.Sprintf("wait for fork %d", p.first), true, false
		}
		p.forks[p.first], p.pc = p.id, 2
		return fmt.Sprintf("pick up fork %d", p.first), false, false
	case 2:
		if p.forks[p.second] >= 0 {
			return fmt.Sprintf("wait for fork %d", p.second), true, false
		}
		p.forks[p.second], p.pc = p.id, 3
		return fmt.Sprintf("pick up fork %d", p.second), false, false
	case 3:
		p.pc = 4
		return "eat", false, false
	default:
		p.forks[p.first], p.forks[p.second] = -1, -1
		p.pc = 0
		*p.meals++
		return "put down both forks", false, *p.meals == p.want
	}
}
