package core

import (
	"fmt"
	"strconv"

	"github.com/braswelljr/alGOrithm/pkg/os/cache"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// cacheEntries is every entry of the cache category.
func cacheEntries() []Entry {
	return []Entry{
		amatEntry(),
		cacheSimulatorEntry(),
	}
}

func cacheConfig(in Input) cache.Config {
	c := cache.Config{Lines: in.opt("lines", 8), BlockSize: in.opt("block", 4), Ways: in.opt("ways", 1), Policy: cache.LRU}
	if in.opt("fifo", 0) == 1 {
		c.Policy = cache.FIFO
	}
	return c
}

func cacheSimulatorEntry() Entry {
	sim := entry("cache-simulator", "CPU Cache Simulator", "os/cache/cache.go", "Simulate")
	sim.Summary, sim.Time, sim.Space, sim.Input, sim.Traced = "Run an address trace through a direct-mapped, set-associative or fully associative cache, and see every hit, miss and eviction.", "O(n * ways)", "O(lines)", Ints, true
	sim.Options = []Option{{"lines", "cache lines", 8, 1, 1024}, {"block", "block size in bytes", 4, 1, 4096}, {"ways", "ways (1 direct-mapped, lines fully associative)", 1, 1, 1024}, {"fifo", "FIFO replacement (1) instead of LRU (0)", 0, 0, 1}}
	sim.Example = &Input{Values: []int{0, 4, 0, 8, 32, 0, 4, 12, 32, 0, 36, 4}, Options: map[string]int{"lines": 4, "block": 4, "ways": 1}}
	sim.Limits = fmt.Sprintf("at most %d addresses", cache.MaxAccesses)
	sim.Run = func(in Input, r trace.Recorder) (Output, error) {
		c := cacheConfig(in)
		res, err := cache.Simulate(c, in.Values)
		if err != nil {
			return Output{}, wrap(err)
		}
		sets, n := c.Sets(), len(res.Accesses)
		t := &TableOutput{Rows: c.Lines, Cols: n, Cells: make([]int, c.Lines*n)}
		state := make([]int, c.Lines) // the tag in each line (set*ways + way) as the trace runs
		for i := range state {
			state[i] = -1
		}
		out := Output{Table: t}
		for col, a := range res.Accesses {
			row := a.Set*c.Ways + a.Way
			state[row] = a.Tag
			for k, tag := range state {
				t.Cells[k*n+col] = tag
			}
			r.Record(trace.Event{Kind: trace.Set, Indices: []int{row*n + col}, Values: []int{a.Tag}})
			note := "fault"
			if a.Result == cache.Hit {
				note = "hit"
			}
			r.Record(trace.Event{Kind: trace.Mark, Indices: []int{row*n + col}, Note: note})
			t.ColLabels = append(t.ColLabels, strconv.Itoa(a.Addr))
			line := fmt.Sprintf("address %d: tag %d, set %d, offset %d: %s", a.Addr, a.Tag, a.Set, a.Offset, a.Result)
			if a.Result != cache.Hit {
				line += " miss"
				if a.Evicted >= 0 {
					line += fmt.Sprintf(", evicting tag %d", a.Evicted)
				}
			}
			out.Lines = append(out.Lines, line)
		}
		for s := 0; s < sets; s++ {
			for w := 0; w < c.Ways; w++ {
				t.RowLabels = append(t.RowLabels, fmt.Sprintf("set %d way %d", s, w))
			}
		}
		out.Number = ptr(res.Hits)
		out.Summary = map[string]float64{"hits": float64(res.Hits), "misses": float64(res.Misses), "hit rate": res.HitRate(),
			"compulsory misses": float64(res.ByKind[cache.Compulsory]), "capacity misses": float64(res.ByKind[cache.Capacity]), "conflict misses": float64(res.ByKind[cache.Conflict])}
		return out, nil
	}
	sim.Call = func(in Input) (string, string) {
		c := cacheConfig(in)
		return "", fmt.Sprintf("cache.Simulate(cache.Config{Lines: %d, BlockSize: %d, Ways: %d, Policy: %q}, %s)", c.Lines, c.BlockSize, c.Ways, c.Policy, ints(in.Values))
	}
	return sim
}

func amatEntry() Entry {
	amat := entry("amat", "Average Memory Access Time", "os/cache/cache.go", "AMAT")
	amat.Summary, amat.Time, amat.Space, amat.Input = "AMAT = hit time + miss rate x miss penalty, for one cache level or two.", "O(1)", "O(1)", Simulation
	amat.Options = []Option{{"hit", "L1 hit time (cycles)", 1, 0, 1000}, {"miss", "L1 miss rate (percent)", 5, 0, 100}, {"penalty", "L1 miss penalty, or L2 hit time with two levels (cycles)", 100, 0, 100_000},
		{"l2miss", "L2 miss rate (percent; 0 for one level)", 0, 0, 100}, {"memory", "memory access time with two levels (cycles)", 200, 0, 100_000}}
	amat.Example = &Input{Options: map[string]int{"hit": 1, "miss": 5, "penalty": 10, "l2miss": 20, "memory": 100}}
	amat.Run = func(in Input, _ trace.Recorder) (Output, error) {
		hit, miss, pen := float64(in.opt("hit", 1)), float64(in.opt("miss", 5))/100, float64(in.opt("penalty", 100))
		if l2 := float64(in.opt("l2miss", 0)) / 100; l2 > 0 {
			l2amat := cache.AMAT(pen, l2, float64(in.opt("memory", 200)))
			total := cache.AMAT(hit, miss, l2amat)
			return Output{Summary: map[string]float64{"AMAT": total, "L2 AMAT": l2amat}, Lines: []string{
				fmt.Sprintf("L2: %g + %g x %d = %g cycles", pen, l2, in.opt("memory", 200), l2amat),
				fmt.Sprintf("L1: %g + %g x %g = %g cycles", hit, miss, l2amat, total),
			}}, nil
		}
		total := cache.AMAT(hit, miss, pen)
		return Output{Summary: map[string]float64{"AMAT": total}, Lines: []string{fmt.Sprintf("%g + %g x %g = %g cycles", hit, miss, pen, total)}}, nil
	}
	amat.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("cache.AMAT(%d, %g, %d)", in.opt("hit", 1), float64(in.opt("miss", 5))/100, in.opt("penalty", 100))
	}
	return amat
}
