//go:build js && wasm

// Command wasm is the WebAssembly build of alGOrithm behind the visualizer.
// It is not a command-line program: once started it installs one object on
// the JavaScript global object and waits to be called. Every call takes and
// returns JSON strings, which the page treats as unknown until it has checked
// them (docs/rules/data-from-outside-is-unknown-until-checked.md):
//
//	alGOrithm.version                    string
//	alGOrithm.catalog()                  {categories, entries, datasets}
//	alGOrithm.run(id, input, limit)      {ok, result} | {ok: false, error}
//	alGOrithm.compare(ids, input)        {ok, results} | {ok: false, error}
//	alGOrithm.parse(format, text, opts)  {ok, values} | {ok: false, error, problems}
//	alGOrithm.preview(text, header)      {ok, header, rows, width} for a CSV file
//	alGOrithm.generate(kind, size, seed) {ok, values} | {ok: false, error}
//	alGOrithm.generateGrid(kind, rows, cols, seed) {ok, text}
//	alGOrithm.source(id)                 {ok, file, code, doc, packageDoc}
//	alGOrithm.snippet(id, input)         {ok, go, cli}
//	alGOrithm.quiz(id, input, count, seed) {ok, questions}
//
// input is a core.Input as JSON; ids a JSON array of entry IDs. format is
// "text", "csv" or "json"; opts is {column, header, skipEmpty, round, field}.
// Nothing here reaches the network or the visitor's files: the page reads a
// file or fetches a URL itself and passes the text in
// (docs/rules/the-users-data-never-leaves-the-browser.md).
//
// When the API is ready, globalThis.alGOrithmReady is called, if set. Build
// with `make web/wasm`.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"syscall/js"

	algorithm "github.com/braswelljr/alGOrithm"
	"github.com/braswelljr/alGOrithm/internal/core"
	"github.com/braswelljr/alGOrithm/internal/version"
	"github.com/braswelljr/alGOrithm/pkg/algorithms/pathfinding"
	"github.com/braswelljr/alGOrithm/pkg/dataset"
)

func main() {
	api := js.Global().Get("Object").New()
	api.Set("version", version.Get())
	api.Set("catalog", fn(func(args []js.Value) any { return catalog() }))
	api.Set("run", fn(runEntry))
	api.Set("compare", fn(compare))
	api.Set("parse", fn(parse))
	api.Set("preview", fn(preview))
	api.Set("generate", fn(generate))
	api.Set("generateGrid", fn(generateGrid))
	api.Set("source", fn(source))
	api.Set("snippet", fn(snippet))
	api.Set("example", fn(example))
	api.Set("measure", fn(measure))
	api.Set("quiz", fn(quiz))
	api.Set("paths", fn(func(args []js.Value) any { return paths() }))
	api.Set("designs", fn(func(args []js.Value) any { return designs() }))
	js.Global().Set("alGOrithm", api)
	if ready := js.Global().Get("alGOrithmReady"); ready.Type() == js.TypeFunction {
		ready.Invoke()
	}
	select {} // serve calls until the page goes away
}

// fn wraps a handler so it always returns a JSON string and never lets a Go
// panic end the wasm instance.
func fn(h func(args []js.Value) any) js.Func {
	return js.FuncOf(func(_ js.Value, args []js.Value) (out any) {
		defer func() {
			if p := recover(); p != nil {
				out = encode(failure(fmt.Errorf("internal error: %v", p)))
			}
		}()
		return encode(h(args))
	})
}

func encode(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		b, _ = json.Marshal(failure(err))
	}
	return string(b)
}

type fail struct {
	OK       bool                 `json:"ok"`
	Error    string               `json:"error"`
	Problems []dataset.ValueError `json:"problems,omitempty"`
}

func failure(err error) fail {
	f := fail{Error: err.Error()}
	var es dataset.Errors
	if errors.As(err, &es) {
		f.Problems = es
	}
	return f
}

func arg(args []js.Value, i int) string {
	if i < len(args) && args[i].Type() == js.TypeString {
		return args[i].String()
	}
	return ""
}

func argInt(args []js.Value, i int) int {
	if i < len(args) && args[i].Type() == js.TypeNumber {
		return args[i].Int()
	}
	return 0
}

func catalog() any {
	kinds := make([]string, len(dataset.Kinds))
	for i, k := range dataset.Kinds {
		kinds[i] = string(k)
	}
	return map[string]any{
		"ok":         true,
		"version":    version.Get(),
		"categories": categories(),
		"entries":    core.Entries(),
		"gridKinds":  pathfinding.Kinds,
		"tags":       tagList(),
		"datasets":   kinds,
		"maxValues":  dataset.MaxValues,
	}
}

func decodeInput(s string) (core.Input, error) {
	var in core.Input
	if s == "" {
		return in, nil
	}
	dec := json.NewDecoder(strings.NewReader(s))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		return in, fmt.Errorf("the input is not valid: %w", err)
	}
	if len(in.Values) > dataset.MaxValues {
		return in, fmt.Errorf("at most %d values", dataset.MaxValues)
	}
	return in, nil
}

func runEntry(args []js.Value) any {
	in, err := decodeInput(arg(args, 1))
	if err != nil {
		return failure(err)
	}
	res, err := core.Execute(arg(args, 0), in, argInt(args, 2))
	if err != nil {
		return failure(err)
	}
	return map[string]any{"ok": true, "result": res}
}

func compare(args []js.Value) any {
	var ids []string
	if err := json.Unmarshal([]byte(arg(args, 0)), &ids); err != nil {
		return failure(fmt.Errorf("ids must be a JSON array of strings: %w", err))
	}
	in, err := decodeInput(arg(args, 1))
	if err != nil {
		return failure(err)
	}
	results, err := core.Compare(ids, in)
	if err != nil {
		return failure(err)
	}
	return map[string]any{"ok": true, "results": results}
}

type parseOptions struct {
	Column    string `json:"column"` // 1-based number or header name; "" for the only column
	Header    *bool  `json:"header"` // default true
	SkipEmpty bool   `json:"skipEmpty"`
	Round     bool   `json:"round"`
	Field     string `json:"field"`
}

func parse(args []js.Value) any {
	format, text := arg(args, 0), arg(args, 1)
	var o parseOptions
	if s := arg(args, 2); s != "" {
		if err := json.Unmarshal([]byte(s), &o); err != nil {
			return failure(fmt.Errorf("the options are not valid: %w", err))
		}
	}
	opts := dataset.Options{Round: o.Round}
	var (
		values []int
		err    error
	)
	switch format {
	case "text":
		values, err = dataset.Parse(text, opts)
	case "json":
		values, err = dataset.ParseJSON([]byte(text), o.Field, opts)
	case "csv":
		header := o.Header == nil || *o.Header
		var t *dataset.Table
		if t, err = dataset.ReadCSV(strings.NewReader(text), header, opts); err == nil {
			col := 0
			if o.Column != "" {
				if n, convErr := strconv.Atoi(o.Column); convErr == nil {
					col = n - 1
				} else if col = t.Find(o.Column); col < 0 {
					return failure(fmt.Errorf("no column %q", o.Column))
				}
			}
			values, err = t.Column(col, o.SkipEmpty, opts)
		}
	default:
		err = fmt.Errorf("unknown format %q (text, csv or json)", format)
	}
	if err != nil {
		return failure(err)
	}
	return map[string]any{"ok": true, "values": values}
}

// preview reads a CSV file's header and first rows, so the page can offer its
// columns before reading one.
func preview(args []js.Value) any {
	header := len(args) < 2 || args[1].Type() != js.TypeBoolean || args[1].Bool()
	t, err := dataset.ReadCSV(strings.NewReader(arg(args, 0)), header, dataset.Options{})
	if err != nil {
		return failure(err)
	}
	rows := t.Rows
	if len(rows) > 8 {
		rows = rows[:8]
	}
	return map[string]any{"ok": true, "header": t.Header, "rows": rows, "width": t.Width(), "count": len(t.Rows)}
}

func generate(args []js.Value) any {
	values, err := dataset.Generate(dataset.Kind(arg(args, 0)), argInt(args, 1), int64(argInt(args, 2)))
	if err != nil {
		return failure(err)
	}
	return map[string]any{"ok": true, "values": values}
}

// categories is core.Categories with each package's doc comment.
func categories() []map[string]any {
	out := make([]map[string]any, len(core.Categories))
	for i, c := range core.Categories {
		out[i] = map[string]any{"id": c.ID, "name": c.Name, "summary": c.Summary, "import": c.Import, "topic": c.Topic, "doc": algorithm.PackageDoc(c.ID)}
	}
	return out
}

func generateGrid(args []js.Value) any {
	g, err := pathfinding.Generate(arg(args, 0), argInt(args, 1), argInt(args, 2), int64(argInt(args, 3)))
	if err != nil {
		return failure(err)
	}
	return map[string]any{"ok": true, "text": g.String()}
}

func source(args []js.Value) any {
	e, ok := core.Lookup(arg(args, 0))
	if !ok {
		return failure(fmt.Errorf("no entry %q", arg(args, 0)))
	}
	code, _ := algorithm.File(e.File)
	return map[string]any{"ok": true, "file": e.File, "code": code, "doc": algorithm.Doc(e.File, e.Func), "packageDoc": algorithm.PackageDoc(e.Package)}
}

func snippet(args []js.Value) any {
	e, ok := core.Lookup(arg(args, 0))
	if !ok {
		return failure(fmt.Errorf("no entry %q", arg(args, 0)))
	}
	in := *e.Example
	if s := arg(args, 1); s != "" {
		var err error
		if in, err = decodeInput(s); err != nil {
			return failure(err)
		}
	}
	sn := core.MakeSnippet(e, in)
	return map[string]any{"ok": true, "go": sn.Go, "cli": sn.CLI}
}

// tagList is the tag vocabulary with each tag's meaning and how many entries
// carry it, sorted by tag.
func tagList() []map[string]any {
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
	out := make([]map[string]any, len(tags))
	for i, t := range tags {
		out[i] = map[string]any{"tag": t, "meaning": core.Vocabulary[t], "count": count[t]}
	}
	return out
}

// paths is every learning path, each step with its entry's name and command.
func paths() any {
	ps := core.Paths()
	views := make([]core.PathView, len(ps))
	for i, p := range ps {
		views[i] = p.View()
	}
	return map[string]any{"ok": true, "paths": views}
}
