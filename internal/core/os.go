package core

import (
	"fmt"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/os/boot"
	"github.com/braswelljr/alGOrithm/pkg/os/deadlock"
	"github.com/braswelljr/alGOrithm/pkg/os/disk"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// diskEntries is every entry of the disk category.
func diskEntries() []Entry {
	return []Entry{
		cLookDiskEntry(),
		cScanDiskEntry(),
		fcfsDiskEntry(),
		lookDiskEntry(),
		scanDiskEntry(),
		sstfDiskEntry(),
	}
}

// deadlockEntries is every entry of the deadlock category.
func deadlockEntries() []Entry {
	return []Entry{
		bankersAlgorithmEntry(),
	}
}

// bootEntries is every entry of the boot category.
func bootEntries() []Entry {
	return simEntries(bootSims())
}

var diskOpts = []Option{{"head", "head starts at", 53, 0, 100_000}, {"size", "tracks", 200, 1, 100_000}, {"down", "moving down (1 yes, 0 no)", 0, 0, 1}}

func diskMk(id, name, fn, summary, time string, run func(reqs []int, head, size int, down bool) (disk.Result, error)) Entry {
	e := entry(id, name, "os/disk/disk.go", fn)
	e.Summary, e.Time, e.Space, e.Input, e.Options, e.Traced = summary, time, "O(n)", Tracks, diskOpts, true
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

func diskIgnoreDir(f func([]int, int, int) (disk.Result, error)) func([]int, int, int, bool) (disk.Result, error) {
	return func(r []int, h, s int, _ bool) (disk.Result, error) { return f(r, h, s) }
}

func fcfsDiskEntry() Entry {
	return diskMk("fcfs-disk", "FCFS Disk Scheduling", "FCFS", "Serve requests in arrival order: fair, but the head swings back and forth.", "O(n)", diskIgnoreDir(disk.FCFS))
}

func sstfDiskEntry() Entry {
	return diskMk("sstf-disk", "Shortest Seek Time First", "SSTF", "Serve the nearest request next: less movement, but far requests can starve.", "O(n^2)", diskIgnoreDir(disk.SSTF))
}

func scanDiskEntry() Entry {
	return diskMk("scan-disk", "SCAN (Elevator)", "SCAN", "Sweep to one end serving requests, then reverse, like an elevator.", "O(n log n)", disk.SCAN)
}

func cScanDiskEntry() Entry {
	return diskMk("c-scan-disk", "C-SCAN", "CSCAN", "Sweep one way to the end, jump back to the start, sweep again: more even waits.", "O(n log n)", disk.CSCAN)
}

func lookDiskEntry() Entry {
	return diskMk("look-disk", "LOOK", "LOOK", "SCAN that turns at the last request instead of the disk's end.", "O(n log n)", disk.LOOK)
}

func cLookDiskEntry() Entry {
	return diskMk("c-look-disk", "C-LOOK", "CLOOK", "C-SCAN that jumps from the last request to the first instead of between the ends.", "O(n log n)", disk.CLOOK)
}

const bankerExample = `available 3 3 2
P0 max 7 5 3 alloc 0 1 0
P1 max 3 2 2 alloc 2 0 0
P2 max 9 0 2 alloc 3 0 2
P3 max 2 2 2 alloc 2 1 1
P4 max 4 3 3 alloc 0 0 2`

func bankersAlgorithmEntry() Entry {
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
	return b
}

// rawInts is a []int printed as Go source.
type rawInts []int

func (r rawInts) String() string { return ints(r) }

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
