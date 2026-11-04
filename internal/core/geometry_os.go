package core

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/math/geometry"
	"github.com/braswelljr/alGOrithm/pkg/os/cache"
	"github.com/braswelljr/alGOrithm/pkg/os/concurrency"
	"github.com/braswelljr/alGOrithm/pkg/os/memory"
	"github.com/braswelljr/alGOrithm/pkg/os/scheduling"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// toPoints converts geometry points for the visualizer.
func toPoints(pts []geometry.Point) []Point {
	out := make([]Point, len(pts))
	for i, p := range pts {
		out[i] = Point{X: p.X, Y: p.Y}
	}
	return out
}

func parsePoints(in Input) ([]geometry.Point, error) {
	pts, err := geometry.ParsePoints(in.Text)
	return pts, wrap(err)
}

// pointsCall is the snippet setup that parses the input's points.
func pointsCall(in Input, call string) (string, string) {
	return fmt.Sprintf("pts, err := geometry.ParsePoints(%s)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", "`"+in.Text+"`"), call
}

// segmentRecorder turns a segment trace's segment indexes into the indexes of
// the segments' two points, which is what the visualizer draws.
type segmentRecorder struct{ r trace.Recorder }

func (s segmentRecorder) Record(e trace.Event) {
	idx := make([]int, 0, 2*len(e.Indices))
	for _, i := range e.Indices {
		idx = append(idx, 2*i, 2*i+1)
	}
	e.Indices = idx
	s.r.Record(e)
}

const pointsExample = "0 0\n7 1\n3 3\n9 4\n5 6\n1 8\n8 9\n4 9\n2 5\n6 3\n10 7\n5 0"

func geometryEntries() []Entry {
	ex := &Input{Text: pointsExample}
	hull := func(id, name, file, fn, summary, time string, traced func([]geometry.Point, trace.Recorder) []int) Entry {
		e := entry(id, name, "math/geometry/"+file, fn)
		e.Summary, e.Time, e.Space, e.Input, e.Traced, e.Example = summary, time, "O(n)", Points, true, ex
		e.Limits = fmt.Sprintf("at most %d points", geometry.MaxPoints)
		e.Run = func(in Input, r trace.Recorder) (Output, error) {
			pts, err := parsePoints(in)
			if err != nil {
				return Output{}, err
			}
			h := traced(pts, r)
			hp := make([]geometry.Point, len(h))
			for i, k := range h {
				hp[i] = pts[k]
			}
			area := float64(geometry.PolygonArea(hp)) / 2
			return Output{Points: toPoints(pts), Values: h, Number: ptr(len(h)), Summary: map[string]float64{"hull points": float64(len(h)), "hull area": area}}, nil
		}
		e.Call = func(in Input) (string, string) { return pointsCall(in, "geometry."+fn+"(pts)") }
		return e
	}

	closest := entry("closest-pair", "Closest Pair of Points", "math/geometry/closest.go", "ClosestPair")
	closest.Summary, closest.Time, closest.Space, closest.Input, closest.Traced, closest.Example = "The two points nearest each other, by divide and conquer: solve each half, then check a narrow strip.", "O(n log^2 n)", "O(n)", Points, true, ex
	closest.Run = func(in Input, r trace.Recorder) (Output, error) {
		pts, err := parsePoints(in)
		if err != nil {
			return Output{}, err
		}
		if len(pts) < 2 {
			return Output{}, inputError("give at least two points")
		}
		i, j, d := geometry.TraceClosestPair(pts, r)
		return Output{Points: toPoints(pts), Values: []int{i, j},
			Lines:   []string{fmt.Sprintf("points %d (%d, %d) and %d (%d, %d)", i, pts[i].X, pts[i].Y, j, pts[j].X, pts[j].Y)},
			Summary: map[string]float64{"distance": math.Sqrt(float64(d)), "squared distance": float64(d)}}, nil
	}
	closest.Call = func(in Input) (string, string) { return pointsCall(in, "geometry.ClosestPair(pts)") }

	segs := entry("segment-intersection", "Segment Intersection", "math/geometry/segments.go", "SegmentsIntersect")
	segs.Summary, segs.Time, segs.Space, segs.Input, segs.Traced = "Which segments cross or touch, tested pair by pair with orientation (cross product) tests.", "O(n^2)", "O(k)", Points, true
	segs.Example = &Input{Text: "0 0\n8 8\n0 8\n8 0\n2 6\n6 7\n9 1\n9 6\n1 1\n3 0"}
	segs.Run = func(in Input, r trace.Recorder) (Output, error) {
		pts, err := parsePoints(in)
		if err != nil {
			return Output{}, err
		}
		ss, err := geometry.Segments(pts)
		if err != nil {
			return Output{}, wrap(err)
		}
		pairs := geometry.TraceSegmentsIntersect(ss, segmentRecorder{r})
		out := Output{Points: toPoints(pts), Number: ptr(len(pairs))}
		for i := 0; i+1 < len(pts); i += 2 {
			out.Segments = append(out.Segments, [2]int{i, i + 1})
		}
		for _, p := range pairs {
			out.Lines = append(out.Lines, fmt.Sprintf("segment %d and segment %d intersect", p[0], p[1]))
		}
		if len(pairs) == 0 {
			out.Lines = []string{"no two segments intersect"}
		}
		return out, nil
	}
	segs.Call = func(in Input) (string, string) {
		s, _ := pointsCall(in, "")
		return s + "segs, err := geometry.Segments(pts)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", "geometry.SegmentsIntersect(segs)"
	}

	pip := entry("point-in-polygon", "Point in Polygon", "math/geometry/polygon.go", "PointInPolygon")
	pip.Summary, pip.Time, pip.Space, pip.Input, pip.Traced = "Whether points lie inside a polygon, by casting a ray to the right and counting the edges it crosses.", "O(n) per point", "O(1)", Points, true
	pip.Example = &Input{Text: "# the polygon, in order\n0 0\n8 0\n8 8\n4 3\n0 8\ntest\n2 2\n4 6\n6 5\n9 4\n4 1"}
	pip.Run = func(in Input, r trace.Recorder) (Output, error) {
		polyText, testText, ok := strings.Cut(in.Text, "test")
		if !ok {
			return Output{}, inputError(`give the polygon's points, a line "test", then the points to test`)
		}
		poly, err := geometry.ParsePoints(polyText)
		if err != nil {
			return Output{}, wrap(err)
		}
		test, err := geometry.ParsePoints(testText)
		if err != nil {
			return Output{}, wrap(err)
		}
		if len(poly) < 3 {
			return Output{}, inputError("a polygon needs at least three points")
		}
		outline := make([]int, len(poly))
		for i := range outline {
			outline[i] = i
		}
		r.Record(trace.Event{Kind: trace.Mark, Indices: outline, Note: "hull"}) // the polygon's outline
		in2 := geometry.TracePointsInPolygon(poly, test, len(poly), r)
		out := Output{Points: toPoints(append(append([]geometry.Point(nil), poly...), test...))}
		inside := 0
		for i, b := range in2 {
			where := "outside"
			if b {
				where, inside = "inside", inside+1
			}
			out.Lines = append(out.Lines, fmt.Sprintf("(%d, %d) is %s", test[i].X, test[i].Y, where))
		}
		out.Number = &inside
		return out, nil
	}

	area := entry("polygon-area", "Polygon Area (shoelace)", "math/geometry/polygon.go", "PolygonArea")
	area.Summary, area.Time, area.Space, area.Input, area.Traced = "The area of a polygon from its vertices in order, exactly, with the shoelace formula.", "O(n)", "O(1)", Points, true
	area.Example = &Input{Text: "0 0\n8 0\n8 8\n4 3\n0 8"}
	area.Run = func(in Input, r trace.Recorder) (Output, error) {
		pts, err := parsePoints(in)
		if err != nil {
			return Output{}, err
		}
		if len(pts) < 3 {
			return Output{}, inputError("a polygon needs at least three points")
		}
		outline := make([]int, len(pts))
		for i := range outline {
			outline[i] = i
		}
		r.Record(trace.Event{Kind: trace.Mark, Indices: outline, Note: "hull"})
		twice := geometry.PolygonArea(pts)
		dir := "counterclockwise"
		if twice < 0 {
			dir = "clockwise"
		}
		return Output{Points: toPoints(pts), Lines: []string{"the vertices run " + dir},
			Summary: map[string]float64{"area": math.Abs(float64(twice)) / 2, "twice the signed area": float64(twice)}}, nil
	}
	area.Call = func(in Input) (string, string) { return pointsCall(in, "geometry.PolygonArea(pts)") }

	return []Entry{
		hull("convex-hull-graham", "Convex Hull (Graham Scan)", "hull.go", "ConvexHullGraham", "Sort the points by angle around the lowest one, then keep a stack that only turns left (Graham, 1972).", "O(n log n)", geometry.TraceConvexHullGraham),
		hull("convex-hull-jarvis", "Convex Hull (Jarvis March)", "hull.go", "ConvexHullJarvis", "Gift wrapping: from the lowest point, repeatedly take the point with every other to its left (Jarvis, 1973).", "O(n * h) for h hull points", geometry.TraceConvexHullJarvis),
		hull("convex-hull-monotone", "Convex Hull (Monotone Chain)", "hull.go", "ConvexHullMonotone", "Sort by x, then build the lower and upper hulls with a left-turn stack (Andrew, 1979).", "O(n log n)", geometry.TraceConvexHullMonotone),
		closest, segs, pip, area,
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

// spans turns a simulation timeline into the visualizer's lanes.
func spans(tl concurrency.Timeline, lane func(int) string) []Span {
	out := make([]Span, len(tl))
	for i, a := range tl {
		out[i] = Span{Lane: lane(a.Thread), Label: a.What, Start: a.Tick, End: a.Tick + 1, State: a.State}
	}
	return out
}

func thread(i int) string { return "thread " + strconv.Itoa(i) }

func moreOSEntries() []Entry {
	var out []Entry
	out = append(out, cacheEntries()...)
	out = append(out, realTimeEntries()...)
	out = append(out, moreMemoryEntries()...)
	out = append(out, classicEntries()...)
	return out
}

func cacheEntries() []Entry {
	sim := entry("cache-simulator", "CPU Cache Simulator", "os/cache/cache.go", "Simulate")
	sim.Summary, sim.Time, sim.Space, sim.Input, sim.Traced = "Run an address trace through a direct-mapped, set-associative or fully associative cache, and see every hit, miss and eviction.", "O(n * ways)", "O(lines)", Ints, true
	sim.Options = []Option{{"lines", "cache lines", 8, 1, 1024}, {"block", "block size in bytes", 4, 1, 4096}, {"ways", "ways (1 direct-mapped, lines fully associative)", 1, 1, 1024}, {"fifo", "FIFO replacement (1) instead of LRU (0)", 0, 0, 1}}
	sim.Example = &Input{Values: []int{0, 4, 0, 8, 32, 0, 4, 12, 32, 0, 36, 4}, Options: map[string]int{"lines": 4, "block": 4, "ways": 1}} // try ways 2: the conflicts between 0 and 32 go away
	sim.Limits = fmt.Sprintf("at most %d addresses", cache.MaxAccesses)
	config := func(in Input) cache.Config {
		c := cache.Config{Lines: in.opt("lines", 8), BlockSize: in.opt("block", 4), Ways: in.opt("ways", 1), Policy: cache.LRU}
		if in.opt("fifo", 0) == 1 {
			c.Policy = cache.FIFO
		}
		return c
	}
	sim.Run = func(in Input, r trace.Recorder) (Output, error) {
		c := config(in)
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
		c := config(in)
		return "", fmt.Sprintf("cache.Simulate(cache.Config{Lines: %d, BlockSize: %d, Ways: %d, Policy: %q}, %s)", c.Lines, c.BlockSize, c.Ways, c.Policy, ints(in.Values))
	}

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
	return []Entry{sim, amat}
}

const tasksExample = "name period burst\nT1 5 2\nT2 7 4"

func realTimeEntries() []Entry {
	lottery := entry("lottery-scheduling", "Lottery Scheduling", "os/scheduling/lottery.go", "Lottery")
	lottery.Summary, lottery.Time, lottery.Space, lottery.Input = "Each tick, draw a ticket at random: a process gets the CPU in proportion to its tickets (the priority column).", "O(T * n)", "O(T)", Processes
	lottery.Options = []Option{{"seed", "draw seed", 1, 0, 1_000_000}}
	lottery.Example = &Input{Text: "name arrival burst tickets\nA 0 6 6\nB 0 6 3\nC 0 6 1", Options: map[string]int{"seed": 1}}
	lottery.Run = func(in Input, _ trace.Recorder) (Output, error) {
		ps, err := scheduling.Parse(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		s, err := scheduling.Lottery(ps, int64(in.opt("seed", 1)))
		if err != nil {
			return Output{}, wrap(err)
		}
		return scheduleOutput(s), nil
	}
	lottery.Call = func(in Input) (string, string) {
		return fmt.Sprintf("ps, err := scheduling.Parse(%s)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", "`"+in.Text+"`"), fmt.Sprintf("scheduling.Lottery(ps, %d)", in.opt("seed", 1))
	}

	rt := func(id, name, fn, summary string, run func([]scheduling.Task) (scheduling.RealTime, error)) Entry {
		e := entry(id, name, "os/scheduling/realtime.go", fn)
		e.Summary, e.Time, e.Space, e.Input = summary, "O(H * n) for a hyperperiod of H", "O(H)", Processes
		e.Example = &Input{Text: tasksExample}
		e.Limits = "tasks as \"name period burst\" lines"
		e.Run = func(in Input, _ trace.Recorder) (Output, error) {
			ts, err := scheduling.ParseTasks(in.Text)
			if err != nil {
				return Output{}, wrap(err)
			}
			res, err := run(ts)
			if err != nil {
				return Output{}, wrap(err)
			}
			out := Output{Bool: ptr(len(res.Missed) == 0), Summary: map[string]float64{
				"utilization": res.Utilization, "Liu-Layland bound": res.Bound, "missed deadlines": float64(len(res.Missed)), "hyperperiod": float64(res.Horizon),
			}}
			for _, sl := range res.Timeline {
				lane, state := sl.Process, "running"
				if lane == "" {
					lane, state = "idle", "idle"
				}
				out.Timeline = append(out.Timeline, Span{Lane: lane, Label: lane, Start: sl.Start, End: sl.End, State: state})
			}
			for _, m := range res.Missed {
				out.Lines = append(out.Lines, fmt.Sprintf("%s's job released at %d missed its deadline %d with %d ticks left", m.Task, m.Release, m.Deadline, m.Left))
			}
			if len(res.Missed) == 0 {
				out.Lines = []string{"every deadline met"}
			}
			return out, nil
		}
		e.Call = func(in Input) (string, string) {
			return fmt.Sprintf("ts, err := scheduling.ParseTasks(%s)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", "`"+in.Text+"`"), "scheduling." + fn + "(ts)"
		}
		return e
	}
	return []Entry{
		lottery,
		rt("edf-scheduling", "Earliest Deadline First", "EDF", "Periodic real-time tasks: always run the job whose deadline is nearest; optimal on one CPU when utilization is at most 1.", scheduling.EDF),
		rt("rate-monotonic", "Rate-Monotonic Scheduling", "RateMonotonic", "Periodic real-time tasks with fixed priorities, shorter period first; safe below the Liu-Layland bound.", scheduling.RateMonotonic),
	}
}

type buddyP struct {
	b *memory.Buddy
	n int // names given to unnamed allocations
}

func (p *buddyP) Snapshot() trace.Snapshot { return p.b.Snapshot() }
func (p *buddyP) apply(op Op, r trace.Recorder) (string, error) {
	f := strings.Fields(op.Value)
	switch op.Name {
	case "alloc":
		if len(f) == 1 { // "alloc SIZE": a name is made up
			p.n++
			f = []string{"P" + strconv.Itoa(p.n), f[0]}
		}
		if len(f) != 2 {
			return "", inputError("alloc needs NAME SIZE (or SIZE)")
		}
		size, err := strconv.Atoi(f[1])
		if err != nil {
			return "", inputError(fmt.Sprintf("alloc needs a whole-number size, not %q", f[1]))
		}
		off, err := p.b.Alloc(f[0], size)
		if err != nil {
			return fmt.Sprintf("%s (%d bytes): %v", f[0], size, err), nil
		}
		r.Record(trace.Event{Kind: trace.Push, Values: []int{off}})
		return fmt.Sprintf("%s got %d bytes at %d (internal fragmentation now %d)", f[0], size, off, p.b.Fragmentation()), nil
	case "free":
		if len(f) != 1 {
			return "", inputError("free needs NAME")
		}
		if err := p.b.Free(f[0]); err != nil {
			return err.Error(), nil //nolint:nilerr // freeing a bad block is a step to show, not a failure of the run
		}
		r.Record(trace.Event{Kind: trace.Pop})
		return fmt.Sprintf("freed %s, merging buddies where both are free", f[0]), nil
	}
	return "", unknownOp(op, "alloc", "free")
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

func moreMemoryEntries() []Entry {
	refs := []int{7, 0, 1, 2, 0, 3, 0, 4, 2, 3, 0, 3, 2, 1, 2, 0, 1, 7, 0, 1}
	page := func(id, name, file, fn, summary string, run func([]int, int) (memory.Result, error)) Entry {
		e := entry(id, name, "os/memory/"+file, fn)
		e.Summary, e.Time, e.Space, e.Input, e.Traced = summary, "O(n * frames)", "O(n * frames)", Pages, true
		e.Options = []Option{{"frames", "frames", 3, 1, memory.MaxFrames}}
		e.Example = &Input{Values: refs, Options: map[string]int{"frames": 3}}
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

	return []Entry{
		page("lfu-page-replacement", "LFU Page Replacement", "frequency.go", "LFU", "Evict the page used least often since it was loaded; keeps popular pages, adapts slowly.", memory.LFU),
		page("aging-page-replacement", "Aging Page Replacement", "frequency.go", "Aging", "An 8-bit counter per frame, shifted right on every reference with the used frame's top bit set: LRU on a budget.", memory.Aging),
		buddy, tr,
	}
}

func classicEntries() []Entry {
	sim := func(id, name, fn, summary string, opts []Option, example map[string]int, run func(in Input) (Output, error), call func(in Input) string) Entry {
		e := entry(id, name, "os/concurrency/classics.go", fn)
		e.Summary, e.Time, e.Space, e.Input, e.Options = summary, "O(ticks)", "O(ticks)", Simulation, opts
		e.Example = &Input{Options: example}
		e.Run = func(in Input, _ trace.Recorder) (Output, error) {
			out, err := run(in)
			return out, wrap(err)
		}
		e.Call = func(in Input) (string, string) { return "", call(in) }
		return e
	}
	seed := Option{"seed", "schedule seed", 1, 0, 1_000_000}
	b := func(v bool) float64 {
		if v {
			return 1
		}
		return 0
	}
	return []Entry{
		sim("readers-writers", "Readers-Writers", "ReadersWriters", "Readers share the data, a writer needs it alone; with reader preference writers can starve.",
			[]Option{{"readers", "readers", 4, 1, 15}, {"writers", "writers", 1, 1, 15}, {"rounds", "rounds each", 3, 1, 50}, {"writersfirst", "writer preference (1) or reader preference (0)", 0, 0, 1}, seed},
			map[string]int{"readers": 4, "writers": 1, "rounds": 3, "writersfirst": 0, "seed": 3},
			func(in Input) (Output, error) {
				readers := in.opt("readers", 4)
				r, err := concurrency.ReadersWriters(readers, in.opt("writers", 1), in.opt("rounds", 3), in.opt("writersfirst", 0) == 1, int64(in.opt("seed", 1)))
				if err != nil {
					return Output{}, err
				}
				lane := func(i int) string {
					if i < readers {
						return "reader " + strconv.Itoa(i)
					}
					return "writer " + strconv.Itoa(i-readers)
				}
				return Output{Timeline: spans(r.Timeline, lane), Summary: map[string]float64{"longest writer wait": float64(r.MaxWriterWait), "longest reader wait": float64(r.MaxReaderWait), "most readers at once": float64(r.MaxReaders)}}, nil
			},
			func(in Input) string {
				return fmt.Sprintf("concurrency.ReadersWriters(%d, %d, %d, %v, %d)", in.opt("readers", 4), in.opt("writers", 1), in.opt("rounds", 3), in.opt("writersfirst", 0) == 1, in.opt("seed", 1))
			}),
		sim("sleeping-barber", "Sleeping Barber", "SleepingBarber", "A barber sleeps until a customer arrives; customers wait in a few chairs or leave when they are full.",
			[]Option{{"customers", "customers", 6, 1, 15}, {"chairs", "waiting chairs", 2, 0, 15}, seed},
			map[string]int{"customers": 6, "chairs": 2, "seed": 1},
			func(in Input) (Output, error) {
				r, err := concurrency.SleepingBarber(in.opt("customers", 6), in.opt("chairs", 2), int64(in.opt("seed", 1)))
				if err != nil {
					return Output{}, err
				}
				lane := func(i int) string {
					if i == 0 {
						return "barber"
					}
					return "customer " + strconv.Itoa(i)
				}
				return Output{Timeline: spans(r.Timeline, lane), Summary: map[string]float64{"served": float64(r.Served), "turned away": float64(r.TurnedAway)}}, nil
			},
			func(in Input) string {
				return fmt.Sprintf("concurrency.SleepingBarber(%d, %d, %d)", in.opt("customers", 6), in.opt("chairs", 2), in.opt("seed", 1))
			}),
		sim("peterson", "Peterson's Algorithm", "Peterson", "Mutual exclusion for two threads with only shared flags and a turn; a check-then-set version lets both in.",
			[]Option{{"rounds", "rounds each", 3, 1, 100}, {"broken", "check-then-set instead (1)", 0, 0, 1}, seed},
			map[string]int{"rounds": 3, "broken": 0, "seed": 1},
			func(in Input) (Output, error) {
				r, err := concurrency.Peterson(in.opt("rounds", 3), in.opt("broken", 0) == 1, int64(in.opt("seed", 1)))
				if err != nil {
					return Output{}, err
				}
				return Output{Timeline: spans(r.Timeline, thread), Bool: ptr(r.Violations == 0), Summary: map[string]float64{"both inside (ticks)": float64(r.Violations), "entries": float64(r.Entries)}}, nil
			},
			func(in Input) string {
				return fmt.Sprintf("concurrency.Peterson(%d, %v, %d)", in.opt("rounds", 3), in.opt("broken", 0) == 1, in.opt("seed", 1))
			}),
		sim("barrier", "Barrier", "Barrier", "Threads work, then wait at a barrier until all arrive, round after round: the fast wait for the slow.",
			[]Option{{"threads", "threads", 4, 1, 16}, {"rounds", "rounds", 3, 1, 50}, seed},
			map[string]int{"threads": 4, "rounds": 3, "seed": 1},
			func(in Input) (Output, error) {
				r, err := concurrency.Barrier(in.opt("threads", 4), in.opt("rounds", 3), int64(in.opt("seed", 1)))
				if err != nil {
					return Output{}, err
				}
				return Output{Timeline: spans(r.Timeline, thread), Summary: map[string]float64{"ticks waited": float64(r.Waited), "rounds": float64(r.Rounds)}}, nil
			},
			func(in Input) string {
				return fmt.Sprintf("concurrency.Barrier(%d, %d, %d)", in.opt("threads", 4), in.opt("rounds", 3), in.opt("seed", 1))
			}),
		sim("semaphore-simulation", "Counting Semaphore", "SemaphoreSim", "Threads wait (P) for one of a few slots, work, and signal (V); never more inside than slots.",
			[]Option{{"threads", "threads", 5, 1, 16}, {"slots", "slots", 2, 1, 16}, {"rounds", "rounds each", 2, 1, 50}, seed},
			map[string]int{"threads": 5, "slots": 2, "rounds": 2, "seed": 1},
			func(in Input) (Output, error) {
				r, err := concurrency.SemaphoreSim(in.opt("threads", 5), in.opt("slots", 2), in.opt("rounds", 2), int64(in.opt("seed", 1)))
				if err != nil {
					return Output{}, err
				}
				return Output{Timeline: spans(r.Timeline, thread), Bool: ptr(r.MaxInside <= in.opt("slots", 2)), Summary: map[string]float64{"most inside": float64(r.MaxInside), "ticks waited": float64(r.Waited), "ok": b(r.MaxInside <= in.opt("slots", 2))}}, nil
			},
			func(in Input) string {
				return fmt.Sprintf("concurrency.SemaphoreSim(%d, %d, %d, %d)", in.opt("threads", 5), in.opt("slots", 2), in.opt("rounds", 2), in.opt("seed", 1))
			}),
	}
}
