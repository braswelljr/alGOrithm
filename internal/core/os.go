package core

import (
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/os/boot"
	"github.com/braswelljr/alGOrithm/pkg/os/concurrency"
	"github.com/braswelljr/alGOrithm/pkg/os/deadlock"
	"github.com/braswelljr/alGOrithm/pkg/os/disk"
	"github.com/braswelljr/alGOrithm/pkg/os/filesystems"
	"github.com/braswelljr/alGOrithm/pkg/os/ipc"
	"github.com/braswelljr/alGOrithm/pkg/os/memory"
	"github.com/braswelljr/alGOrithm/pkg/os/processes"
	"github.com/braswelljr/alGOrithm/pkg/os/scheduling"
	"github.com/braswelljr/alGOrithm/pkg/os/synchronization"
	"github.com/braswelljr/alGOrithm/pkg/os/virtualmemory"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

const processesExample = "name arrival burst priority\nP1 0 8 3\nP2 1 4 1\nP3 2 9 4\nP4 3 5 2\nP5 6 2 5"

// scheduleOutput turns a Schedule into a timeline, a statistics table and a
// summary.
func scheduleOutput(s scheduling.Schedule) Output {
	out := Output{Summary: map[string]float64{
		"average waiting": s.AvgWaiting, "average turnaround": s.AvgTurnaround, "average response": s.AvgResponse,
		"context switches": float64(s.ContextSwitches), "makespan": float64(s.Makespan),
	}}
	for _, sl := range s.Timeline {
		lane, state, label := sl.Process, "running", sl.Process
		if lane == "" {
			lane, state, label = "idle", "idle", "idle"
		}
		if sl.Queue > 0 {
			label += " (queue " + strconv.Itoa(sl.Queue) + ")"
		}
		out.Timeline = append(out.Timeline, Span{Lane: lane, Label: label, Start: sl.Start, End: sl.End, State: state})
	}
	cols := []string{"arrival", "burst", "priority", "start", "finish", "turnaround", "waiting", "response"}
	t := &TableOutput{Rows: len(s.Stats), Cols: len(cols), ColLabels: cols}
	for _, st := range s.Stats {
		t.RowLabels = append(t.RowLabels, st.Name)
		t.Cells = append(t.Cells, st.Arrival, st.Burst, st.Priority, st.Start, st.Finish, st.Turnaround, st.Waiting, st.Response)
	}
	out.Table = t
	return out
}

func schedulingEntries() []Entry {
	mk := func(id, name, fn, summary string, opts []Option, run func(ps []scheduling.Process, in Input) (scheduling.Schedule, error)) Entry {
		e := entry(id, name, "os/scheduling/algorithms.go", fn)
		e.Summary, e.Time, e.Space, e.Input, e.Options = summary, "O(T * n) for T ticks", "O(T)", Processes, opts
		e.Example = &Input{Text: processesExample, Options: map[string]int{}}
		e.Limits = fmt.Sprintf("at most %d processes and %d ticks of work", scheduling.MaxProcesses, scheduling.MaxTime)
		e.Run = func(in Input, _ trace.Recorder) (Output, error) {
			ps, err := scheduling.Parse(in.Text)
			if err != nil {
				return Output{}, wrap(err)
			}
			s, err := run(ps, in)
			if err != nil {
				return Output{}, wrap(err)
			}
			return scheduleOutput(s), nil
		}
		return e
	}
	return []Entry{
		mk("fcfs-scheduling", "First Come, First Served", "FCFS", "Run processes to completion in order of arrival; one long job makes everyone wait (the convoy effect).", nil,
			func(ps []scheduling.Process, _ Input) (scheduling.Schedule, error) { return scheduling.FCFS(ps) }),
		mk("sjf-scheduling", "Shortest Job First", "SJF", "When the CPU frees up, run the shortest ready job to completion; least average waiting without preemption.", nil,
			func(ps []scheduling.Process, _ Input) (scheduling.Schedule, error) { return scheduling.SJF(ps) }),
		mk("srtf-scheduling", "Shortest Remaining Time First", "SRTF", "Preemptive SJF: a new arrival with less work left takes the CPU; least average waiting of all.", nil,
			func(ps []scheduling.Process, _ Input) (scheduling.Schedule, error) { return scheduling.SRTF(ps) }),
		mk("round-robin", "Round Robin", "RoundRobin", "Each ready process runs for one time quantum in turn; good response, nothing starves.",
			[]Option{{"quantum", "time quantum", 2, 1, 100}},
			func(ps []scheduling.Process, in Input) (scheduling.Schedule, error) {
				return scheduling.RoundRobin(ps, in.opt("quantum", 2))
			}),
		mk("priority-scheduling", "Priority Scheduling", "Priority", "Run the most urgent ready process (lowest number); with preemption a more urgent arrival takes over.",
			[]Option{{"preempt", "preemptive (1 yes, 0 no)", 0, 0, 1}},
			func(ps []scheduling.Process, in Input) (scheduling.Schedule, error) {
				return scheduling.Priority(ps, in.opt("preempt", 0) == 1)
			}),
		mk("mlfq-scheduling", "Multilevel Feedback Queue", "MLFQ", "Round-robin queues by urgency; using a whole quantum moves a process down, so short jobs finish fast.",
			[]Option{{"quantum", "top quantum (doubles each level)", 2, 1, 50}, {"levels", "levels", 3, 1, 6}, {"boost", "priority boost every N ticks (0 off)", 0, 0, 10000}},
			func(ps []scheduling.Process, in Input) (scheduling.Schedule, error) {
				q := make([]int, in.opt("levels", 3))
				for i := range q {
					q[i] = in.opt("quantum", 2) << i
				}
				return scheduling.MLFQ(ps, q, in.opt("boost", 0))
			}),
	}
}

func memoryEntries() []Entry {
	refs := []int{7, 0, 1, 2, 0, 3, 0, 4, 2, 3, 0, 3, 2, 1, 2, 0, 1, 7, 0, 1}
	frames := []Option{{"frames", "frames", 3, 1, memory.MaxFrames}}
	mk := func(id, name, fn, summary string, run func([]int, int) (memory.Result, error)) Entry {
		e := entry(id, name, "os/memory/memory.go", fn)
		e.Summary, e.Time, e.Space, e.Input, e.Options, e.Traced = summary, "O(n * frames)", "O(n * frames)", Pages, frames, true
		e.Example = &Input{Values: refs, Options: map[string]int{"frames": 3}}
		e.Run = func(in Input, r trace.Recorder) (Output, error) {
			res, err := run(in.Values, in.opt("frames", 3))
			if err != nil {
				return Output{}, wrap(err)
			}
			f := in.opt("frames", 3)
			t := &TableOutput{Rows: f, Cols: len(res.Steps), Cells: make([]int, f*len(res.Steps)), Blank: ptr(-1)}
			for c, st := range res.Steps {
				t.ColLabels = append(t.ColLabels, strconv.Itoa(st.Page))
				for row, p := range st.Frames {
					t.Cells[row*len(res.Steps)+c] = p
					r.Record(trace.Event{Kind: trace.Set, Indices: []int{row*len(res.Steps) + c}, Values: []int{p}})
				}
				note := "hit"
				if st.Fault {
					note = "fault"
				}
				r.Record(trace.Event{Kind: trace.Mark, Indices: []int{st.Frame*len(res.Steps) + c}, Note: note})
			}
			t.RowLabels = numberLabels(f, 0)
			for i := range t.RowLabels {
				t.RowLabels[i] = "frame " + t.RowLabels[i]
			}
			return Output{Table: t, Number: &res.Faults, Summary: map[string]float64{
				"faults": float64(res.Faults), "hits": float64(res.Hits),
				"fault rate": float64(res.Faults) / float64(len(res.Steps)),
			}}, nil
		}
		return e
	}
	fit := func(id, name, fn, summary string, run func(parts, procs []int) (memory.Allocation, error)) Entry {
		e := entry(id, name, "os/memory/allocation.go", fn)
		e.Summary, e.Time, e.Space, e.Input = summary, "O(n * m)", "O(n * m)", Partitions
		e.Example = &Input{Values: []int{100, 500, 200, 300, 600}, Text: "212 417 112 426"}
		e.Run = func(in Input, _ trace.Recorder) (Output, error) {
			var procs []int
			for _, f := range strings.Fields(strings.ReplaceAll(in.Text, ",", " ")) {
				n, err := strconv.Atoi(f)
				if err != nil {
					return Output{}, inputError(fmt.Sprintf("process size %q is not a whole number", f))
				}
				procs = append(procs, n)
			}
			a, err := run(in.Values, procs)
			if err != nil {
				return Output{}, wrap(err)
			}
			out := Output{Summary: map[string]float64{"internal fragmentation": float64(a.Internal), "not placed": float64(a.Unplaced)}}
			for i, parts := range a.Steps {
				s := trace.Snapshot{Layout: "row", Nodes: make([]trace.Node, len(parts))}
				for j, p := range parts {
					n := trace.Node{ID: j, Parent: -1, Label: fmt.Sprintf("%d free", p.Size), Role: "free"}
					if p.Process >= 0 {
						n.Label, n.Role = fmt.Sprintf("%d: P%d uses %d", p.Size, p.Process+1, p.Used), ""
						if p.Process == i {
							n.Role = "new"
						}
					}
					s.Nodes[j] = n
				}
				res := fmt.Sprintf("P%d (%d) did not fit", i+1, procs[i])
				if a.Placed[i] >= 0 {
					res = fmt.Sprintf("P%d (%d) went into partition %d (%d)", i+1, procs[i], a.Placed[i]+1, in.Values[a.Placed[i]])
				}
				out.Steps = append(out.Steps, Step{Op: Op{Name: "place", Value: "P" + strconv.Itoa(i+1)}, Result: res, Snapshot: s})
			}
			return out, nil
		}
		return e
	}
	return []Entry{
		mk("fifo-page-replacement", "FIFO Page Replacement", "FIFO", "Evict the page loaded longest ago; adding frames can add faults (Belady's anomaly).", memory.FIFO),
		mk("lru-page-replacement", "LRU Page Replacement", "LRU", "Evict the page used longest ago: the recent past predicts the near future.", memory.LRU),
		mk("optimal-page-replacement", "Optimal Page Replacement", "Optimal", "Evict the page used furthest in the future: impossible to run for real, the fewest faults possible.", memory.Optimal),
		mk("clock-page-replacement", "Clock (Second Chance)", "Clock", "A hand sweeps the frames clearing reference bits, evicting the first page not used since the last sweep.", memory.Clock),
		fit("first-fit", "First Fit", "FirstFit", "Place each process in the first free partition big enough.", memory.FirstFit),
		fit("best-fit", "Best Fit", "BestFit", "Place each process in the smallest free partition big enough.", memory.BestFit),
		fit("worst-fit", "Worst Fit", "WorstFit", "Place each process in the largest free partition, so the leftover stays useful.", memory.WorstFit),
	}
}

func diskEntries() []Entry {
	opts := []Option{{"head", "head starts at", 53, 0, 100_000}, {"size", "tracks", 200, 1, 100_000}, {"down", "moving down (1 yes, 0 no)", 0, 0, 1}}
	mk := func(id, name, fn, summary, time string, run func(reqs []int, head, size int, down bool) (disk.Result, error)) Entry {
		e := entry(id, name, "os/disk/disk.go", fn)
		e.Summary, e.Time, e.Space, e.Input, e.Options, e.Traced = summary, time, "O(n)", Tracks, opts, true
		e.Example = &Input{Values: []int{98, 183, 37, 122, 14, 124, 65, 67}, Options: map[string]int{"head": 53, "size": 200}}
		e.Run = func(in Input, r trace.Recorder) (Output, error) {
			size := in.opt("size", 200)
			res, err := run(in.Values, in.opt("head", 53), size, in.opt("down", 0) == 1)
			if err != nil {
				return Output{}, wrap(err)
			}
			for _, t := range res.Path {
				r.Record(trace.Event{Kind: trace.Visit, Indices: []int{t}})
			}
			return Output{Track: &TrackOutput{Size: size, Path: res.Path}, Values: res.Order, Number: &res.Movement,
				Summary: map[string]float64{"head movement": float64(res.Movement), "average seek": float64(res.Movement) / float64(len(res.Order))}}, nil
		}
		return e
	}
	ignoreDir := func(f func([]int, int, int) (disk.Result, error)) func([]int, int, int, bool) (disk.Result, error) {
		return func(r []int, h, s int, _ bool) (disk.Result, error) { return f(r, h, s) }
	}
	return []Entry{
		mk("fcfs-disk", "FCFS Disk Scheduling", "FCFS", "Serve requests in arrival order: fair, but the head swings back and forth.", "O(n)", ignoreDir(disk.FCFS)),
		mk("sstf-disk", "Shortest Seek Time First", "SSTF", "Serve the nearest request next: less movement, but far requests can starve.", "O(n^2)", ignoreDir(disk.SSTF)),
		mk("scan-disk", "SCAN (Elevator)", "SCAN", "Sweep to one end serving requests, then reverse, like an elevator.", "O(n log n)", disk.SCAN),
		mk("c-scan-disk", "C-SCAN", "CSCAN", "Sweep one way to the end, jump back to the start, sweep again: more even waits.", "O(n log n)", disk.CSCAN),
		mk("look-disk", "LOOK", "LOOK", "SCAN that turns at the last request instead of the disk's end.", "O(n log n)", disk.LOOK),
		mk("c-look-disk", "C-LOOK", "CLOOK", "C-SCAN that jumps from the last request to the first instead of between the ends.", "O(n log n)", disk.CLOOK),
	}
}

const bankerExample = `available 3 3 2
P0 max 7 5 3 alloc 0 1 0
P1 max 3 2 2 alloc 2 0 0
P2 max 9 0 2 alloc 3 0 2
P3 max 2 2 2 alloc 2 1 1
P4 max 4 3 3 alloc 0 0 2`

func deadlockEntries() []Entry {
	b := entry("bankers-algorithm", "Banker's Algorithm", "os/deadlock/deadlock.go", "Safe")
	b.Summary, b.Time, b.Space, b.Input = "Find an order in which every process can finish with what is free, or show the state is unsafe (Dijkstra, 1965).", "O(n^2 * m)", "O(n * m)", Banker
	b.Example = &Input{Text: bankerExample}
	b.Run = func(in Input, _ trace.Recorder) (Output, error) {
		st, err := deadlock.ParseState(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		safe, order, steps, err := deadlock.Safe(st.Available, st.Max, st.Allocation)
		if err != nil {
			return Output{}, wrap(err)
		}
		out := Output{Bool: &safe}
		m := len(st.Available)
		t := &TableOutput{Rows: len(st.Max), Cols: 3 * m}
		for r := 0; r < m; r++ {
			t.ColLabels = append(t.ColLabels, fmt.Sprintf("max %c", 'A'+r))
		}
		for r := 0; r < m; r++ {
			t.ColLabels = append(t.ColLabels, fmt.Sprintf("alloc %c", 'A'+r))
		}
		for r := 0; r < m; r++ {
			t.ColLabels = append(t.ColLabels, fmt.Sprintf("need %c", 'A'+r))
		}
		for p := range st.Max {
			t.RowLabels = append(t.RowLabels, st.Names[p])
			t.Cells = append(t.Cells, st.Max[p]...)
			t.Cells = append(t.Cells, st.Allocation[p]...)
			for r := 0; r < m; r++ {
				t.Cells = append(t.Cells, st.Max[p][r]-st.Allocation[p][r])
			}
		}
		out.Table = t
		for _, s := range steps {
			out.Lines = append(out.Lines, fmt.Sprintf("%s can finish (needs %v); it releases its resources: available %v", st.Names[s.Process], s.Need, s.Available))
		}
		if safe {
			names := make([]string, len(order))
			for i, p := range order {
				names[i] = st.Names[p]
			}
			out.Lines = append(out.Lines, "safe order: "+strings.Join(names, ", "))
		} else {
			out.Lines = append(out.Lines, "unsafe: no remaining process can finish with what is available")
		}
		return out, nil
	}
	return []Entry{b}
}

func concurrencyEntries() []Entry {
	threadName := thread

	race := entry("race-condition", "Race Condition", "os/concurrency/simulate.go", "Race")
	race.Summary, race.Time, race.Space, race.Input = "Threads add to a shared counter in three steps (read, add, write); without a lock, updates are lost.", "O(threads * increments)", "O(threads * increments)", Simulation
	race.Options = []Option{{"threads", "threads", 3, 1, concurrency.MaxThreads}, {"increments", "increments each", 3, 1, 100}, {"locked", "use a lock (1 yes, 0 no)", 0, 0, 1}, {"seed", "schedule seed", 1, 0, 1_000_000}}
	race.Example = &Input{Options: map[string]int{"threads": 3, "increments": 3, "locked": 0, "seed": 2}}
	race.Run = func(in Input, _ trace.Recorder) (Output, error) {
		res, err := concurrency.Race(in.opt("threads", 3), in.opt("increments", 3), in.opt("locked", 0) == 1, int64(in.opt("seed", 1)))
		if err != nil {
			return Output{}, wrap(err)
		}
		return Output{Timeline: spans(res.Timeline, threadName), Number: &res.Final,
			Summary: map[string]float64{"final counter": float64(res.Final), "expected": float64(res.Expected), "lost updates": float64(res.Lost)}}, nil
	}

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
	return []Entry{race, pc, dp}
}

// sim describes one simulation entry: the Entry fields, how to run it, and
// the Go call that reproduces a run (its arguments, in order).
type sim struct {
	id, name, file, fn  string
	summary, time, room string
	input               InputKind
	example             Input
	options             []Option
	run                 func(in Input) (Output, error)
	args                func(in Input) []any
}

// entry builds the Entry.
func (s sim) entry() Entry {
	e := entry(s.id, s.name, s.file, s.fn)
	e.Summary, e.Time, e.Space, e.Input = s.summary, s.time, s.room, s.input
	ex := s.example
	e.Example = &ex
	e.Options = s.options
	run := s.run
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		out, err := run(in)
		return out, wrap(err)
	}
	if s.args != nil {
		pkg, fn, args := path.Base(path.Dir(s.file)), s.fn, s.args
		e.Call = func(in Input) (string, string) { return "", goCallOf(pkg, fn, args(in)...) }
	}
	return e
}

// goCallOf writes pkg.fn(args...) as Go source.
func goCallOf(pkg, fn string, args ...any) string {
	parts := make([]string, len(args))
	for i, a := range args {
		switch v := a.(type) {
		case string:
			parts[i] = strconv.Quote(v)
		case []int:
			parts[i] = ints(v)
		default:
			parts[i] = fmt.Sprint(v)
		}
	}
	return pkg + "." + fn + "(" + strings.Join(parts, ", ") + ")"
}

// tick turns one-tick events into timeline bars.
func tick(lane, what, state string, at int) Span {
	return Span{Lane: lane, Label: what, Start: at, End: at + 1, State: state}
}

// summary turns counts into a summary.
func summary(counts map[string]int) map[string]float64 {
	out := map[string]float64{}
	for k, v := range counts {
		out[k] = float64(v)
	}
	return out
}

func on(in Input, name string, def int) bool { return in.opt(name, def) == 1 }

func osMoreEntries() []Entry {
	var out []Entry
	for _, group := range [][]sim{processSims(), syncSims(), ipcSims(), fsSims(), vmSims(), bootSims()} {
		for _, s := range group {
			out = append(out, s.entry())
		}
	}
	return out
}

func processSims() []sim {
	stateOf := map[processes.State]string{processes.Ready: "idle", processes.Running: "running", processes.Waiting: "blocked", processes.Terminated: "done"}
	return []sim{
		{
			id: "process-lifecycle", name: "Process Lifecycle (five states)", file: "os/processes/lifecycle.go", fn: "Lifecycle",
			summary: "Drive a process through new, ready, running, waiting and terminated, and see which events each state allows.",
			time:    "O(n)", room: "O(n)", input: Text,
			example: Input{Text: "admit dispatch timeout dispatch wait wake dispatch exit"},
			run: func(in Input) (Output, error) {
				tr, err := processes.Lifecycle(in.Text)
				if err != nil {
					return Output{}, err
				}
				var out Output
				for _, t := range tr {
					out.Lines = append(out.Lines, fmt.Sprintf("%d. %s: %s -> %s, because %s", t.Step, t.Event, t.From, t.To, t.Why))
					out.Timeline = append(out.Timeline, tick("process", string(t.To), stateOf[t.To], t.Step-1))
				}
				return out, nil
			},
			args: func(in Input) []any { return []any{in.Text} },
		},
		{
			id: "fork-exec-wait", name: "fork, exec, wait and exit", file: "os/processes/forktree.go", fn: "ForkTree",
			summary: "Build a process tree with fork and exec, and watch exit leave zombies until a parent waits, and orphans adopted by init.",
			time:    "O(n * p)", room: "O(p)", input: Text,
			example: Input{Text: "2 fork\n3 exec ls\n2 fork\n4 fork\n3 exit 0\n4 exit 2\n2 wait\n5 exit 0"},
			run: func(in Input) (Output, error) {
				r, err := processes.ForkTree(in.Text)
				if err != nil {
					return Output{}, err
				}
				out := Output{Lines: r.Log, Summary: map[string]float64{}}
				t := &TableOutput{Cols: 2, ColLabels: []string{"parent", "exit status"}}
				for _, p := range r.Procs {
					t.Rows++
					t.Cells = append(t.Cells, p.Parent, p.Status)
					t.RowLabels = append(t.RowLabels, fmt.Sprintf("%d %s (%s)", p.PID, p.Program, p.State))
					out.Summary[p.State]++
				}
				out.Table = t
				return out, nil
			},
			args: func(in Input) []any { return []any{in.Text} },
		},
		{
			id: "system-calls", name: "System Calls (user and kernel mode)", file: "os/processes/syscalls.go", fn: "SystemCalls",
			summary: "Run a program's work and system calls, and see every trap into kernel mode and the time it costs.",
			time:    "O(ticks)", room: "O(ticks)", input: Text,
			example: Input{Text: "compute 3\nopen notes.txt\nread 3 100\ncompute 2\nwrite 1 hello\nclose 3\nread 3 10"},
			run: func(in Input) (Output, error) {
				r, err := processes.SystemCalls(in.Text)
				if err != nil {
					return Output{}, err
				}
				out := Output{Lines: r.Log, Summary: map[string]float64{"user ticks": float64(r.UserTicks), "kernel ticks": float64(r.KernelTicks), "mode switches": float64(r.Switches)}}
				for _, e := range r.Timeline {
					out.Timeline = append(out.Timeline, tick(e.Lane, e.What, e.State, e.Tick))
				}
				return out, nil
			},
			args: func(in Input) []any { return []any{in.Text} },
		},
		{
			id: "signals", name: "Signals and Handlers", file: "os/processes/signals.go", fn: "Signals",
			summary: "Install handlers, block and ignore signals, and see why SIGKILL always wins.",
			time:    "O(n)", room: "O(signals)", input: Text,
			example: Input{Text: "handle SIGINT\nsend SIGINT\nblock SIGTERM\nsend SIGTERM\nsend SIGTERM\nignore SIGKILL\nsend SIGUSR1"},
			run: func(in Input) (Output, error) {
				r, err := processes.Signals(in.Text)
				if err != nil {
					return Output{}, err
				}
				lines := r.Log
				if len(r.Pending) > 0 {
					lines = append(lines, "still pending: "+strings.Join(r.Pending, ", "))
				}
				return Output{Lines: lines, Bool: ptr(r.Alive), Summary: map[string]float64{"pending": float64(len(r.Pending))}}, nil
			},
			args: func(in Input) []any { return []any{in.Text} },
		},
		{
			id: "threads-vs-processes", name: "Threads Share, Processes Copy", file: "os/processes/threads.go", fn: "ThreadsVsProcesses",
			summary: "Workers add to one counter: as threads they share it, as forked processes each changes its own copy.",
			time:    "O(workers * increments)", room: "O(workers)", input: Simulation,
			example: Input{Options: map[string]int{"workers": 3, "increments": 5, "threads": 1}},
			options: []Option{{"workers", "workers", 3, 1, 16}, {"increments", "adds per worker", 5, 1, 1000}, {"threads", "threads (1) or processes (0)", 1, 0, 1}},
			run: func(in Input) (Output, error) {
				r, err := processes.ThreadsVsProcesses(in.opt("workers", 3), in.opt("increments", 5), on(in, "threads", 1))
				if err != nil {
					return Output{}, err
				}
				return Output{Lines: append(r.Log, fmt.Sprintf("the parent sees %d", r.Parent)), Values: r.Workers, Number: ptr(r.Parent)}, nil
			},
			args: func(in Input) []any {
				return []any{in.opt("workers", 3), in.opt("increments", 5), on(in, "threads", 1)}
			},
		},
		{
			id: "thread-models", name: "Thread Models (many-to-one, one-to-one, many-to-many)", file: "os/processes/threads.go", fn: "ThreadModels",
			summary: "Map user threads onto kernel threads three ways, and see what one blocking system call does to the others.",
			time:    "O(ticks * threads)", room: "O(ticks * threads)", input: Simulation,
			example: Input{Options: map[string]int{"model": 0, "threads": 3, "kernel": 2, "cpus": 2, "work": 3, "block": 4}},
			options: []Option{{"model", "0 many-to-one, 1 one-to-one, 2 many-to-many", 0, 0, 2}, {"threads", "user threads", 3, 1, 8}, {"kernel", "kernel threads (many-to-many)", 2, 1, 8}, {"cpus", "CPUs", 2, 1, 8}, {"work", "ticks of work each", 3, 1, 20}, {"block", "ticks thread 0 blocks", 4, 0, 20}},
			run: func(in Input) (Output, error) {
				model := []processes.ThreadModel{processes.ManyToOne, processes.OneToOne, processes.ManyToMany}[in.opt("model", 0)]
				tl, err := processes.ThreadModels(model, in.opt("threads", 3), in.opt("kernel", 2), in.opt("cpus", 2), in.opt("work", 3), in.opt("block", 4))
				if err != nil {
					return Output{}, err
				}
				var out Output
				end := 0
				for _, e := range tl {
					out.Timeline = append(out.Timeline, tick(e.Lane, e.What, e.State, e.Tick))
					end = max(end, e.Tick+1)
				}
				out.Number = ptr(end)
				out.Lines = []string{fmt.Sprintf("%s: every thread is done after %d ticks", model, end)}
				return out, nil
			},
			args: func(in Input) []any {
				return []any{[]string{"many-to-one", "one-to-one", "many-to-many"}[in.opt("model", 0)], in.opt("threads", 3), in.opt("kernel", 2), in.opt("cpus", 2), in.opt("work", 3), in.opt("block", 4)}
			},
		},
	}
}

// syncOutput turns a synchronization result into an Output.
func syncOutput(r synchronization.Result, err error) (Output, error) {
	if err != nil {
		return Output{}, err
	}
	out := Output{Lines: r.Log, Summary: summary(r.Counts)}
	out.Summary["most inside at once"] = float64(r.MaxInside)
	out.Summary["ticks"] = float64(r.Ticks)
	for _, e := range r.Timeline {
		out.Timeline = append(out.Timeline, tick(e.Lane, e.What, e.State, e.Tick))
	}
	ok := !r.Violated && !r.Deadlocked
	out.Bool = &ok
	switch {
	case r.Violated:
		out.Lines = append(out.Lines, fmt.Sprintf("VIOLATED: %d were inside at once", r.MaxInside))
	case r.Deadlocked:
		out.Lines = append(out.Lines, "DEADLOCK: every thread left is waiting for another")
	default:
		out.Lines = append(out.Lines, fmt.Sprintf("safe: at most %d inside at once, done in %d ticks", r.MaxInside, r.Ticks))
	}
	return out, nil
}

func syncSims() []sim {
	threads := func(def int) Option { return Option{"threads", "threads", def, 2, 8} }
	locks := []synchronization.Lock{synchronization.NonAtomic, synchronization.TestAndSet}
	return []sim{
		{
			id: "peterson-model-check", name: "Peterson's Algorithm, Model Checked", file: "os/synchronization/peterson.go", fn: "Check",
			summary: "Check every interleaving of two threads: Peterson's algorithm is safe and live; the two simpler attempts are not, and here is how they fail.",
			time:    "O(states)", room: "O(states)", input: Simulation,
			example: Input{Options: map[string]int{"variant": 0}},
			options: []Option{{"variant", "0 Peterson, 1 check then set, 2 flags first", 0, 0, 2}},
			run: func(in Input) (Output, error) {
				v := []synchronization.Variant{synchronization.Peterson, synchronization.CheckThenSet, synchronization.FlagsFirst}[in.opt("variant", 0)]
				r, err := synchronization.Check(v)
				if err != nil {
					return Output{}, err
				}
				lines := []string{r.String()}
				if len(r.Counterexample) > 0 {
					lines = append(lines, "counterexample:")
					for i, s := range r.Counterexample {
						lines = append(lines, fmt.Sprintf("  %d. %s", i+1, s))
					}
				}
				ok := r.MutualExcl && r.DeadlockFree
				return Output{Lines: lines, Bool: &ok, Summary: map[string]float64{"states": float64(r.States)}}, nil
			},
			args: func(in Input) []any {
				return []any{[]string{"peterson", "check-then-set", "flags-first"}[in.opt("variant", 0)]}
			},
		},
		{
			id: "lamport-bakery", name: "Lamport's Bakery Algorithm", file: "os/synchronization/bakery.go", fn: "Bakery",
			summary: "Mutual exclusion for any number of threads with only reads and writes: take a ticket, wait for every smaller one.",
			time:    "O(threads^2 * rounds)", room: "O(threads)", input: Simulation,
			example: Input{Options: map[string]int{"threads": 3, "rounds": 2, "cs": 2, "seed": 1}},
			options: []Option{{"threads", "threads", 3, 2, 6}, {"rounds", "rounds each", 2, 1, 5}, {"cs", "critical section ticks", 2, 1, 5}, seedOption},
			run: func(in Input) (Output, error) {
				return syncOutput(synchronization.Bakery(in.opt("threads", 3), in.opt("rounds", 2), in.opt("cs", 2), int64(in.opt("seed", 1))))
			},
			args: func(in Input) []any {
				return []any{in.opt("threads", 3), in.opt("rounds", 2), in.opt("cs", 2), in.opt("seed", 1)}
			},
		},
		{
			id: "test-and-set-lock", name: "Test-and-Set Spinlock", file: "os/synchronization/spinlock.go", fn: "Spinlock",
			summary: "A spinlock on one atomic instruction, and the same lock split into a read and a write, which lets two threads in.",
			time:    "O(threads * rounds * cs)", room: "O(threads)", input: Simulation,
			example: Input{Options: map[string]int{"threads": 3, "rounds": 2, "cs": 2, "atomic": 1, "seed": 1}},
			options: []Option{threads(3), {"rounds", "rounds each", 2, 1, 10}, {"cs", "critical section ticks", 2, 1, 10}, {"atomic", "atomic test-and-set (1) or a separate read and write (0)", 1, 0, 1}, seedOption},
			run: func(in Input) (Output, error) {
				return syncOutput(synchronization.Spinlock(locks[in.opt("atomic", 1)], in.opt("threads", 3), in.opt("rounds", 2), in.opt("cs", 2), int64(in.opt("seed", 1))))
			},
			args: func(in Input) []any {
				return []any{string(locks[in.opt("atomic", 1)]), in.opt("threads", 3), in.opt("rounds", 2), in.opt("cs", 2), in.opt("seed", 1)}
			},
		},
		{
			id: "compare-and-swap-lock", name: "Compare-and-Swap Spinlock", file: "os/synchronization/spinlock.go", fn: "Spinlock",
			summary: "A spinlock on compare-and-swap, the instruction lock-free code is built on, counting the spins a waiter burns.",
			time:    "O(threads * rounds * cs)", room: "O(threads)", input: Simulation,
			example: Input{Options: map[string]int{"threads": 3, "rounds": 2, "cs": 3, "seed": 2}},
			options: []Option{threads(3), {"rounds", "rounds each", 2, 1, 10}, {"cs", "critical section ticks", 3, 1, 10}, seedOption},
			run: func(in Input) (Output, error) {
				return syncOutput(synchronization.Spinlock(synchronization.CompareAndSwap, in.opt("threads", 3), in.opt("rounds", 2), in.opt("cs", 3), int64(in.opt("seed", 2))))
			},
			args: func(in Input) []any {
				return []any{"compare-and-swap", in.opt("threads", 3), in.opt("rounds", 2), in.opt("cs", 3), in.opt("seed", 2)}
			},
		},
		{
			id: "condition-variable", name: "Condition Variables: while, not if", file: "os/synchronization/condvar.go", fn: "ConditionVariable",
			summary: "Consumers wait on a condition variable for items; waiting in an if lets a woken consumer take from an empty buffer.",
			time:    "O(items * consumers)", room: "O(consumers)", input: Simulation,
			example: Input{Options: map[string]int{"loop": 1, "consumers": 3, "items": 6, "seed": 3}},
			options: []Option{{"loop", "wait in a while loop (1) or an if (0)", 1, 0, 1}, {"consumers", "consumers", 3, 2, 6}, {"items", "items (a multiple of consumers)", 6, 1, 20}, seedOption},
			run: func(in Input) (Output, error) {
				return syncOutput(synchronization.ConditionVariable(on(in, "loop", 1), in.opt("consumers", 3), in.opt("items", 6), int64(in.opt("seed", 3))))
			},
			args: func(in Input) []any {
				return []any{on(in, "loop", 1), in.opt("consumers", 3), in.opt("items", 6), in.opt("seed", 3)}
			},
		},
		{
			id: "cigarette-smokers", name: "Cigarette Smokers", file: "os/synchronization/smokers.go", fn: "CigaretteSmokers",
			summary: "An agent puts out two of three ingredients; grabbing them one at a time deadlocks, a pusher that sees the pair does not.",
			time:    "O(rounds)", room: "O(rounds)", input: Simulation,
			example: Input{Options: map[string]int{"rounds": 6, "naive": 0, "seed": 1}},
			options: []Option{{"rounds", "rounds", 6, 1, 30}, {"naive", "grab one at a time (1) or use a pusher (0)", 0, 0, 1}, seedOption},
			run: func(in Input) (Output, error) {
				return syncOutput(synchronization.CigaretteSmokers(in.opt("rounds", 6), on(in, "naive", 0), int64(in.opt("seed", 1))))
			},
			args: func(in Input) []any { return []any{in.opt("rounds", 6), on(in, "naive", 0), in.opt("seed", 1)} },
		},
		{
			id: "priority-inversion", name: "Priority Inversion and Inheritance", file: "os/synchronization/inversion.go", fn: "PriorityInversion",
			summary: "A low-priority task holds a lock the high one needs while a medium one hogs the CPU: the Mars Pathfinder bug, and its fix.",
			time:    "O(ticks)", room: "O(ticks)", input: Simulation,
			example: Input{Options: map[string]int{"medium": 6, "inherit": 0}},
			options: []Option{{"medium", "ticks the medium task computes", 6, 1, 20}, {"inherit", "priority inheritance (1) or not (0)", 0, 0, 1}},
			run: func(in Input) (Output, error) {
				return syncOutput(synchronization.PriorityInversion(in.opt("medium", 6), on(in, "inherit", 0)))
			},
			args: func(in Input) []any { return []any{in.opt("medium", 6), on(in, "inherit", 0)} },
		},
		{
			id: "aba-problem", name: "The ABA Problem", file: "os/synchronization/aba.go", fn: "ABA",
			summary: "A lock-free stack's compare-and-swap succeeds on a pointer that changed and changed back; a tag counter catches it.",
			time:    "O(1)", room: "O(1)", input: Simulation,
			example: Input{Options: map[string]int{"tagged": 0}},
			options: []Option{{"tagged", "tagged pointers (1) or plain (0)", 0, 0, 1}},
			run: func(in Input) (Output, error) {
				r := synchronization.ABA(on(in, "tagged", 0))
				return Output{Lines: r.Log, Text: r.Final, Bool: ptr(!r.Corrupted)}, nil
			},
			args: func(in Input) []any { return []any{on(in, "tagged", 0)} },
		},
	}
}

// ipcOutput turns an ipc result into an Output.
func ipcOutput(r ipc.Result, err error) (Output, error) {
	if err != nil {
		return Output{}, err
	}
	out := Output{Lines: r.Log, Summary: summary(r.Counts)}
	for _, e := range r.Timeline {
		out.Timeline = append(out.Timeline, tick(e.Lane, e.What, e.State, e.Tick))
	}
	return out, nil
}

func ipcSims() []sim {
	return []sim{
		{
			id: "pipe", name: "Pipe", file: "os/ipc/pipe.go", fn: "Pipe",
			summary: "A writer and a reader on a small pipe: writes block when it is full, reads when it is empty, and close means end of file.",
			time:    "O(bytes + calls)", room: "O(capacity)", input: Text,
			example: Input{Text: "writer: hello world\nwriter: close\nreader: sleep\nreader: sleep\nreader: 3\nreader: 8\nreader: 8\nreader: 8", Options: map[string]int{"capacity": 4}},
			options: []Option{{"capacity", "pipe capacity in bytes", 4, 1, 64}},
			run:     func(in Input) (Output, error) { return ipcOutput(ipc.Pipe(in.Text, in.opt("capacity", 4))) },
			args:    func(in Input) []any { return []any{in.Text, in.opt("capacity", 4)} },
		},
		{
			id: "message-queue", name: "Message Queue", file: "os/ipc/messagequeue.go", fn: "MessageQueue",
			summary: "Messages keep their boundaries and come out by priority, oldest first among equals; a full queue says EAGAIN.",
			time:    "O(n log c)", room: "O(c)", input: Text,
			example: Input{Text: "send 1 low\nsend 5 urgent\nsend 1 also low\nreceive\nreceive\nsend 3 normal\nreceive\nreceive\nreceive", Options: map[string]int{"capacity": 4}},
			options: []Option{{"capacity", "queue capacity", 4, 1, 32}},
			run:     func(in Input) (Output, error) { return ipcOutput(ipc.MessageQueue(in.Text, in.opt("capacity", 4))) },
			args:    func(in Input) []any { return []any{in.Text, in.opt("capacity", 4)} },
		},
		{
			id: "shared-memory", name: "Shared Memory and Torn Reads", file: "os/ipc/sharedmemory.go", fn: "SharedMemory",
			summary: "Two processes share a record: without a semaphore the reader can copy it half written.",
			time:    "O(rounds)", room: "O(rounds)", input: Simulation,
			example: Input{Options: map[string]int{"rounds": 5, "sync": 0, "seed": 2}},
			options: []Option{{"rounds", "records written", 5, 1, 30}, {"sync", "use a semaphore (1) or not (0)", 0, 0, 1}, seedOption},
			run: func(in Input) (Output, error) {
				return ipcOutput(ipc.SharedMemory(in.opt("rounds", 5), on(in, "sync", 0), int64(in.opt("seed", 2))))
			},
			args: func(in Input) []any { return []any{in.opt("rounds", 5), on(in, "sync", 0), in.opt("seed", 2)} },
		},
		{
			id: "local-socket", name: "Local Socket Server", file: "os/ipc/socket.go", fn: "LocalSocket",
			summary: "Clients connect to a Unix-domain socket: queued in the listen backlog until accepted, refused when it is full.",
			time:    "O(T + clients)", room: "O(clients)", input: Simulation,
			example: Input{Options: map[string]int{"clients": 8, "backlog": 2, "accept": 2, "seed": 1}},
			options: []Option{{"clients", "clients", 8, 1, 30}, {"backlog", "listen backlog", 2, 0, 30}, {"accept", "ticks per accept", 2, 1, 10}, seedOption},
			run: func(in Input) (Output, error) {
				return ipcOutput(ipc.LocalSocket(in.opt("clients", 8), in.opt("backlog", 2), in.opt("accept", 2), int64(in.opt("seed", 1))))
			},
			args: func(in Input) []any {
				return []any{in.opt("clients", 8), in.opt("backlog", 2), in.opt("accept", 2), in.opt("seed", 1)}
			},
		},
		{
			id: "remote-procedure-call", name: "Remote Procedure Call", file: "os/ipc/rpc.go", fn: "RPC",
			summary: "Stubs marshal a call over a lossy network; a lost reply makes a retry run it twice, unless the server dedupes by request id.",
			time:    "O(retries)", room: "O(retries)", input: Text,
			example: Input{Text: "add 2 3", Options: map[string]int{"loss": 40, "retries": 4, "dedupe": 0, "seed": 3}},
			options: []Option{{"loss", "loss percent", 40, 0, 90}, {"retries", "retries", 4, 0, 10}, {"dedupe", "server dedupes by request id", 0, 0, 1}, seedOption},
			run: func(in Input) (Output, error) {
				return ipcOutput(ipc.RPC(in.Text, in.opt("loss", 40), in.opt("retries", 4), on(in, "dedupe", 0), int64(in.opt("seed", 3))))
			},
			args: func(in Input) []any {
				return []any{in.Text, in.opt("loss", 40), in.opt("retries", 4), on(in, "dedupe", 0), in.opt("seed", 3)}
			},
		},
	}
}

func fsSims() []sim {
	allocation := func(m filesystems.Method, id, name, summary string, ex string) sim {
		return sim{
			id: id, name: name, file: "os/filesystems/allocation.go", fn: "Allocate", summary: summary,
			time: "O(n * blocks)", room: "O(n * blocks)", input: Text,
			example: Input{Text: ex, Options: map[string]int{"blocks": 12}},
			options: []Option{{"blocks", "disk blocks", 12, 4, 64}},
			run: func(in Input) (Output, error) {
				r, err := filesystems.Allocate(m, in.opt("blocks", 12), in.Text)
				if err != nil {
					return Output{}, err
				}
				ids := map[string]int{}
				n := in.opt("blocks", 12)
				t := &TableOutput{Cols: n, Blank: ptr(0)}
				calls := linesOf(in.Text)
				for i, mp := range r.Maps {
					for _, owner := range mp {
						who := strings.TrimSuffix(owner, "*")
						if who != "" && ids[who] == 0 {
							ids[who] = len(ids) + 1
						}
						t.Cells = append(t.Cells, ids[who])
					}
					t.Rows++
					t.RowLabels = append(t.RowLabels, calls[i])
				}
				for b := 0; b < n; b++ {
					t.ColLabels = append(t.ColLabels, strconv.Itoa(b))
				}
				lines := r.Log
				if m == filesystems.Linked {
					lines = append(lines, fmt.Sprintf("file allocation table (next block, -1 end, -2 free): %v", r.FAT))
				}
				return Output{Lines: lines, Table: t, Summary: map[string]float64{"failed": float64(r.Failed)}}, nil
			},
			args: func(in Input) []any { return []any{string(m), in.opt("blocks", 12), in.Text} },
		}
	}
	const frag = "create a 3\ncreate b 2\ncreate c 4\ndelete a\ndelete c\ncreate d 5\ncreate e 2"
	return []sim{
		allocation(filesystems.Contiguous, "contiguous-allocation", "Contiguous Allocation", "Each file in one run of blocks: fast to read, but deletes leave holes too small to reuse.", frag),
		allocation(filesystems.Linked, "linked-allocation", "Linked Allocation (FAT)", "Each block points to the next in a file allocation table: no fragmentation, but reaching block k takes k hops.", frag),
		allocation(filesystems.Indexed, "indexed-allocation", "Indexed Allocation", "Each file has an index block listing its blocks: any block in one lookup, for one extra block per file.", frag),
		{
			id: "inode-pointers", name: "Inode Pointers and Maximum File Size", file: "os/filesystems/inode.go", fn: "Inode",
			summary: "Direct, single, double and triple indirect pointers: how big a file can get, and how many reads reach a byte.",
			time:    "O(1)", room: "O(1)", input: Simulation,
			example: Input{Options: map[string]int{"block": 4096, "pointer": 4, "direct": 12, "offset": 5000000}},
			options: []Option{{"block", "block size in bytes", 4096, 64, 65536}, {"pointer", "pointer size in bytes", 4, 4, 8}, {"direct", "direct pointers", 12, 0, 32}, {"offset", "byte offset to find", 5000000, 0, 2_000_000_000}},
			run: func(in Input) (Output, error) {
				l, err := filesystems.Inode(in.opt("block", 4096), in.opt("pointer", 4), in.opt("direct", 12), int64(in.opt("offset", 5000000)))
				if err != nil {
					return Output{}, err
				}
				t := &TableOutput{Cols: 2, ColLabels: []string{"blocks", "MiB"}}
				lines := []string{fmt.Sprintf("%d pointers per indirect block; largest file %d bytes (%.1f GiB)", l.PointersPerBlock, l.MaxBytes, float64(l.MaxBytes)/(1<<30))}
				for _, lv := range l.Levels {
					t.Rows++
					t.Cells = append(t.Cells, int(lv.Blocks), int(lv.Bytes>>20))
					t.RowLabels = append(t.RowLabels, lv.Name)
				}
				return Output{Lines: append(lines, l.Lookup...), Table: t, Number: ptr(int(l.MaxBytes)), Summary: map[string]float64{"disk reads": float64(l.DiskReads)}}, nil
			},
			args: func(in Input) []any {
				return []any{in.opt("block", 4096), in.opt("pointer", 4), in.opt("direct", 12), in.opt("offset", 5000000)}
			},
		},
		{
			id: "hard-and-symbolic-links", name: "Hard and Symbolic Links", file: "os/filesystems/namespace.go", fn: "Namespace",
			summary: "Names, inodes and link counts: a hard link keeps a file alive, a symbolic link only remembers its path.",
			time:    "O(n * d)", room: "O(inodes)", input: Text,
			example: Input{Text: "mkdir /docs\ncreate /docs/notes.txt buy milk\nlink /docs/notes.txt /backup\nsymlink /docs/notes.txt /shortcut\ncat /shortcut\nrm /docs/notes.txt\ncat /backup\ncat /shortcut"},
			run: func(in Input) (Output, error) {
				r, err := filesystems.Namespace(in.Text)
				if err != nil {
					return Output{}, err
				}
				return Output{Lines: r.Log, Table: inodeTable(r.Inodes)}, nil
			},
			args: func(in Input) []any { return []any{in.Text} },
		},
		{
			id: "path-resolution", name: "Path Resolution", file: "os/filesystems/namespace.go", fn: "Resolve",
			summary: "Follow a path component by component through directory inodes and symbolic links, as the kernel does on every open.",
			time:    "O(n * d)", room: "O(inodes)", input: TwoTexts,
			example: Input{Text: "mkdir /home\nmkdir /home/ada\nmkdir /home/ada/code\ncreate /home/ada/code/main.go package main\nsymlink /home/ada/code /src", Other: "/src/main.go"},
			run: func(in Input) (Output, error) {
				r, err := filesystems.Resolve(in.Text, in.Other)
				if err != nil {
					return Output{}, err
				}
				return Output{Lines: r.Log, Table: inodeTable(r.Inodes)}, nil
			},
			args: func(in Input) []any { return []any{in.Text, in.Other} },
		},
		{
			id: "free-space-management", name: "Free Space: Bitmap and Free List", file: "os/filesystems/freespace.go", fn: "FreeSpace",
			summary: "Track free blocks with a bitmap (fixed size, easy to scan) or a list of free runs (small until space fragments).",
			time:    "O(n * blocks)", room: "O(n * blocks)", input: Text,
			example: Input{Text: "alloc a 4\nalloc b 3\nalloc c 5\nfree b\nalloc d 2\nfree a", Options: map[string]int{"blocks": 16}},
			options: []Option{{"blocks", "disk blocks", 16, 8, 256}},
			run: func(in Input) (Output, error) {
				r, err := filesystems.FreeSpace(in.opt("blocks", 16), in.Text)
				if err != nil {
					return Output{}, err
				}
				n := in.opt("blocks", 16)
				t := &TableOutput{Cols: n}
				var lines []string
				for _, s := range r.Steps {
					t.Rows++
					for _, c := range s.Bitmap {
						t.Cells = append(t.Cells, int(c-'0'))
					}
					t.RowLabels = append(t.RowLabels, s.Call)
					lines = append(lines, fmt.Sprintf("%s: bitmap %s, free list %v", s.Call, s.Bitmap, s.List))
				}
				return Output{Lines: lines, Table: t, Summary: map[string]float64{"bitmap bytes": float64(r.BitmapBytes), "free list bytes": float64(r.ListBytes), "bits scanned": float64(r.Scanned)}}, nil
			},
			args: func(in Input) []any { return []any{in.opt("blocks", 16), in.Text} },
		},
		{
			id: "journaling", name: "Journaling and Crash Recovery", file: "os/filesystems/journal.go", fn: "Journaling",
			summary: "Crash in the middle of appending a block, with and without a journal, and see what recovery finds.",
			time:    "O(1)", room: "O(1)", input: Simulation,
			example: Input{Options: map[string]int{"crash": 2, "journal": 0}},
			options: []Option{{"crash", "crash after this many writes", 2, 0, 10}, {"journal", "journaled (1) or not (0)", 0, 0, 1}},
			run: func(in Input) (Output, error) {
				r, err := filesystems.Journaling(in.opt("crash", 2), on(in, "journal", 0))
				if err != nil {
					return Output{}, err
				}
				lines := append(append(r.Writes, "recovery:"), r.Recovery...)
				return Output{Lines: lines, Bool: ptr(r.Consistent)}, nil
			},
			args: func(in Input) []any { return []any{in.opt("crash", 2), on(in, "journal", 0)} },
		},
		{
			id: "page-cache", name: "Page Cache: Write-Through and Write-Back", file: "os/filesystems/pagecache.go", fn: "PageCache",
			summary: "Write-through pays the disk on every write; write-back batches writes and loses the dirty ones in a crash.",
			time:    "O(n log p)", room: "O(p)", input: Text,
			example: Input{Text: "write a\nwrite a\nwrite b\nread a\nread c\nsync\nwrite c\nwrite a\ncrash\nread a", Options: map[string]int{"back": 1}},
			options: []Option{{"back", "write-back (1) or write-through (0)", 1, 0, 1}},
			run: func(in Input) (Output, error) {
				r, err := filesystems.PageCache(on(in, "back", 1), in.Text)
				if err != nil {
					return Output{}, err
				}
				return Output{Lines: r.Log, Summary: map[string]float64{"disk writes": float64(r.DiskWrites), "disk reads": float64(r.DiskReads), "cache hits": float64(r.Hits), "writes lost": float64(r.Lost)}}, nil
			},
			args: func(in Input) []any { return []any{on(in, "back", 1), in.Text} },
		},
		{
			id: "raid", name: "RAID 0, 1, 5 and 6", file: "os/filesystems/raid.go", fn: "RAID",
			summary: "Stripe, mirror, or add parity across disks; fail one or two and rebuild the data with XOR and Reed-Solomon.",
			time:    "O(n * disks)", room: "O(n)", input: Ints,
			example: Input{Values: []int{10, 20, 30, 40, 50, 60, 70, 80, 90}, Options: map[string]int{"level": 5, "disks": 4, "fail": 1, "fail2": -1}},
			options: []Option{{"level", "RAID level (0, 1, 5, 6)", 5, 0, 6}, {"disks", "disks", 4, 2, 8}, {"fail", "a failed disk (-1 none)", 1, -1, 7}, {"fail2", "a second failed disk (-1 none)", -1, -1, 7}},
			run: func(in Input) (Output, error) {
				r, err := filesystems.RAID(in.opt("level", 5), in.opt("disks", 4), in.Values, failedDisks(in))
				if err != nil {
					return Output{}, err
				}
				var t *TableOutput
				if len(r.Disks) > 0 && len(r.Disks[0]) > 0 {
					t = &TableOutput{Rows: len(r.Disks), Cols: len(r.Disks[0])}
					down := map[int]bool{}
					for _, d := range r.Failed {
						down[d] = true
					}
					for d, blocks := range r.Disks {
						t.Cells = append(t.Cells, blocks...)
						label := fmt.Sprintf("disk %d", d)
						if down[d] {
							label += " (failed)"
						}
						t.RowLabels = append(t.RowLabels, label)
					}
					for s := range r.Disks[0] {
						var roles []string
						for d := range r.Disks {
							roles = append(roles, r.Roles[d][s])
						}
						t.ColLabels = append(t.ColLabels, "stripe "+strconv.Itoa(s))
						r.Log = append(r.Log, fmt.Sprintf("stripe %d: %s", s, strings.Join(roles, " | ")))
					}
				}
				return Output{Lines: r.Log, Table: t, Values: r.Recovered, Bool: ptr(r.Intact), Summary: map[string]float64{"usable fraction": r.Usable}}, nil
			},
			args: func(in Input) []any {
				return []any{in.opt("level", 5), in.opt("disks", 4), in.Values, rawInts(failedDisks(in))}
			},
		},
		{
			id: "log-structured-fs", name: "Log-Structured File System Cleaning", file: "os/filesystems/lfs.go", fn: "LFS",
			summary: "Every write appends to a log; overwrites leave dead blocks, and the cleaner copies live ones out of the emptiest segment.",
			time:    "O(n * segments * size)", room: "O(segments * size)", input: Ints,
			example: Input{Values: []int{1, 2, 3, 4, 1, 2, 5, 1, 2, 6, 1, 2, 7, 1, 2, 8}, Options: map[string]int{"segments": 4, "size": 4}},
			options: []Option{{"segments", "segments", 4, 3, 16}, {"size", "blocks per segment", 4, 2, 16}},
			run: func(in Input) (Output, error) {
				r, err := filesystems.LFS(in.opt("segments", 4), in.opt("size", 4), in.Values)
				if err != nil {
					return Output{}, err
				}
				t := &TableOutput{Blank: ptr(-1)}
				for s, slots := range r.Segments {
					t.Rows++
					t.Cols = len(slots)
					t.Cells = append(t.Cells, slots...)
					t.RowLabels = append(t.RowLabels, "segment "+strconv.Itoa(s))
				}
				return Output{Lines: r.Log, Table: t, Summary: map[string]float64{"logical writes": float64(r.Logical), "disk writes": float64(r.Physical), "cleanings": float64(r.Cleaned), "write amplification": float64(r.Physical) / float64(r.Logical)}}, nil
			},
			args: func(in Input) []any { return []any{in.opt("segments", 4), in.opt("size", 4), in.Values} },
		},
	}
}

// rawInts is a []int printed as Go source.
type rawInts []int

func (r rawInts) String() string { return ints(r) }

func failedDisks(in Input) []int {
	var out []int
	for _, k := range []string{"fail", "fail2"} {
		if d := in.opt(k, -1); d >= 0 {
			out = append(out, d)
		}
	}
	return out
}

func inodeTable(inodes []filesystems.INode) *TableOutput {
	t := &TableOutput{Cols: 1, ColLabels: []string{"links"}}
	for _, n := range inodes {
		t.Rows++
		t.Cells = append(t.Cells, n.Links)
		label := fmt.Sprintf("inode %d (%s)", n.Number, n.Kind)
		if n.Kind == "dir" {
			var names []string
			for name := range n.Entries {
				if name != "." && name != ".." {
					names = append(names, name)
				}
			}
			sort.Strings(names)
			label += ": " + strings.Join(names, ", ")
		} else if n.Data != "" {
			label += fmt.Sprintf(": %q", n.Data)
		}
		t.RowLabels = append(t.RowLabels, label)
	}
	return t
}

// linesOf returns a script's non-empty lines, trimmed.
func linesOf(text string) []string {
	var out []string
	for _, l := range strings.Split(text, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out
}

func vmSims() []sim {
	return []sim{
		{
			id: "page-table-walk", name: "Multi-Level Page Table Walk", file: "os/virtualmemory/pagewalk.go", fn: "PageWalk",
			summary: "Split a virtual address into per-level indices and walk the tables to a frame, or to a page fault.",
			time:    "O(levels + mappings)", room: "O(mappings * levels)", input: Text,
			example: Input{Text: "0x401 7\n0x402 9\n0x80000 3", Options: map[string]int{"address": 0x401abc, "levels": 2, "bits": 10, "offset": 12}},
			options: []Option{{"address", "virtual address", 0x401abc, 0, 1 << 30}, {"levels", "levels", 2, 1, 5}, {"bits", "index bits per level", 10, 1, 12}, {"offset", "offset bits", 12, 4, 16}},
			run: func(in Input) (Output, error) {
				w, err := pageWalk(in)
				if err != nil {
					return Output{}, err
				}
				return Output{Lines: w.Steps, Values: w.Indices, Number: ptr(int(w.Physical)), Summary: map[string]float64{"memory reads": float64(w.MemoryReads), "table pages": float64(w.Tables), "flat table entries": float64(w.FlatEntries)}}, nil
			},
			args: func(in Input) []any {
				return []any{in.Text, in.opt("address", 0x401abc), rawInts(levelBits(in)), in.opt("offset", 12)}
			},
		},
		{
			id: "tlb-effective-access-time", name: "TLB and Effective Access Time", file: "os/virtualmemory/tlb.go", fn: "TLB",
			summary: "Run page references through a small TLB and compute the effective access time: why locality makes paging cheap.",
			time:    "O(n * size)", room: "O(n * size)", input: Ints,
			example: Input{Values: []int{1, 2, 1, 3, 1, 2, 4, 1, 2, 5, 1, 2}, Options: map[string]int{"size": 3, "tlb": 1, "mem": 100, "levels": 4}},
			options: []Option{{"size", "TLB entries", 3, 1, 64}, {"tlb", "TLB time (ns)", 1, 0, 100}, {"mem", "memory time (ns)", 100, 1, 1000}, {"levels", "page table levels", 4, 1, 5}},
			run: func(in Input) (Output, error) {
				r, err := virtualmemory.TLB(in.Values, in.opt("size", 3), in.opt("tlb", 1), in.opt("mem", 100), in.opt("levels", 4))
				if err != nil {
					return Output{}, err
				}
				size := in.opt("size", 3)
				t := &TableOutput{Rows: size, Cols: len(in.Values), Cells: make([]int, size*len(in.Values)), Blank: ptr(-1)}
				for c, contents := range r.Contents {
					for row := 0; row < size; row++ {
						v := -1
						if row < len(contents) {
							v = contents[row]
						}
						t.Cells[row*len(in.Values)+c] = v
					}
					mark := "miss"
					if r.Hits[c] {
						mark = "hit"
					}
					t.ColLabels = append(t.ColLabels, fmt.Sprintf("%d %s", in.Values[c], mark))
				}
				for row := 0; row < size; row++ {
					t.RowLabels = append(t.RowLabels, "entry "+strconv.Itoa(row))
				}
				return Output{Lines: r.Log, Table: t, Summary: map[string]float64{"hit rate": r.HitRate, "effective ns": r.Effective, "measured ns": r.Measured}}, nil
			},
			args: func(in Input) []any {
				return []any{in.Values, in.opt("size", 3), in.opt("tlb", 1), in.opt("mem", 100), in.opt("levels", 4)}
			},
		},
		{
			id: "inverted-page-table", name: "Inverted Page Table", file: "os/virtualmemory/inverted.go", fn: "InvertedPageTable",
			summary: "One entry per physical frame, found by hashing (process, page): a table the size of memory, not of address spaces.",
			time:    "O(n) average, O(n * frames) worst", room: "O(n * frames)", input: Text,
			example: Input{Text: "1:0 1:1 2:0 1:0 2:1 3:0 1:2 2:0", Options: map[string]int{"frames": 4}},
			options: []Option{{"frames", "frames", 4, 1, 32}},
			run: func(in Input) (Output, error) {
				r, err := virtualmemory.InvertedPageTable(in.opt("frames", 4), in.Text)
				if err != nil {
					return Output{}, err
				}
				total := 0
				for _, p := range r.Probes {
					total += p
				}
				return Output{Lines: r.Log, Values: r.Probes, Summary: map[string]float64{"faults": float64(r.Faults), "average probes": float64(total) / float64(len(r.Probes))}}, nil
			},
			args: func(in Input) []any { return []any{in.opt("frames", 4), in.Text} },
		},
		{
			id: "working-set", name: "Working Set", file: "os/virtualmemory/workingset.go", fn: "WorkingSet",
			summary: "The pages used in the last few references: how many frames a process needs right now, phase by phase.",
			time:    "O(n * window)", room: "O(n * window)", input: Ints,
			example: Input{Values: []int{1, 2, 1, 3, 1, 2, 7, 8, 9, 7, 8, 9, 7, 1, 2, 1}, Options: map[string]int{"window": 4}},
			options: []Option{{"window", "window (references)", 4, 1, 50}},
			run: func(in Input) (Output, error) {
				r, err := virtualmemory.WorkingSet(in.Values, in.opt("window", 4))
				if err != nil {
					return Output{}, err
				}
				var lines []string
				for i, s := range r.Sets {
					lines = append(lines, fmt.Sprintf("after page %d: working set %v (%d pages)", in.Values[i], s, len(s)))
				}
				return Output{Lines: lines, Values: r.Sizes, Number: ptr(r.Peak)}, nil
			},
			args: func(in Input) []any { return []any{in.Values, in.opt("window", 4)} },
		},
		{
			id: "thrashing", name: "Thrashing", file: "os/virtualmemory/workingset.go", fn: "Thrashing",
			summary: "More processes keep the CPU busier until their working sets stop fitting; then utilization collapses.",
			time:    "O(processes)", room: "O(processes)", input: Simulation,
			example: Input{Options: map[string]int{"frames": 100, "ws": 20, "io": 70, "procs": 10}},
			options: []Option{{"frames", "frames", 100, 1, 10000}, {"ws", "working set per process", 20, 1, 10000}, {"io", "I/O wait percent", 70, 1, 99}, {"procs", "most processes", 10, 1, 50}},
			run: func(in Input) (Output, error) {
				pts, err := virtualmemory.Thrashing(in.opt("frames", 100), in.opt("ws", 20), in.opt("io", 70), in.opt("procs", 10))
				if err != nil {
					return Output{}, err
				}
				t := &TableOutput{Cols: 1, ColLabels: []string{"CPU utilization %"}}
				var lines []string
				for _, p := range pts {
					t.Rows++
					t.Cells = append(t.Cells, int(100*p.Utilization+0.5))
					t.RowLabels = append(t.RowLabels, fmt.Sprintf("%d processes", p.Processes))
					lines = append(lines, fmt.Sprintf("%d processes: %.0f%% (%s)", p.Processes, 100*p.Utilization, p.Note))
				}
				return Output{Lines: lines, Table: t}, nil
			},
			args: func(in Input) []any {
				return []any{in.opt("frames", 100), in.opt("ws", 20), in.opt("io", 70), in.opt("procs", 10)}
			},
		},
		{
			id: "copy-on-write", name: "Copy-on-Write Fork", file: "os/virtualmemory/cow.go", fn: "CopyOnWrite",
			summary: "After fork, parent and child share every frame read-only; a write faults and copies just that page.",
			time:    "O(n + pages)", room: "O(n * pages)", input: Text,
			example: Input{Text: "child read 0\nchild write 1\nparent read 1\nparent write 2\nchild write 1\nparent write 1", Options: map[string]int{"pages": 4}},
			options: []Option{{"pages", "pages", 4, 1, 32}},
			run: func(in Input) (Output, error) {
				r, err := virtualmemory.CopyOnWrite(in.opt("pages", 4), in.Text)
				if err != nil {
					return Output{}, err
				}
				pages := in.opt("pages", 4)
				t := &TableOutput{Cols: 2 * pages}
				for p := 0; p < pages; p++ {
					t.ColLabels = append(t.ColLabels, fmt.Sprintf("parent %d", p))
				}
				for p := 0; p < pages; p++ {
					t.ColLabels = append(t.ColLabels, fmt.Sprintf("child %d", p))
				}
				steps := linesOf(in.Text)
				for i := range r.Parent {
					t.Rows++
					t.Cells = append(append(t.Cells, r.Parent[i]...), r.Child[i]...)
					t.RowLabels = append(t.RowLabels, steps[i])
				}
				return Output{Lines: r.Log, Table: t, Summary: map[string]float64{"pages copied": float64(r.Copies), "eager fork copies": float64(r.Eager)}}, nil
			},
			args: func(in Input) []any { return []any{in.opt("pages", 4), in.Text} },
		},
		{
			id: "memory-mapped-file", name: "Memory-Mapped Files", file: "os/virtualmemory/mmap.go", fn: "MemoryMapped",
			summary: "Read a file with read() (a system call and a copy each time) or map it with mmap() (a fault on first touch, then plain memory).",
			time:    "O(accesses)", room: "O(pages)", input: Ints,
			example: Input{Values: []int{0, 1, 2, 0, 1, 0, 3, 0}, Options: map[string]int{"pages": 8, "write": 1}},
			options: []Option{{"pages", "file pages", 8, 1, 256}, {"write", "a page to write (-1 none)", 1, -1, 255}},
			run: func(in Input) (Output, error) {
				r, err := virtualmemory.MemoryMapped(in.opt("pages", 8), in.Values, writesOf(in))
				if err != nil {
					return Output{}, err
				}
				return Output{Lines: r.Log, Summary: map[string]float64{"read() calls": float64(r.ReadCalls), "read() copies": float64(r.ReadCopies), "mmap faults": float64(r.MmapFaults), "disk reads": float64(r.DiskReads), "dirty pages written": float64(r.DirtyWritten)}}, nil
			},
			args: func(in Input) []any { return []any{in.opt("pages", 8), in.Values, rawInts(writesOf(in))} },
		},
	}
}

func writesOf(in Input) []int {
	if w := in.opt("write", -1); w >= 0 {
		return []int{w}
	}
	return nil
}

func levelBits(in Input) []int {
	bits := make([]int, in.opt("levels", 2))
	for i := range bits {
		bits[i] = in.opt("bits", 10)
	}
	return bits
}

func pageWalk(in Input) (virtualmemory.Walk, error) {
	return virtualmemory.PageWalk(in.Text, int64(in.opt("address", 0x401abc)), levelBits(in), in.opt("offset", 12))
}

func bootSims() []sim {
	stages := func(st []boot.Stage) []Span {
		out := make([]Span, len(st))
		for i, s := range st {
			out[i] = Span{Lane: s.Lane, Label: s.What, Start: s.Start, End: s.End, State: s.State}
		}
		return out
	}
	return []sim{
		{
			id: "boot-sequence", name: "Boot Sequence", file: "os/boot/sequence.go", fn: "Sequence",
			summary: "From the reset vector through firmware, the boot loader and the kernel to init and a login prompt, BIOS or UEFI.",
			time:    "O(services)", room: "O(services)", input: Simulation,
			example: Input{Options: map[string]int{"uefi": 1, "services": 6}},
			options: []Option{{"uefi", "UEFI (1) or legacy BIOS (0)", 1, 0, 1}, {"services", "services init starts", 6, 1, 12}},
			run: func(in Input) (Output, error) {
				r, err := boot.Sequence(on(in, "uefi", 1), in.opt("services", 6))
				if err != nil {
					return Output{}, err
				}
				return Output{Timeline: stages(r.Stages), Lines: r.Notes, Number: ptr(r.Total)}, nil
			},
			args: func(in Input) []any { return []any{on(in, "uefi", 1), in.opt("services", 6)} },
		},
		{
			id: "nested-interrupts", name: "Nested Interrupts and Priorities", file: "os/boot/interrupts.go", fn: "Interrupts",
			summary: "Devices raise interrupts: the CPU saves its state, looks up the vector table, and lets higher lines preempt lower handlers.",
			time:    "O(ticks * irqs)", room: "O(ticks)", input: Text,
			example: Input{Text: "2 1 3\n3 14 4\n4 0 1\n9 12 2", Options: map[string]int{"ticks": 12}},
			options: []Option{{"ticks", "ticks of program work", 12, 1, 200}},
			run: func(in Input) (Output, error) {
				r, err := boot.Interrupts(in.Text, in.opt("ticks", 12))
				if err != nil {
					return Output{}, err
				}
				return Output{Timeline: stages(r.Stages), Lines: r.Log, Values: r.Latency}, nil
			},
			args: func(in Input) []any { return []any{in.Text, in.opt("ticks", 12)} },
		},
		{
			id: "dma-vs-programmed-io", name: "DMA and Programmed I/O", file: "os/boot/dma.go", fn: "DMA",
			summary: "Copy a device's data word by word with the CPU, or let a DMA controller move it while the CPU does other work.",
			time:    "O(bytes / word)", room: "O(bytes / word)", input: Simulation,
			example: Input{Options: map[string]int{"bytes": 64, "word": 4, "block": 32}},
			options: []Option{{"bytes", "bytes to move", 64, 1, 65536}, {"word", "bytes per bus transfer (1, 2, 4, 8)", 4, 1, 8}, {"block", "bytes per DMA block", 32, 1, 65536}},
			run: func(in Input) (Output, error) {
				r, err := boot.DMA(in.opt("bytes", 64), in.opt("word", 4), in.opt("block", 32))
				if err != nil {
					return Output{}, err
				}
				return Output{Timeline: stages(r.Stages), Lines: []string{
					fmt.Sprintf("programmed I/O: the CPU is busy for %d ticks", r.PIOCPU),
					fmt.Sprintf("DMA: the CPU spends %d ticks (setup and %d interrupts) and is free for %d", r.DMACPU, r.Interrupts, r.DMAFree),
				}, Summary: map[string]float64{"PIO CPU ticks": float64(r.PIOCPU), "DMA CPU ticks": float64(r.DMACPU)}}, nil
			},
			args: func(in Input) []any { return []any{in.opt("bytes", 64), in.opt("word", 4), in.opt("block", 32)} },
		},
	}
}
