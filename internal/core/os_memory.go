package core

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/os/memory"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// memoryEntries is every entry of the memory category.
func memoryEntries() []Entry {
	return []Entry{
		addressTranslationEntry(),
		agingPageReplacementEntry(),
		bestFitEntry(),
		buddyAllocatorEntry(),
		clockPageReplacementEntry(),
		fifoPageReplacementEntry(),
		firstFitEntry(),
		lfuPageReplacementEntry(),
		lruPageReplacementEntry(),
		optimalPageReplacementEntry(),
		worstFitEntry(),
	}
}

// pageOutput turns a page-replacement result into the frames-over-time table
// the memory entries share, reporting each cell and each hit or fault to r.
func pageOutput(res memory.Result, frames int, r trace.Recorder) Output {
	n := len(res.Steps)
	t := &TableOutput{Rows: frames, Cols: n, Cells: make([]int, frames*n), Blank: ptr(-1)}
	for c, st := range res.Steps {
		t.ColLabels = append(t.ColLabels, strconv.Itoa(st.Page))
		for row, p := range st.Frames {
			t.Cells[row*n+c] = p
			r.Record(trace.Event{Kind: trace.Set, Indices: []int{row*n + c}, Values: []int{p}})
		}
		note := "hit"
		if st.Fault {
			note = "fault"
		}
		r.Record(trace.Event{Kind: trace.Mark, Indices: []int{st.Frame*n + c}, Note: note})
	}
	for f := 0; f < frames; f++ {
		t.RowLabels = append(t.RowLabels, "frame "+strconv.Itoa(f))
	}
	return Output{Table: t, Number: ptr(res.Faults), Summary: map[string]float64{
		"faults": float64(res.Faults), "hits": float64(res.Hits), "fault rate": float64(res.Faults) / float64(n),
	}}
}

// pageTable returns the address-translation input's page table: the Text
// ("page frame" lines), or one made up from the seed, with about a quarter of
// the pages not resident.
func pageTable(in Input) (map[int]int, error) {
	if strings.TrimSpace(in.Text) != "" {
		t, err := memory.ParsePageTable(in.Text)
		return t, wrap(err)
	}
	r := rand.New(rand.NewSource(int64(in.opt("seed", 1))))
	pages := in.opt("pages", 8)
	frames := r.Perm(pages)
	t := map[int]int{}
	for p := 0; p < pages; p++ {
		if r.Intn(4) != 0 {
			t[p] = frames[p]
		}
	}
	return t, nil
}

func memoryPage(id, name, file, fn, summary string, run func([]int, int) (memory.Result, error)) Entry {
	e := entry(id, name, "os/memory/"+file, fn)
	e.Summary, e.Time, e.Space, e.Input, e.Traced = summary, "O(n * frames)", "O(n * frames)", Pages, true
	e.Options = []Option{{"frames", "frames", 3, 1, memory.MaxFrames}}
	e.Example = &Input{Values: memoryRefs, Options: map[string]int{"frames": 3}}
	e.Run = func(in Input, r trace.Recorder) (Output, error) {
		res, err := run(in.Values, in.opt("frames", 3))
		if err != nil {
			return Output{}, wrap(err)
		}
		return pageOutput(res, in.opt("frames", 3), r), nil
	}
	e.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("memory.%s(%s, %d)", fn, ints(in.Values), in.opt("frames", 3))
	}
	return e
}

func buddyAllocatorEntry() Entry {
	buddy := structureEntry("buddy-allocator", "Buddy Allocator", "os/memory/buddy.go", "Buddy",
		"Round each request up to a power of two, split bigger blocks in halves to fit, and merge buddies back when both are free.",
		"O(log(size / smallest block)) alloc and free", "O(blocks)",
		"alloc", []string{"alloc NAME SIZE", "free NAME"}, func(in Input) playground {
			b, err := memory.NewBuddy(in.opt("size", 1024), in.opt("min", 32))
			if err != nil {
				b, _ = memory.NewBuddy(1024, 32)
			}
			return &buddyP{b: b}
		})
	buddy.Category = "memory"
	buddy.Options = []Option{{"size", "arena size (power of two)", 1024, 2, 1 << 20}, {"min", "smallest block (power of two)", 32, 1, 1 << 20}}
	buddy.Example = &Input{Ops: ParseOps("alloc A 100, alloc B 240, alloc C 60, alloc D 130, free C, free A, free B"), Options: map[string]int{"size": 1024, "min": 32}}
	return buddy
}

func addressTranslationEntry() Entry {
	tr := entry("address-translation", "Address Translation (paging and TLB)", "os/memory/translation.go", "Translate")
	tr.Summary, tr.Time, tr.Space, tr.Input = "Split virtual addresses into page and offset, look the page up (in a TLB first), and see faults and TLB hits.", "O(n * TLB size)", "O(n)", Pages
	tr.Options = []Option{{"pagesize", "page size (power of two)", 256, 1, 1 << 20}, {"tlb", "TLB entries (0 for none)", 2, 0, memory.MaxFrames}, {"pages", "pages, when the table is made up", 8, 1, 4096}, {"seed", "made-up table seed", 1, 0, 1_000_000}}
	tr.Example = &Input{Values: []int{20, 300, 520, 270, 1050, 40, 300, 1800, 700, 20}, Text: "0 5\n1 2\n2 7\n4 1\n7 3", Options: map[string]int{"pagesize": 256, "tlb": 2}}
	tr.Limits = "the page table is the text (\"page frame\" lines), or made up from the seed"
	tr.Run = func(in Input, _ trace.Recorder) (Output, error) {
		table, err := pageTable(in)
		if err != nil {
			return Output{}, err
		}
		res, err := memory.Translate(in.opt("pagesize", 256), table, in.Values, in.opt("tlb", 2))
		if err != nil {
			return Output{}, wrap(err)
		}
		cols := []string{"virtual", "page", "offset", "frame", "physical", "TLB hit"}
		t := &TableOutput{Rows: len(res.Steps), Cols: len(cols), ColLabels: cols}
		for _, s := range res.Steps {
			hit := 0
			if s.TLBHit {
				hit = 1
			}
			t.Cells = append(t.Cells, s.Virtual, s.Page, s.Offset, s.Frame, s.Physical, hit)
			label := strconv.Itoa(s.Virtual)
			if s.Fault {
				label += " (fault)"
			}
			t.RowLabels = append(t.RowLabels, label)
		}
		pages := make([]int, 0, len(table))
		for p := range table {
			pages = append(pages, p)
		}
		sort.Ints(pages)
		line := "page table:"
		for _, p := range pages {
			line += fmt.Sprintf(" %d->%d", p, table[p])
		}
		n := float64(len(res.Steps))
		return Output{Table: t, Lines: []string{line}, Number: ptr(res.Faults), Summary: map[string]float64{
			"page faults": float64(res.Faults), "TLB hits": float64(res.TLBHits), "TLB hit rate": float64(res.TLBHits) / n}}, nil
	}
	return tr
}

func lfuPageReplacementEntry() Entry {
	return memoryPage("lfu-page-replacement", "LFU Page Replacement", "frequency.go", "LFU", "Evict the page used least often since it was loaded; keeps popular pages, adapts slowly.", memory.LFU)
}

func agingPageReplacementEntry() Entry {
	return memoryPage("aging-page-replacement", "Aging Page Replacement", "frequency.go", "Aging", "An 8-bit counter per frame, shifted right on every reference with the used frame's top bit set: LRU on a budget.", memory.Aging)
}

var memoryRefs = []int{7, 0, 1, 2, 0, 3, 0, 4, 2, 3, 0, 3, 2, 1, 2, 0, 1, 7, 0, 1}

var memoryFrames = []Option{{"frames", "frames", 3, 1, memory.MaxFrames}}

func memoryMk(id, name, fn, summary string, run func([]int, int) (memory.Result, error)) Entry {
	e := entry(id, name, "os/memory/memory.go", fn)
	e.Summary, e.Time, e.Space, e.Input, e.Options, e.Traced = summary, "O(n * frames)", "O(n * frames)", Pages, memoryFrames, true
	e.Example = &Input{Values: memoryRefs, Options: map[string]int{"frames": 3}}
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

func memoryFit(id, name, fn, summary string, run func(parts, procs []int) (memory.Allocation, error)) Entry {
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

func fifoPageReplacementEntry() Entry {
	return memoryMk("fifo-page-replacement", "FIFO Page Replacement", "FIFO", "Evict the page loaded longest ago; adding frames can add faults (Belady's anomaly).", memory.FIFO)
}

func lruPageReplacementEntry() Entry {
	return memoryMk("lru-page-replacement", "LRU Page Replacement", "LRU", "Evict the page used longest ago: the recent past predicts the near future.", memory.LRU)
}

func optimalPageReplacementEntry() Entry {
	return memoryMk("optimal-page-replacement", "Optimal Page Replacement", "Optimal", "Evict the page used furthest in the future: impossible to run for real, the fewest faults possible.", memory.Optimal)
}

func clockPageReplacementEntry() Entry {
	return memoryMk("clock-page-replacement", "Clock (Second Chance)", "Clock", "A hand sweeps the frames clearing reference bits, evicting the first page not used since the last sweep.", memory.Clock)
}

func firstFitEntry() Entry {
	return memoryFit("first-fit", "First Fit", "FirstFit", "Place each process in the first free partition big enough.", memory.FirstFit)
}

func bestFitEntry() Entry {
	return memoryFit("best-fit", "Best Fit", "BestFit", "Place each process in the smallest free partition big enough.", memory.BestFit)
}

func worstFitEntry() Entry {
	return memoryFit("worst-fit", "Worst Fit", "WorstFit", "Place each process in the largest free partition, so the leftover stays useful.", memory.WorstFit)
}
