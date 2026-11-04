package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/braswelljr/alGOrithm/internal/core"
)

// printResult writes a run for a person to read: the result in the shape it
// has, then the counts, the time, and with trace every event.
func printResult(out io.Writer, e core.Entry, in core.Input, res core.Result, trace bool) error {
	fmt.Fprintf(out, "%s (time %s, space %s)\n", e.Name, e.Time, e.Space)
	o := res.Output
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	r := shapeRun{out: out, w: w, e: e, in: in, o: o}
	for _, sh := range shapes {
		if sh.is(r) {
			sh.print(r)
			break
		}
	}
	printExtras(w, o)
	printCounts(w, res)
	if err := w.Flush(); err != nil {
		return err
	}
	if trace {
		printEvents(out, res)
	}
	return nil
}

// shapeRun is what a shape printer reads: where to write, the entry, the
// input and its output.
type shapeRun struct {
	out io.Writer
	w   *tabwriter.Writer
	e   core.Entry
	in  core.Input
	o   core.Output
}

// shapes print an output by the first shape it has, in this order; a new
// output shape is a new row.
var shapes = []struct {
	is    func(shapeRun) bool
	print func(shapeRun)
}{
	{func(r shapeRun) bool { return r.o.Grid != nil }, printGridResult},
	{func(r shapeRun) bool { return r.o.Timeline != nil }, func(r shapeRun) {
		drawTimeline(r.out, r.o.Timeline)
		if r.o.Table != nil {
			drawTable(r.out, r.o.Table)
		}
	}},
	{func(r shapeRun) bool { return r.o.Track != nil }, func(r shapeRun) {
		fmt.Fprintf(r.w, "head path\t%s\n", strings.Trim(fmt.Sprint(r.o.Track.Path), "[]"))
		fmt.Fprintf(r.w, "served\t%s\n", strings.Trim(fmt.Sprint(r.o.Values), "[]"))
	}},
	{func(r shapeRun) bool { return r.o.Table != nil }, printTableResult},
	{func(r shapeRun) bool { return r.o.Matches != nil }, func(r shapeRun) {
		if !strings.ContainsAny(r.in.Other, " ,") { // one pattern: bracket its matches
			fmt.Fprintf(r.w, "text\t%s\n", markMatches(r.o.Text, r.o.Matches, len(r.in.Other)))
		}
		fmt.Fprintf(r.w, "result\t%d matches at %v\n", len(r.o.Matches), r.o.Matches)
	}},
	{func(r shapeRun) bool { return r.e.Category == "sorting" }, func(r shapeRun) {
		fmt.Fprintf(r.w, "input\t%v\nsorted\t%v\n", r.o.Initial, r.o.Values)
	}},
	{func(r shapeRun) bool { return r.o.Index != nil }, printIndexResult},
	{func(r shapeRun) bool { return len(r.o.Steps) > 0 }, printStepsResult},
	{func(r shapeRun) bool { return r.o.Initial != nil && r.o.Values != nil }, func(r shapeRun) {
		fmt.Fprintf(r.w, "input\t%v\nresult\t%v\n", r.o.Initial, r.o.Values)
		if r.o.Number != nil {
			fmt.Fprintf(r.w, "value\t%d\n", *r.o.Number)
		}
	}},
	{func(r shapeRun) bool { return r.o.Lines != nil }, func(shapeRun) {}}, // printed with the extras
	{func(r shapeRun) bool { return r.o.Number != nil }, func(r shapeRun) { fmt.Fprintf(r.w, "result\t%d\n", *r.o.Number) }},
	{func(r shapeRun) bool { return r.o.Bool != nil }, func(r shapeRun) { fmt.Fprintf(r.w, "result\t%v\n", *r.o.Bool) }},
	{func(r shapeRun) bool { return r.o.Values != nil }, func(r shapeRun) { fmt.Fprintf(r.w, "result\t%v\n", r.o.Values) }},
	{func(shapeRun) bool { return true }, func(r shapeRun) { fmt.Fprintf(r.w, "result\t%s\n", r.o.Text) }},
}

func printGridResult(r shapeRun) {
	drawGrid(r.out, r.o.Grid, r.o.Path)
	if r.o.Path == nil {
		fmt.Fprintln(r.w, "result\tno path from S to E")
		return
	}
	fmt.Fprintf(r.w, "result\t%d steps, cost %d\n", len(r.o.Path)-1, *r.o.Number)
}

func printTableResult(r shapeRun) {
	drawTable(r.out, r.o.Table)
	o := r.o
	switch {
	case o.Text != "" && o.Number != nil:
		fmt.Fprintf(r.w, "result\t%q (length %d)\n", o.Text, *o.Number)
	case o.Values != nil && o.Number != nil:
		fmt.Fprintf(r.w, "result\t%d: %v\n", *o.Number, o.Values)
	case o.Number != nil:
		fmt.Fprintf(r.w, "result\t%d\n", *o.Number)
	case o.Text != "":
		fmt.Fprintf(r.w, "result\t%s\n", o.Text)
	}
}

func printIndexResult(r shapeRun) {
	fmt.Fprintf(r.w, "searched\t%v\n", r.o.Initial)
	if *r.o.Index < 0 {
		fmt.Fprintf(r.w, "result\t%d is not there (-1)\n", r.in.Target)
		return
	}
	fmt.Fprintf(r.w, "result\t%d is at index %d\n", r.in.Target, *r.o.Index)
}

func printStepsResult(r shapeRun) {
	for i, st := range r.o.Steps {
		label := strings.TrimSpace(st.Op.Name + " " + st.Op.Value)
		result := st.Result
		if st.Error != "" {
			result = "error: " + st.Error
		}
		fmt.Fprintf(r.w, "%d. %s\t%s\t%s\n", i, label, result, snapshotLine(st))
	}
}

// printExtras writes what any shape can carry besides its main result: the
// lines, a final number or answer, and the named summary figures.
func printExtras(w io.Writer, o core.Output) {
	for i, l := range o.Lines {
		fmt.Fprintf(w, "%d.\t%s\n", i+1, l)
	}
	if len(o.Steps) > 0 && o.Number != nil {
		fmt.Fprintf(w, "result\t%d\n", *o.Number)
	}
	if o.Bool != nil && (o.Lines != nil || o.Timeline != nil || len(o.Steps) > 0) {
		fmt.Fprintf(w, "result\t%v\n", *o.Bool)
	}
	keys := make([]string, 0, len(o.Summary))
	for k := range o.Summary {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(w, "%s\t%g\n", k, o.Summary[k])
	}
}

// printCounts writes the event counts and the untraced run's time.
func printCounts(w io.Writer, res core.Result) {
	var parts []string
	for _, k := range []string{"compare", "swap", "set", "visit", "push", "pop"} {
		if n := res.Counts[k]; n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", n, k))
		}
	}
	if len(parts) > 0 {
		fmt.Fprintf(w, "counts\t%s\n", strings.Join(parts, ", "))
	}
	fmt.Fprintf(w, "time\t%v\n", time.Duration(res.Nanos))
}

// printEvents writes every event kept, one a line.
func printEvents(out io.Writer, res core.Result) {
	for i, ev := range res.Events {
		fmt.Fprintf(out, "%6d  %-7s %v %v %s\n", i+1, ev.Kind, ev.Indices, ev.Values, ev.Note)
	}
	if res.Truncated {
		fmt.Fprintf(out, "(%d events kept; raise --limit to see more)\n", len(res.Events))
	}
}

func snapshotLine(st core.Step) string {
	labels := make([]string, len(st.Snapshot.Nodes))
	for i, n := range st.Snapshot.Nodes {
		labels[i] = n.Label
	}
	return "[" + strings.Join(labels, " | ") + "]"
}

// drawGrid prints the grid with the path marked "*".
func drawGrid(w io.Writer, g *core.GridOutput, path []int) {
	on := map[int]bool{}
	for _, c := range path {
		on[c] = true
	}
	var b strings.Builder
	for i, cost := range g.Cost {
		switch {
		case i == g.Start:
			b.WriteByte('S')
		case i == g.End:
			b.WriteByte('E')
		case on[i]:
			b.WriteByte('*')
		case cost == 0:
			b.WriteByte('#')
		case cost == 1:
			b.WriteByte('.')
		default:
			b.WriteByte(byte('0' + cost))
		}
		if i%g.Cols == g.Cols-1 {
			b.WriteByte('\n')
		}
	}
	fmt.Fprint(w, b.String())
}

// drawTimeline prints a Gantt chart, one line per lane, one column per tick
// (up to 100 ticks): "#" running, "." idle, "-" blocked; in a timing diagram,
// "#" high and "_" low.
func drawTimeline(w io.Writer, spans []core.Span) {
	end, lanes, order := 0, map[string][]core.Span{}, []string{}
	for _, s := range spans {
		if _, ok := lanes[s.Lane]; !ok {
			order = append(order, s.Lane)
		}
		lanes[s.Lane] = append(lanes[s.Lane], s)
		end = max(end, s.End)
	}
	if end > 100 {
		fmt.Fprintf(w, "(a timeline of %d ticks; --json prints it)\n", end)
		return
	}
	width := 0
	for _, l := range order {
		width = max(width, len(l))
	}
	for _, l := range order {
		row := []byte(strings.Repeat(" ", end))
		for _, s := range lanes[l] {
			c := byte('#')
			switch s.State {
			case "blocked":
				c = '-'
			case "idle":
				c = '.'
			case "low": // a timing diagram: low is the baseline, high is "#"
				c = '_'
			}
			for t := s.Start; t < s.End; t++ {
				row[t] = c
			}
		}
		fmt.Fprintf(w, "%-*s |%s|\n", width, l, row)
	}
	fmt.Fprintf(w, "%-*s  0%s%d\n", width, "", strings.Repeat(" ", max(0, end-1-len(fmt.Sprint(end)))), end)
}

// drawTable prints a table with its labels.
func drawTable(w io.Writer, t *core.TableOutput) {
	if t.Cols > 40 || t.Rows > 40 {
		fmt.Fprintf(w, "(a %dx%d table; --json prints it)\n", t.Rows, t.Cols)
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, 1, ' ', tabwriter.AlignRight)
	if len(t.ColLabels) > 0 {
		if len(t.RowLabels) > 0 {
			fmt.Fprint(tw, "\t")
		}
		for _, l := range t.ColLabels {
			fmt.Fprintf(tw, "%s\t", l)
		}
		fmt.Fprintln(tw)
	}
	for r := 0; r < t.Rows; r++ {
		if len(t.RowLabels) > 0 {
			fmt.Fprintf(tw, "%s\t", t.RowLabels[r])
		}
		for c := 0; c < t.Cols; c++ {
			v := t.Cells[r*t.Cols+c]
			if v < 0 {
				fmt.Fprint(tw, "-\t")
			} else {
				fmt.Fprintf(tw, "%d\t", v)
			}
		}
		fmt.Fprintln(tw)
	}
	_ = tw.Flush()
}

// markMatches prints text with each match wrapped in [brackets].
func markMatches(text string, matches []int, length int) string {
	var b strings.Builder
	end := -1
	starts := map[int]bool{}
	for _, m := range matches {
		starts[m] = true
	}
	for i := 0; i < len(text); i++ {
		if starts[i] && i > end {
			b.WriteByte('[')
			end = i + length - 1
		}
		b.WriteByte(text[i])
		if i == end {
			b.WriteByte(']')
		}
	}
	return b.String()
}
