package core

import (
	"fmt"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/os/processes"
)

// processesEntries is every entry of the processes category.
func processesEntries() []Entry {
	return simEntries(processSims())
}

var processSimStateOf = map[processes.State]string{processes.Ready: "idle", processes.Running: "running", processes.Waiting: "blocked", processes.Terminated: "done"}

var processSimTable = []sim{
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
				out.Timeline = append(out.Timeline, tick("process", string(t.To), processSimStateOf[t.To], t.Step-1))
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

func processSims() []sim {
	return processSimTable
}
