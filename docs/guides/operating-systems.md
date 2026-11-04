# Operating systems

alGOrithm simulates the operating-system concepts courses teach: how the CPU is shared between processes, how memory is managed, how a disk serves requests, how deadlock is avoided, and what goes wrong when threads share data. Each is a Go package you can call and an entry you can run from the command line or watch in the visualizer. Every example here was run against the current tree.

## CPU scheduling: `scheduling`

Processes are written one per line, `name arrival burst [priority]`. Every scheduler returns a timeline (a Gantt chart) and each process's waiting, turnaround and response times.

```text
$ algorithm run round-robin --text "P1 0 8
P2 1 4
P3 2 9
P4 3 5" --option quantum=2
Round Robin (time O(T * n) for T ticks, space O(T))
P1 |##    ##      ##    ##    |
P2 |  ##      ##              |
P3 |    ##      ##    ##   ###|
P4 |        ##      ##    #   |
```

Compare the others on the same processes: `fcfs-scheduling` (the convoy effect), `sjf-scheduling` and `srtf-scheduling` (the least waiting), `priority-scheduling` (`--option preempt=1`), and `mlfq-scheduling` (`--option quantum=2 --option levels=3 --option boost=20`).

```go
ps, _ := scheduling.Parse("P1 0 8\nP2 1 4\nP3 2 9\nP4 3 5")
s, _ := scheduling.RoundRobin(ps, 2)
s.AvgWaiting       // 12.75
s.ContextSwitches  // 12
s.Timeline[0]      // {P1 0 2 0}: P1 ran from tick 0 to 2
```

## Memory: `memory`

**Page replacement** chooses which page to evict on a page fault. The table shows each frame after every reference:

```text
$ algorithm run lru-page-replacement 7 0 1 2 0 3 0 4 --option frames=3
LRU Page Replacement (time O(n * frames), space O(n * frames))
         7 0 1 2 0 3 0 4
 frame 0 7 7 7 2 2 2 2 4
 frame 1 - 0 0 0 0 0 0 0
 frame 2 - - 1 1 1 3 3 3
...
faults      6
```

Try `fifo-page-replacement` (and Belady's anomaly: `1 2 3 4 1 2 5 1 2 3 4 5` faults 9 times with 3 frames and 10 with 4), `optimal-page-replacement` (the fewest faults possible) and `clock-page-replacement`. In Go: `memory.LRU(refs, 3)` returns every step and the fault count.

**Allocation** places processes in fixed partitions:

```text
$ algorithm run first-fit --example
0. place P1             P1 (212) went into partition 2 (500)  [100 free | 500: P1 uses 212 | 200 free | 300 free | 600 free]
...
3. place P4             P4 (426) did not fit                  [...]
internal fragmentation  559
```

`best-fit` places all four; `worst-fit` leaves P4 out again.

## Disk scheduling: `disk`

The head's path, and how far it moved, for the textbook queue (head at 53, tracks 0 to 199):

```text
$ algorithm run scan-disk --example
SCAN (Elevator) (time O(n log n), space O(n))
head path      53 65 67 98 122 124 183 199 37 14
...
head movement  331
```

`fcfs-disk` moves 640 tracks, `sstf-disk` 236, `c-scan-disk` 382, `look-disk` 299, `c-look-disk` 322. Options: `head`, `size`, and `down=1` to sweep down first. In Go: `disk.SCAN(requests, 53, 200, false)`.

## Deadlock: `deadlock`

The Banker's algorithm grants resources only if every process can still finish in some order:

```text
$ algorithm run bankers-algorithm --example
...
6.      safe order: P1, P3, P4, P0, P2
result  true
```

The state is written `available 3 3 2`, then one line per process: `P0 max 7 5 3 alloc 0 1 0`. In Go: `deadlock.Safe(available, max, allocation)` and `deadlock.Request(p, request, ...)`; `deadlock.Cycle` finds a cycle in a wait-for graph.

## Concurrency: `concurrency`

The simulations run threads as small programs under a seeded scheduler, so every interleaving can be replayed ([rule](../../docs/rules/simulations-are-deterministic.md)).

**Race condition.** Each increment of a shared counter is three steps, read, add and write; without a lock, threads overwrite each other:

```bash
algorithm run race-condition --option threads=3 --option increments=2 --option seed=2
# final counter 2, expected 6, lost updates 4
algorithm run race-condition --option threads=3 --option increments=2 --option seed=2 --option locked=1
# final counter 6, lost updates 0
```

**Producer-consumer.** Producers fill a bounded buffer and consumers empty it, each waiting when it is full or empty: `algorithm run producer-consumer --example`.

**Dining philosophers.** Picking up the left fork first can deadlock; picking the lower-numbered fork first cannot:

```bash
algorithm run dining-philosophers --option philosophers=3 --option meals=3 --option seed=4                # deadlocked
algorithm run dining-philosophers --option philosophers=3 --option meals=3 --option seed=4 --option ordered=1
```

The same in Go:

```go
r, _ := concurrency.Race(3, 2, false, 2)                // r.Final 2, r.Expected 6, r.Lost 4
d, _ := concurrency.DiningPhilosophers(3, 3, false, 4)  // d.Deadlocked true
```

The package also has the real tools these problems lead to, safe to use in programs: `WorkerPool`, `Pipe`, `Semaphore` and `BoundedBuffer`.

## Related

- The [catalog](../catalog.md), operating systems section.
- What is planned next: the [roadmap](../../docs/requirements/roadmap.md).
