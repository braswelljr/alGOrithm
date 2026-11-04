package cli

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	algorithm "github.com/braswelljr/alGOrithm"
	"github.com/braswelljr/alGOrithm/internal/core"
	"github.com/braswelljr/alGOrithm/internal/version"
	"github.com/braswelljr/alGOrithm/pkg/dataset"
)

func lookupArg(rest []string) (core.Entry, error) {
	if len(rest) == 0 {
		return core.Entry{}, usagef("name an entry (see `%s list`)", version.AppName)
	}
	e, ok := core.Lookup(rest[0])
	if !ok {
		return core.Entry{}, usagef("no entry %q (see `%s list`)", rest[0], version.AppName)
	}
	return e, nil
}

func runList(args []string, s Streams) error {
	o := newOptions("list")
	rest, err := o.parse(args)
	if err != nil {
		return err
	}
	entries := core.Entries()
	if len(rest) > 0 {
		if entries = core.InCategory(rest[0]); len(entries) == 0 {
			ids := make([]string, len(core.Categories))
			for i, c := range core.Categories {
				ids[i] = c.ID
			}
			return usagef("unknown category %q (one of %s)", rest[0], strings.Join(ids, ", "))
		}
	}
	entries = filter(entries, o)
	if o.asJSON {
		return writeJSON(s.Out, entries)
	}
	w := tabwriter.NewWriter(s.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tCATEGORY\tLEVEL\tTIME")
	for _, e := range entries {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", e.ID, e.Name, e.Category, e.Level, e.Time)
	}
	return w.Flush()
}

func runInfo(args []string, s Streams) error {
	o := newOptions("info")
	rest, err := o.parse(args)
	if err != nil {
		return err
	}
	e, err := lookupArg(rest)
	if err != nil {
		return err
	}
	if o.asJSON {
		return writeJSON(s.Out, e)
	}
	w := tabwriter.NewWriter(s.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "%s\t%s\n", e.Name, e.Summary)
	for _, kv := range [][2]string{{"id", e.ID}, {"kind", e.Kind}, {"level", e.Level}, {"tags", strings.Join(e.Tags, ", ")}, {"category", e.Category}, {"go", e.Import + "." + e.Func}, {"source", e.File},
		{"time", e.Time}, {"space", e.Space}, {"input", string(e.Input)}, {"animated", strconv.FormatBool(e.Traced)}} {
		fmt.Fprintf(w, "%s\t%s\n", kv[0], kv[1])
	}
	if e.Why != "" {
		fmt.Fprintf(w, "why\t%s\n", e.Why)
	}
	if n := e.Notes; n != nil {
		fmt.Fprintf(w, "tl;dr\t%s\n", n.TLDR)
		fmt.Fprintf(w, "key idea\t%s\n", n.KeyIdea)
		fmt.Fprintf(w, "use it for\t%s\n", n.WhenToUse)
		fmt.Fprintf(w, "pitfall\t%s\n", n.Pitfall)
	}
	for _, u := range e.RealWorld {
		fmt.Fprintf(w, "used in\t%s: %s\n", u.Title, u.Where)
		if u.Example != "" {
			fmt.Fprintf(w, "example\tgo run ./%s\n", u.Example)
		}
	}
	if e.Traced && core.Measurable(e) {
		fmt.Fprintf(w, "measure\talgorithm bench %s\n", e.ID)
	}
	if len(e.Ops) > 0 {
		fmt.Fprintf(w, "operations\t%s\n", strings.Join(e.Ops, ", "))
	}
	for _, op := range e.Options {
		fmt.Fprintf(w, "--option %s=N\t%s (default %d, %d to %d)\n", op.Name, op.Label, op.Default, op.Min, op.Max)
	}
	if e.Limits != "" {
		fmt.Fprintf(w, "limits\t%s\n", e.Limits)
	}
	fmt.Fprintf(w, "example\t%s\n", core.MakeSnippet(e, *e.Example).CLI)
	return w.Flush()
}

func runRun(args []string, s Streams) error {
	o := newOptions("run")
	rest, err := o.parse(args)
	if err != nil {
		return err
	}
	e, err := lookupArg(rest)
	if err != nil {
		return err
	}
	in, err := o.input(e, rest[1:], s)
	if err != nil {
		return err
	}
	limit := o.limit
	if !o.trace && !o.asJSON {
		limit = -1 // nothing prints the events; keep only the counts
	}
	res, err := core.Execute(e.ID, in, limit)
	if err != nil {
		return err
	}
	if o.asJSON {
		return writeJSON(s.Out, res)
	}
	return printResult(s.Out, e, in, res, o.trace)
}

func runCompare(args []string, s Streams) error {
	o := newOptions("compare")
	rest, err := o.parse(args)
	if err != nil {
		return err
	}
	var ids, operands []string
	for _, a := range rest {
		if _, ok := core.Lookup(a); ok {
			ids = append(ids, a)
		} else {
			operands = append(operands, a)
		}
	}
	chosen := len(ids) > 0
	if !chosen {
		for _, e := range core.InCategory("sorting") {
			ids = append(ids, e.ID)
		}
	}
	values, err := o.load(operands, s)
	if err != nil {
		return err
	}
	if values == nil {
		if values, err = dataset.Generate(dataset.Random, 1000, o.seed); err != nil {
			return err
		}
	}
	in := core.Input{Values: values, Target: o.target, Options: o.opts}
	var results []core.Result
	if chosen {
		// Entries the caller named: an input one of them refuses is an error.
		if results, err = core.Compare(ids, in); err != nil {
			return err
		}
	} else {
		// Every sort: one that refuses this input (Bogo Sort takes at most 8
		// values) is left out, and the reader is told.
		for _, id := range ids {
			r, err := core.Compare([]string{id}, in)
			if errors.Is(err, core.ErrInput) {
				fmt.Fprintf(s.Err, "algorithm: skipping %s: %v\n", id, err)
				continue
			}
			if err != nil {
				return err
			}
			results = append(results, r...)
		}
	}
	if o.asJSON {
		return writeJSON(s.Out, results)
	}
	fmt.Fprintf(s.Out, "%d values\n", len(values))
	w := tabwriter.NewWriter(s.Out, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(w, "ID\tCOMPARES\tSWAPS\tWRITES\tTIME\t")
	for _, r := range results {
		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%v\t\n", r.ID, r.Counts["compare"], r.Counts["swap"], r.Counts["set"], time.Duration(r.Nanos))
	}
	return w.Flush()
}

func runSource(args []string, s Streams) error {
	o := newOptions("source")
	rest, err := o.parse(args)
	if err != nil {
		return err
	}
	e, err := lookupArg(rest)
	if err != nil {
		return err
	}
	code, _ := algorithm.File(e.File)
	if o.asJSON {
		return writeJSON(s.Out, map[string]string{"file": e.File, "code": code, "doc": algorithm.Doc(e.File, e.Func), "packageDoc": algorithm.PackageDoc(e.Package)})
	}
	_, err = fmt.Fprintf(s.Out, "// %s (%s.%s)\n\n%s", e.File, e.Import, e.Func, code)
	return err
}

func runSnippet(args []string, s Streams) error {
	o := newOptions("snippet")
	rest, err := o.parse(args)
	if err != nil {
		return err
	}
	e, err := lookupArg(rest)
	if err != nil {
		return err
	}
	in := *e.Example
	if !o.example && len(rest) > 1 || o.values != "" || o.file != "" || o.generate != "" || o.text != "" {
		if in, err = o.input(e, rest[1:], s); err != nil {
			return err
		}
	}
	sn := core.MakeSnippet(e, in)
	if o.asJSON {
		return writeJSON(s.Out, sn)
	}
	if sn.Go != "" {
		fmt.Fprintf(s.Out, "%s\n", sn.Go)
	}
	_, err = fmt.Fprintf(s.Out, "# the same from the command line\n%s\n", sn.CLI)
	return err
}

func runData(args []string, s Streams) error {
	o := newOptions("data")
	rest, err := o.parse(args)
	if err != nil {
		return err
	}
	values, err := o.load(rest, s)
	if err != nil {
		return err
	}
	if values == nil {
		return usagef("give data: values, --values, --file or --generate")
	}
	if o.asJSON {
		return writeJSON(s.Out, values)
	}
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = strconv.Itoa(v)
	}
	_, err = fmt.Fprintln(s.Out, strings.Join(parts, " "))
	return err
}

// filter keeps the entries matching --tag, --kind and --level.
func filter(entries []core.Entry, o *options) []core.Entry {
	var out []core.Entry
	for _, e := range entries {
		if (o.tag == "" || slices.Contains(e.Tags, strings.ToLower(o.tag))) &&
			(o.kind == "" || strings.EqualFold(e.Kind, o.kind)) &&
			(o.level == "" || strings.EqualFold(e.Level, o.level)) {
			out = append(out, e)
		}
	}
	return out
}

func runTags(args []string, s Streams) error {
	o := newOptions("tags")
	if _, err := o.parse(args); err != nil {
		return err
	}
	count := map[string]int{}
	for _, e := range core.Entries() {
		for _, t := range e.Tags {
			count[t]++
		}
	}
	tags := make([]string, 0, len(core.Vocabulary))
	for t := range core.Vocabulary {
		tags = append(tags, t)
	}
	slices.Sort(tags)
	if o.asJSON {
		out := make([]map[string]any, len(tags))
		for i, t := range tags {
			out[i] = map[string]any{"tag": t, "meaning": core.Vocabulary[t], "entries": count[t]}
		}
		return writeJSON(s.Out, out)
	}
	w := tabwriter.NewWriter(s.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TAG\tENTRIES\tMEANING")
	for _, t := range tags {
		fmt.Fprintf(w, "%s\t%d\t%s\n", t, count[t], core.Vocabulary[t])
	}
	return w.Flush()
}
