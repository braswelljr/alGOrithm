package core

import (
	"fmt"
	"strconv"

	"github.com/braswelljr/alGOrithm/pkg/os/concurrency"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// concurrencyEntries is every entry of the concurrency category.
func concurrencyEntries() []Entry {
	return []Entry{
		barrierEntry(),
		diningPhilosophersEntry(),
		petersonEntry(),
		producerConsumerEntry(),
		raceConditionEntry(),
		readersWritersEntry(),
		semaphoreSimulationEntry(),
		sleepingBarberEntry(),
	}
}

func concurrencySim(id, name, fn, summary string, opts []Option, example map[string]int, run func(in Input) (Output, error), call func(in Input) string) Entry {
	e := entry(id, name, "os/concurrency/classics.go", fn)
	e.Summary, e.Time, e.Space, e.Input, e.Options = summary, "O(ticks)", "O(ticks)", Simulation, opts
	e.Example = &Input{Options: example}
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		out, err := run(in)
		return out, wrap(err)
	}
	e.Call = func(in Input) (string, string) { return "", call(in) }
	return e
}

var concurrencySeed = Option{"seed", "schedule seed", 1, 0, 1_000_000}

func concurrencyB(v bool) float64 {
	if v {
		return 1
	}
	return 0
}

func readersWritersEntry() Entry {
	return concurrencySim("readers-writers", "Readers-Writers", "ReadersWriters", "Readers share the data, a writer needs it alone; with reader preference writers can starve.",
		[]Option{{"readers", "readers", 4, 1, 15}, {"writers", "writers", 1, 1, 15}, {"rounds", "rounds each", 3, 1, 50}, {"writersfirst", "writer preference (1) or reader preference (0)", 0, 0, 1}, concurrencySeed},
		map[string]int{"readers": 4, "writers": 1, "rounds": 3, "writersfirst": 0, "seed": 3},
		func(in Input) (Output, error) {
			readers := in.opt("readers", 4)
			r, err := concurrency.ReadersWriters(readers, in.opt("writers", 1), in.opt("rounds", 3), in.opt("writersfirst", 0) == 1, int64(in.opt("seed", 1)))
			if err != nil {
				return Output{}, err
			}
			lane := func(i int) string {
				if i < readers {
					return "reader " + strconv.Itoa(i)
				}
				return "writer " + strconv.Itoa(i-readers)
			}
			return Output{Timeline: spans(r.Timeline, lane), Summary: map[string]float64{"longest writer wait": float64(r.MaxWriterWait), "longest reader wait": float64(r.MaxReaderWait), "most readers at once": float64(r.MaxReaders)}}, nil
		},
		func(in Input) string {
			return fmt.Sprintf("concurrency.ReadersWriters(%d, %d, %d, %v, %d)", in.opt("readers", 4), in.opt("writers", 1), in.opt("rounds", 3), in.opt("writersfirst", 0) == 1, in.opt("seed", 1))
		})
}

func sleepingBarberEntry() Entry {
	return concurrencySim("sleeping-barber", "Sleeping Barber", "SleepingBarber", "A barber sleeps until a customer arrives; customers wait in a few chairs or leave when they are full.",
		[]Option{{"customers", "customers", 6, 1, 15}, {"chairs", "waiting chairs", 2, 0, 15}, concurrencySeed},
		map[string]int{"customers": 6, "chairs": 2, "seed": 1},
		func(in Input) (Output, error) {
			r, err := concurrency.SleepingBarber(in.opt("customers", 6), in.opt("chairs", 2), int64(in.opt("seed", 1)))
			if err != nil {
				return Output{}, err
			}
			lane := func(i int) string {
				if i == 0 {
					return "barber"
				}
				return "customer " + strconv.Itoa(i)
			}
			return Output{Timeline: spans(r.Timeline, lane), Summary: map[string]float64{"served": float64(r.Served), "turned away": float64(r.TurnedAway)}}, nil
		},
		func(in Input) string {
			return fmt.Sprintf("concurrency.SleepingBarber(%d, %d, %d)", in.opt("customers", 6), in.opt("chairs", 2), in.opt("seed", 1))
		})
}

func petersonEntry() Entry {
	return concurrencySim("peterson", "Peterson's Algorithm", "Peterson", "Mutual exclusion for two threads with only shared flags and a turn; a check-then-set version lets both in.",
		[]Option{{"rounds", "rounds each", 3, 1, 100}, {"broken", "check-then-set instead (1)", 0, 0, 1}, concurrencySeed},
		map[string]int{"rounds": 3, "broken": 0, "seed": 1},
		func(in Input) (Output, error) {
			r, err := concurrency.Peterson(in.opt("rounds", 3), in.opt("broken", 0) == 1, int64(in.opt("seed", 1)))
			if err != nil {
				return Output{}, err
			}
			return Output{Timeline: spans(r.Timeline, thread), Bool: ptr(r.Violations == 0), Summary: map[string]float64{"both inside (ticks)": float64(r.Violations), "entries": float64(r.Entries)}}, nil
		},
		func(in Input) string {
			return fmt.Sprintf("concurrency.Peterson(%d, %v, %d)", in.opt("rounds", 3), in.opt("broken", 0) == 1, in.opt("seed", 1))
		})
}

func barrierEntry() Entry {
	return concurrencySim("barrier", "Barrier", "Barrier", "Threads work, then wait at a barrier until all arrive, round after round: the fast wait for the slow.",
		[]Option{{"threads", "threads", 4, 1, 16}, {"rounds", "rounds", 3, 1, 50}, concurrencySeed},
		map[string]int{"threads": 4, "rounds": 3, "seed": 1},
		func(in Input) (Output, error) {
			r, err := concurrency.Barrier(in.opt("threads", 4), in.opt("rounds", 3), int64(in.opt("seed", 1)))
			if err != nil {
				return Output{}, err
			}
			return Output{Timeline: spans(r.Timeline, thread), Summary: map[string]float64{"ticks waited": float64(r.Waited), "rounds": float64(r.Rounds)}}, nil
		},
		func(in Input) string {
			return fmt.Sprintf("concurrency.Barrier(%d, %d, %d)", in.opt("threads", 4), in.opt("rounds", 3), in.opt("seed", 1))
		})
}

func semaphoreSimulationEntry() Entry {
	return concurrencySim("semaphore-simulation", "Counting Semaphore", "SemaphoreSim", "Threads wait (P) for one of a few slots, work, and signal (V); never more inside than slots.",
		[]Option{{"threads", "threads", 5, 1, 16}, {"slots", "slots", 2, 1, 16}, {"rounds", "rounds each", 2, 1, 50}, concurrencySeed},
		map[string]int{"threads": 5, "slots": 2, "rounds": 2, "seed": 1},
		func(in Input) (Output, error) {
			r, err := concurrency.SemaphoreSim(in.opt("threads", 5), in.opt("slots", 2), in.opt("rounds", 2), int64(in.opt("seed", 1)))
			if err != nil {
				return Output{}, err
			}
			return Output{Timeline: spans(r.Timeline, thread), Bool: ptr(r.MaxInside <= in.opt("slots", 2)), Summary: map[string]float64{"most inside": float64(r.MaxInside), "ticks waited": float64(r.Waited), "ok": concurrencyB(r.MaxInside <= in.opt("slots", 2))}}, nil
		},
		func(in Input) string {
			return fmt.Sprintf("concurrency.SemaphoreSim(%d, %d, %d, %d)", in.opt("threads", 5), in.opt("slots", 2), in.opt("rounds", 2), in.opt("seed", 1))
		})
}

var concurrencyThreadName = thread

func raceConditionEntry() Entry {
	race := entry("race-condition", "Race Condition", "os/concurrency/simulate.go", "Race")
	race.Summary, race.Time, race.Space, race.Input = "Threads add to a shared counter in three steps (read, add, write); without a lock, updates are lost.", "O(threads * increments)", "O(threads * increments)", Simulation
	race.Options = []Option{{"threads", "threads", 3, 1, concurrency.MaxThreads}, {"increments", "increments each", 3, 1, 100}, {"locked", "use a lock (1 yes, 0 no)", 0, 0, 1}, {"seed", "schedule seed", 1, 0, 1_000_000}}
	race.Example = &Input{Options: map[string]int{"threads": 3, "increments": 3, "locked": 0, "seed": 2}}
	race.Run = func(in Input, _ trace.Recorder) (Output, error) {
		res, err := concurrency.Race(in.opt("threads", 3), in.opt("increments", 3), in.opt("locked", 0) == 1, int64(in.opt("seed", 1)))
		if err != nil {
			return Output{}, wrap(err)
		}
		return Output{Timeline: spans(res.Timeline, concurrencyThreadName), Number: &res.Final,
			Summary: map[string]float64{"final counter": float64(res.Final), "expected": float64(res.Expected), "lost updates": float64(res.Lost)}}, nil
	}
	return race
}

func producerConsumerEntry() Entry {
	pc := entry("producer-consumer", "Producer-Consumer", "os/concurrency/simulate.go", "ProducerConsumer")
	pc.Summary, pc.Time, pc.Space, pc.Input = "Producers fill a bounded buffer and consumers empty it; each waits when the buffer is full or empty.", "O(items)", "O(items)", Simulation
	pc.Options = []Option{{"producers", "producers", 2, 1, 8}, {"consumers", "consumers", 2, 1, 8}, {"capacity", "buffer capacity", 2, 1, 50}, {"items", "items", 10, 1, 500}, {"seed", "schedule seed", 1, 0, 1_000_000}}
	pc.Example = &Input{Options: map[string]int{"producers": 2, "consumers": 2, "capacity": 2, "items": 8, "seed": 3}}
	pc.Run = func(in Input, _ trace.Recorder) (Output, error) {
		p := in.opt("producers", 2)
		res, err := concurrency.ProducerConsumer(p, in.opt("consumers", 2), in.opt("capacity", 2), in.opt("items", 10), int64(in.opt("seed", 1)))
		if err != nil {
			return Output{}, wrap(err)
		}
		names := func(i int) string {
			if i < p {
				return "producer " + strconv.Itoa(i)
			}
			return "consumer " + strconv.Itoa(i-p)
		}
		return Output{Timeline: spans(res.Timeline, names),
			Summary: map[string]float64{"produced": float64(res.Produced), "consumed": float64(res.Consumed), "fullest buffer": float64(res.MaxBuffered)}}, nil
	}
	return pc
}

func diningPhilosophersEntry() Entry {
	dp := entry("dining-philosophers", "Dining Philosophers", "os/concurrency/simulate.go", "DiningPhilosophers")
	dp.Summary, dp.Time, dp.Space, dp.Input = "Philosophers need both neighboring forks to eat; picking up the left fork first can deadlock, ordering the forks cannot.", "O(n * meals)", "O(n * meals)", Simulation
	dp.Options = []Option{{"philosophers", "philosophers", 5, 2, concurrency.MaxThreads}, {"meals", "meals each", 2, 1, 50}, {"ordered", "lower fork first (1 yes, 0 no)", 0, 0, 1}, {"seed", "schedule seed", 1, 0, 1_000_000}}
	dp.Example = &Input{Options: map[string]int{"philosophers": 3, "meals": 3, "ordered": 0, "seed": 4}}
	dp.Run = func(in Input, _ trace.Recorder) (Output, error) {
		res, err := concurrency.DiningPhilosophers(in.opt("philosophers", 5), in.opt("meals", 2), in.opt("ordered", 0) == 1, int64(in.opt("seed", 1)))
		if err != nil {
			return Output{}, wrap(err)
		}
		names := func(i int) string { return "philosopher " + strconv.Itoa(i) }
		meals := 0
		for _, m := range res.Meals {
			meals += m
		}
		return Output{Timeline: spans(res.Timeline, names), Bool: &res.Deadlocked,
			Summary: map[string]float64{"meals eaten": float64(meals), "deadlocked": map[bool]float64{true: 1}[res.Deadlocked]}}, nil
	}
	return dp
}
