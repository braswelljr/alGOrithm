// Package concurrency teaches processes, threads and synchronization in two
// halves.
//
// The tools are real Go concurrency, safe to use: WorkerPool runs jobs on a
// fixed number of goroutines, Pipe chains stages over channels, Semaphore
// limits how many goroutines enter a section, and BoundedBuffer is a
// producer-consumer queue built on sync.Cond.
//
// The simulations show what goes wrong and why the fixes work, step by step:
// threads are small programs, and a seeded scheduler picks which one runs each
// tick, so every interleaving can be replayed exactly (real goroutines would
// interleave differently each run). Race shows lost updates on a shared
// counter and the lock that prevents them; ProducerConsumer shows a bounded
// buffer's waiting; DiningPhilosophers shows a deadlock and the lock ordering
// that avoids it. Each returns a Timeline: what every thread did at every
// tick.
package concurrency

// Action is what one thread did at one tick of a simulation.
type Action struct {
	Tick   int    `json:"tick"`
	Thread int    `json:"thread"`
	What   string `json:"what"`  // "read 3", "write 4", "lock", "wait lock", "eat", ...
	State  string `json:"state"` // "running", "blocked", "done"
	// Shared is the value of the shared state after the tick (the counter,
	// the buffer length), for drawing it.
	Shared int `json:"shared"`
}

// Timeline is every action of a simulation, in tick order.
type Timeline []Action
