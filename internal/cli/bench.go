package cli

import (
	"flag"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/braswelljr/alGOrithm/internal/core"
)

// runBench measures an entry at growing sizes and prints the growth class
// that fits it, beside the cost the entry states.
func runBench(args []string, s Streams) error {
	fs := flag.NewFlagSet("bench", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	sizes := fs.String("sizes", "", "")
	order := fs.String("input", "random", "")
	seed := fs.Int64("seed", 1, "")
	asJSON := fs.Bool("json", false, "")
	var ids []string
	for len(args) > 0 {
		if err := fs.Parse(args); err != nil {
			return usagef("%v", err)
		}
		args = fs.Args()
		if len(args) > 0 {
			ids, args = append(ids, args[0]), args[1:]
		}
	}
	if len(ids) != 1 {
		return usagef("bench takes one entry ID: algorithm bench quick-sort [--sizes 100,200,400,800] [--input random|sorted|reversed]")
	}
	var ns []int
	if *sizes != "" {
		for _, f := range strings.FieldsFunc(*sizes, func(r rune) bool { return r == ',' || r == ' ' }) {
			n, err := strconv.Atoi(f)
			if err != nil {
				return usagef("--sizes takes numbers, not %q", f)
			}
			ns = append(ns, n)
		}
	}
	m, err := core.MeasureOrder(ids[0], *order, ns, *seed)
	if err != nil {
		return err
	}
	if *asJSON {
		return writeJSON(s.Out, m)
	}
	printMeasurement(s.Out, m)
	return nil
}

func printMeasurement(out io.Writer, m core.Measurement) {
	unit := "ops"
	if m.ByTime {
		unit = "time"
	}
	fmt.Fprintf(out, "%s on %s values\n", m.Name, m.Order)
	fmt.Fprintln(out)
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintf(w, "n\tops\tops / prev\ttime\t\n")
	for i, p := range m.Samples {
		ratio := "-"
		if i > 0 && m.Samples[i-1].Ops > 0 {
			ratio = fmt.Sprintf("x%.2f", float64(p.Ops)/float64(m.Samples[i-1].Ops))
		}
		fmt.Fprintf(w, "%d\t%d\t%s\t%s\t\n", p.N, p.Ops, ratio, time.Duration(p.Nanos).Round(time.Microsecond))
	}
	_ = w.Flush()
	fmt.Fprintf(out, "\n%s\n", chart(m))
	fmt.Fprintf(out, "fit      %s on %s, off by %.1f%% (%s)\n", m.Best, unit, 100*m.Error, m.Quality)
	if len(m.Fits) > 1 {
		next := m.Fits[1]
		fmt.Fprintf(out, "next     %s, off by %s\n", next.Class, percent(next.Error))
	}
	fmt.Fprintf(out, "stated   %s\n", m.Stated)
	switch {
	case m.Agrees == nil:
		fmt.Fprintf(out, "verdict  the stated cost uses a size this measurement holds fixed, so they cannot be compared\n")
	case *m.Agrees:
		fmt.Fprintf(out, "verdict  agrees: it grows no faster than stated\n")
	default:
		fmt.Fprintf(out, "verdict  grows faster than stated on this input; see `algorithm info %s`\n", m.ID)
	}
	if m.Note != "" {
		fmt.Fprintf(out, "note     %s\n", m.Note)
	}
	if m.ByTime {
		fmt.Fprintf(out, "note     the entry reports no events, so this is fitted on time, which varies between runs\n")
	}
}

func percent(e float64) string {
	if e < 0 {
		return "too much to count"
	}
	return fmt.Sprintf("%.1f%%", 100*e)
}

// chart draws the measured counts as bars, one row per size, scaled to the
// largest, with the fitted class's prediction marked by a '|'.
func chart(m core.Measurement) string {
	const width = 40
	y := func(p core.Sample) float64 {
		if m.ByTime {
			return float64(p.Nanos)
		}
		return float64(p.Ops)
	}
	top := 1.0
	for _, p := range m.Samples {
		top = math.Max(top, y(p))
	}
	var b strings.Builder
	for _, p := range m.Samples {
		bar := int(math.Round(width * y(p) / top))
		row := []byte(strings.Repeat("#", bar) + strings.Repeat(" ", width-bar+1))
		if !m.ByTime && len(m.Fits) > 0 {
			if at := int(math.Round(width * m.Fits[0].Predict(p.N) / top)); at >= 0 && at <= width {
				row[at] = '|'
			}
		}
		fmt.Fprintf(&b, "%8d %s\n", p.N, row)
	}
	b.WriteString("         # measured   | the fit\n")
	return b.String()
}
