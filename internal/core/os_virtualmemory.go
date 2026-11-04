package core

import (
	"fmt"
	"strconv"

	"github.com/braswelljr/alGOrithm/pkg/os/virtualmemory"
)

// virtualmemoryEntries is every entry of the virtualmemory category.
func virtualmemoryEntries() []Entry {
	return simEntries(vmSims())
}

var vmSimTable = []sim{
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

func vmSims() []sim {
	return vmSimTable
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
