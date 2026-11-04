package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/internal/core"
	"github.com/braswelljr/alGOrithm/pkg/algorithms/pathfinding"
	"github.com/braswelljr/alGOrithm/pkg/dataset"
)

const flagsHelp = `Data (run, compare, data, snippet):
  VALUE...               the values: algorithm run quick-sort 5 3 9 1
  --values LIST          the same as one argument: --values "5, 3, 9, 1"
  --file PATH            read from a file (- for standard input): .csv, .json, or text
  --format F             text, csv or json, when the file name does not say
  --column C             the CSV column, by number (1 is the first) or header name
  --no-header            the CSV file has no header row
  --skip-empty           skip empty CSV cells instead of reporting them
  --field NAME           the JSON field holding the values
  --round                accept decimals and round them
  --generate KIND        generate values (random, sorted, reversed, nearly-sorted, few-unique,
                         negative) or a grid (maze, walls, weighted, open)
  --size N, --seed N     how many values (default 20) and the seed (default 1)
  --rows N, --cols N     a generated grid's size (default 15x31)

Run:
  --target N, --amount N the value to search for, the amount for coin change
  --text TEXT            text input: a cipher's text, a grid, processes, a Banker's state
  --other TEXT           the second word, or the pattern (also --pattern)
  --key KEY, --shift N   a Vigenère keyword, a Caesar shift (default 3)
  --ops LIST             structure operations: --ops "push 4, push 7, pop"
  --option NAME=N        a numeric option (see info ID): --option quantum=2 --option frames=4
  --example              run the entry's example input
  --trace                print every event
  --limit N              keep at most N events (default 100000)
  --json                 print the result as JSON

Examples:
  algorithm run quick-sort --generate nearly-sorted --size 30 --trace
  algorithm run binary-search --file scores.csv --column score --target 90
  algorithm run a-star --generate maze --rows 15 --cols 31
  algorithm run round-robin --example --option quantum=3
  algorithm run lru-page-replacement 7 0 1 2 0 3 0 4 --option frames=3
  algorithm run race-condition --option threads=4 --option locked=0
  algorithm source dijkstra
`

// options are the flags every command shares.
type options struct {
	fs *flag.FlagSet

	values, file, format, column, field, generate string
	noHeader, skipEmpty, round                    bool
	size, rows, cols                              int
	seed                                          int64

	target, shift, limit  int
	text, other, key, ops string
	trace, asJSON, hasTgt bool
	example, hasShift     bool
	opts                  map[string]int
	addr                  string
	tag, kind, level      string
	count                 int // quiz questions
}

func newOptions(name string) *options {
	o := &options{fs: flag.NewFlagSet(name, flag.ContinueOnError), opts: map[string]int{}}
	fs := o.fs
	fs.SetOutput(io.Discard)
	fs.StringVar(&o.values, "values", "", "")
	fs.StringVar(&o.file, "file", "", "")
	fs.StringVar(&o.format, "format", "", "")
	fs.StringVar(&o.column, "column", "", "")
	fs.StringVar(&o.field, "field", "", "")
	fs.StringVar(&o.generate, "generate", "", "")
	fs.BoolVar(&o.noHeader, "no-header", false, "")
	fs.BoolVar(&o.skipEmpty, "skip-empty", false, "")
	fs.BoolVar(&o.round, "round", false, "")
	fs.IntVar(&o.size, "size", 20, "")
	fs.IntVar(&o.rows, "rows", 15, "")
	fs.IntVar(&o.cols, "cols", 31, "")
	fs.Int64Var(&o.seed, "seed", 1, "")
	target := func(v string) error {
		n, err := strconv.Atoi(v)
		o.target, o.hasTgt = n, true
		return err
	}
	fs.Func("target", "", target)
	fs.Func("amount", "", target)
	fs.Func("shift", "", func(v string) error {
		n, err := strconv.Atoi(v)
		o.shift, o.hasShift = n, true
		return err
	})
	fs.IntVar(&o.limit, "limit", 0, "")
	fs.StringVar(&o.text, "text", "", "")
	fs.StringVar(&o.other, "other", "", "")
	fs.StringVar(&o.other, "pattern", "", "")
	fs.StringVar(&o.key, "key", "", "")
	fs.StringVar(&o.ops, "ops", "", "")
	fs.BoolVar(&o.trace, "trace", false, "")
	fs.BoolVar(&o.asJSON, "json", false, "")
	fs.BoolVar(&o.example, "example", false, "")
	fs.StringVar(&o.addr, "addr", "127.0.0.1:8080", "")
	fs.StringVar(&o.tag, "tag", "", "")
	fs.StringVar(&o.kind, "kind", "", "")
	fs.StringVar(&o.level, "level", "", "")
	fs.IntVar(&o.count, "count", 5, "")
	fs.Func("option", "", func(v string) error {
		name, val, ok := strings.Cut(v, "=")
		n, err := strconv.Atoi(val)
		if !ok || err != nil {
			return fmt.Errorf("--option takes NAME=NUMBER, not %q", v)
		}
		o.opts[strings.TrimSpace(name)] = n
		return nil
	})
	return o
}

// parse reads flags anywhere among the arguments. A negative number is a
// value, not a flag ("run js-reduce 3 -1 4"), unless it follows a flag that
// takes a value ("--size -5").
func (o *options) parse(args []string) ([]string, error) {
	var operands []string
	for len(args) > 0 && isNegative(args[0]) {
		operands, args = append(operands, args[0]), args[1:]
	}
	for {
		if i := o.negativeOperand(args); i >= 0 {
			if err := o.fs.Parse(args[:i]); err != nil {
				return nil, usagef("%v", err)
			}
			operands = append(append(operands, o.fs.Args()...), args[i])
			args = args[i+1:]
			continue
		}
		if err := o.fs.Parse(args); err != nil {
			return nil, usagef("%v", err)
		}
		args = o.fs.Args()
		if len(args) == 0 {
			return operands, nil
		}
		if args[0] == "--" {
			return append(operands, args[1:]...), nil
		}
		operands, args = append(operands, args[0]), args[1:]
	}
}

func (o *options) dataOpts() dataset.Options { return dataset.Options{Round: o.round} }

// readFile returns the --file contents (- is standard input).
func (o *options) readFile(s Streams) ([]byte, error) {
	if o.file == "-" {
		return io.ReadAll(io.LimitReader(s.In, 64<<20))
	}
	return os.ReadFile(o.file)
}

// load returns the values the data options and operands give, or nil.
func (o *options) load(operands []string, s Streams) ([]int, error) {
	sources := 0
	for _, set := range []bool{len(operands) > 0, o.values != "", o.file != "", o.generate != ""} {
		if set {
			sources++
		}
	}
	if sources > 1 {
		return nil, usagef("give values one way: as arguments, --values, --file or --generate")
	}
	switch {
	case o.generate != "":
		return dataset.Generate(dataset.Kind(o.generate), o.size, o.seed)
	case o.values != "":
		return dataset.Parse(o.values, o.dataOpts())
	case len(operands) > 0:
		return dataset.Parse(strings.Join(operands, " "), o.dataOpts())
	case o.file != "":
		return o.loadFile(s)
	}
	return nil, nil
}

func (o *options) loadFile(s Streams) ([]int, error) {
	data, err := o.readFile(s)
	if err != nil {
		return nil, err
	}
	format := o.format
	if format == "" {
		switch strings.ToLower(filepath.Ext(o.file)) {
		case ".csv":
			format = "csv"
		case ".json":
			format = "json"
		default:
			format = "text"
		}
	}
	switch format {
	case "text":
		return dataset.Parse(string(data), o.dataOpts())
	case "json":
		return dataset.ParseJSON(data, o.field, o.dataOpts())
	case "csv":
		t, err := dataset.ReadCSV(strings.NewReader(string(data)), !o.noHeader, o.dataOpts())
		if err != nil {
			return nil, err
		}
		col := 0
		if o.column != "" {
			if n, err := strconv.Atoi(o.column); err == nil {
				col = n - 1
			} else if col = t.Find(o.column); col < 0 {
				return nil, fmt.Errorf("the CSV file has no column %q (it has %s)", o.column, strings.Join(t.Header, ", "))
			}
		} else if t.Width() > 1 {
			return nil, usagef("the CSV file has %d columns; pick one with --column", t.Width())
		}
		return t.Column(col, o.skipEmpty, o.dataOpts())
	}
	return nil, usagef("unknown format %q (text, csv or json)", format)
}

// textInput returns the text an entry reads whole (a grid, processes, a
// Banker's state): generated, from --file, --text, or the operands as lines.
func (o *options) textInput(e core.Entry, operands []string, s Streams) (string, error) {
	switch {
	case o.generate != "" && e.Input == core.GridInput:
		g, err := pathfinding.Generate(o.generate, o.rows, o.cols, o.seed)
		if err != nil {
			return "", err
		}
		return g.String(), nil
	case o.file != "":
		b, err := o.readFile(s)
		return string(b), err
	case o.text != "":
		return o.text, nil
	case len(operands) > 0:
		return strings.Join(operands, "\n"), nil
	}
	return "", usagef("give %s its input with --text, --file or as arguments (or --example)", e.ID)
}

// input builds an entry's Input from the options and operands.
func (o *options) input(e core.Entry, operands []string, s Streams) (core.Input, error) {
	if o.example && e.Example != nil {
		in := *e.Example
		in.Options = merge(in.Options, o.opts)
		return in, nil
	}
	in := core.Input{Key: o.key, Shift: 3, Other: o.other, Options: o.opts}
	if o.hasShift {
		in.Shift = o.shift
	}
	if o.hasTgt {
		in.Target = o.target
	}
	switch e.Input {
	case core.Text, core.TwoTexts, core.Pattern:
		in.Text = o.text
		if in.Text == "" && len(operands) > 0 {
			in.Text, operands = operands[0], operands[1:]
			if e.Input == core.Text {
				in.Text = strings.Join(append([]string{in.Text}, operands...), " ")
			} else if in.Other == "" && len(operands) > 0 {
				in.Other = strings.Join(operands, " ")
			}
		}
		return in, nil
	case core.GridInput, core.Processes, core.Banker, core.GraphInput, core.Points:
		text, err := o.textInput(e, operands, s)
		in.Text = text
		return in, err
	case core.Simulation:
		return in, nil
	case core.Partitions:
		in.Text = o.text
	case core.Ops:
		in.Ops = core.ParseOps(o.ops)
	default: // the rest take values, read below
	}
	values, err := o.load(operands, s)
	if err != nil {
		return in, err
	}
	in.Values = values
	if (e.Input == core.SortedInts || e.Input == core.AnyInts || e.Input == core.Amount) && !o.hasTgt {
		return in, usagef("%s needs --target (or --amount)", e.ID)
	}
	return in, nil
}

func merge(a, b map[string]int) map[string]int {
	out := map[string]int{}
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

func isNegative(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil && strings.HasPrefix(s, "-")
}

// negativeOperand returns the index of the first negative number in args that
// is a value of its own rather than the value of the flag before it, or -1.
func (o *options) negativeOperand(args []string) int {
	for i, a := range args {
		if !isNegative(a) {
			continue
		}
		if i > 0 && o.takesValue(args[i-1]) {
			continue
		}
		return i
	}
	return -1
}

// takesValue reports whether arg is a flag, written without "=", that needs a
// value after it.
func (o *options) takesValue(arg string) bool {
	name := strings.TrimLeft(arg, "-")
	if !strings.HasPrefix(arg, "-") || strings.Contains(name, "=") {
		return false
	}
	f := o.fs.Lookup(name)
	if f == nil {
		return false
	}
	b, ok := f.Value.(interface{ IsBoolFlag() bool })
	return !ok || !b.IsBoolFlag()
}
