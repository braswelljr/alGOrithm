# Simulations are deterministic

The concurrency simulations (`concurrency.Race`, `ProducerConsumer`, `DiningPhilosophers`) run their threads under a scheduler driven by a seed, so the same arguments always give the same interleaving, the same timeline and the same outcome. The real concurrency tools in the same package are tested under the race detector instead.

## Why

A real race appears in one run and not the next, which makes it impossible to teach from: you cannot show "this interleaving lost an update" twice. Simulating threads as small step functions, with a seeded scheduler picking who runs each tick, turns every interleaving into one you can replay, step through in the visualizer, and share by its seed. The same is true of the other simulations (CPU scheduling, page replacement, disk scheduling), which are deterministic by construction.

The real tools (`WorkerPool`, `Pipe`, `Semaphore`, `BoundedBuffer`) are real goroutines and are nondeterministic, as Go is; they are tested for what must hold in every interleaving, under `-race`.

## What it looks like

From [`os/concurrency/simulate.go`](../../pkg/os/concurrency/simulate.go):

```go
type thread interface {
	step() (what string, blocked, done bool)
}

func run(threads []thread, seed int64, shared func() int) (Timeline, bool) {
	r := rand.New(rand.NewSource(seed))
	...
```

On the command line the seed is an option:

```bash
algorithm run race-condition --option threads=3 --option increments=2 --option seed=2
```

Do: give every simulation a seed parameter and draw all its randomness from it.

Do not: use real goroutines, `time`, or the global random source inside a simulation.

## How it is checked

In [`os/concurrency/concurrency_test.go`](../../pkg/os/concurrency/concurrency_test.go):

- `TestRace` runs each seed twice and requires the same result and timeline; it also requires some seed to lose updates without the lock and none to lose any with it.
- `TestDiningPhilosophers` requires the naive table to deadlock for some seed and the ordered table never to.
- `TestWorkerPool`, `TestPipe` and `TestSemaphoreAndBuffer` test the real tools under `go test -race`.

## Related

- [A test fails when the answer is wrong](a-test-fails-when-the-answer-is-wrong.md)
- The [operating systems guide](../guides/operating-systems.md).
