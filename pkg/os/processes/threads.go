package processes

import "fmt"

// MemoryResult is ThreadsVsProcesses: the counter each worker saw at the
// end, and what the parent sees.
type MemoryResult struct {
	Model   string   `json:"model"` // "threads" or "processes"
	Workers []int    `json:"workers"`
	Parent  int      `json:"parent"`
	Log     []string `json:"log"`
}

// ThreadsVsProcesses has workers each add increments to a counter that
// starts at 0. As threads they share one address space, so the parent sees
// workers*increments (the adds are serialized here; see the concurrency
// package's race condition for what happens when they are not). As
// processes each fork gets its own copy of memory, so every child sees only
// its own adds and the parent still sees 0: sharing data between processes
// needs IPC.
//
// Time: O(workers * increments). Space: O(workers).
func ThreadsVsProcesses(workers, increments int, threads bool) (MemoryResult, error) {
	if workers < 1 || workers > 16 || increments < 1 || increments > 1000 {
		return MemoryResult{}, inputErr(0, "1-16 workers and 1-1000 increments")
	}
	res := MemoryResult{Model: "processes", Workers: make([]int, workers)}
	shared := 0
	if threads {
		res.Model = "threads"
	}
	for w := 0; w < workers; w++ {
		own := 0
		for i := 0; i < increments; i++ {
			if threads {
				shared++
			} else {
				own++
			}
		}
		if threads {
			res.Workers[w] = shared
			res.Log = append(res.Log, fmt.Sprintf("thread %d adds %d to the shared counter: now %d", w, increments, shared))
		} else {
			res.Workers[w] = own
			res.Log = append(res.Log, fmt.Sprintf("process %d adds %d to its own copy: its counter is %d, the parent's is still 0", w, increments, own))
		}
	}
	res.Parent = shared
	return res, nil
}

// ThreadModel maps user threads onto kernel threads.
type ThreadModel string

// The three classic models.
const (
	ManyToOne  ThreadModel = "many-to-one"  // all user threads on one kernel thread
	OneToOne   ThreadModel = "one-to-one"   // each user thread its own kernel thread (Linux, Windows)
	ManyToMany ThreadModel = "many-to-many" // user threads multiplexed onto a pool (Go's goroutines)
)

// ThreadModels runs threads user threads, each needing work ticks of CPU,
// on cpus CPUs under the model. Thread 0 makes a blocking system call of
// blockFor ticks after its first tick. Under many-to-one the one kernel
// thread blocks, so every thread waits; under one-to-one only thread 0
// waits; under many-to-many (with the given kernel threads) the others keep
// the CPUs busy. The returned timeline has one lane per user thread.
//
// Time: O(total ticks * threads). Space: O(total ticks * threads).
func ThreadModels(model ThreadModel, threads, kernel, cpus, work, blockFor int) (Timeline, error) {
	if threads < 1 || threads > 8 || cpus < 1 || cpus > 8 || kernel < 1 || kernel > 8 || work < 1 || work > 20 || blockFor < 0 || blockFor > 20 {
		return nil, inputErr(0, "1-8 threads, kernel threads and CPUs; work 1-20; block 0-20")
	}
	switch model {
	case ManyToOne:
		kernel = 1
	case OneToOne:
		kernel = threads
	case ManyToMany:
	default:
		return nil, inputErr(0, "unknown model %q", model)
	}
	left := make([]int, threads)
	for i := range left {
		left[i] = work
	}
	blockedUntil, didBlock := -1, false
	var tl Timeline
	for tick := 0; tick < 500; tick++ {
		done := true
		for _, l := range left {
			if l > 0 {
				done = false
			}
		}
		if done {
			break
		}
		blocked := tick < blockedUntil
		// slots is how many threads can run this tick: one per kernel thread
		// with a CPU, less the kernel thread a blocking call holds.
		slots := min(kernel, cpus)
		if blocked && model == ManyToMany {
			slots = min(kernel-1, cpus)
		}
		running := 0
		for i := 0; i < threads; i++ {
			lane := fmt.Sprintf("thread %d", i)
			switch {
			case left[i] == 0:
				continue
			case i == 0 && blocked:
				tl = append(tl, Event{Tick: tick, Lane: lane, What: "blocked in a system call", State: "blocked"})
			case blocked && model == ManyToOne:
				tl = append(tl, Event{Tick: tick, Lane: lane, What: "stuck: the only kernel thread is blocked", State: "blocked"})
			case running < slots:
				running++
				left[i]--
				tl = append(tl, Event{Tick: tick, Lane: lane, What: "running", State: "running"})
				if i == 0 && !didBlock && blockFor > 0 {
					didBlock, blockedUntil = true, tick+1+blockFor
				}
			default:
				tl = append(tl, Event{Tick: tick, Lane: lane, What: "ready", State: "idle"})
			}
		}
	}
	return tl, nil
}
