package design

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInput is wrapped by every error about the input of a model.
var ErrInput = errors.New("design: invalid input")

// check returns an input error with the message when ok is false.
func check(ok bool, format string, args ...any) error {
	if ok {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrInput, fmt.Sprintf(format, args...))
}

// Report is what every model returns: what happened in words, the numbers
// that matter, an optional table, and for the case studies the architecture
// with its requests step by step.
type Report struct {
	Lines   []string      `json:"lines,omitempty"`
	Metrics []Metric      `json:"metrics,omitempty"`
	Table   *Table        `json:"table,omitempty"`
	Arch    *Architecture `json:"architecture,omitempty"`
}

// Metric is one named number of a report, in the unit its name gives.
type Metric struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

// Table is a table of whole numbers with labeled rows and columns; a column
// name carries its unit ("latency (us)").
type Table struct {
	Columns []string `json:"columns"`
	Rows    []Row    `json:"rows"`
}

// Row is one labeled row of a Table.
type Row struct {
	Label string `json:"label"`
	Cells []int  `json:"cells"`
}

// Metric returns the value of the metric called name, and whether it is there.
func (r Report) Metric(name string) (float64, bool) {
	for _, m := range r.Metrics {
		if m.Name == name {
			return m.Value, true
		}
	}
	return 0, false
}

// String is the report as plain text: the lines, then the metrics.
func (r Report) String() string {
	var b strings.Builder
	for _, l := range r.Lines {
		b.WriteString(l)
		b.WriteByte('\n')
	}
	for _, m := range r.Metrics {
		fmt.Fprintf(&b, "%s: %g\n", m.Name, m.Value)
	}
	return b.String()
}

func (r *Report) say(format string, args ...any) {
	r.Lines = append(r.Lines, fmt.Sprintf(format, args...))
}

func (r *Report) put(name string, v float64) { r.Metrics = append(r.Metrics, Metric{name, v}) }

func (r *Report) table(cols ...string) *Table {
	r.Table = &Table{Columns: cols}
	return r.Table
}

func (t *Table) add(label string, cells ...int) { t.Rows = append(t.Rows, Row{label, cells}) }

// Kind is what a component of an architecture is.
type Kind string

// The kinds of component.
const (
	Client   Kind = "client"
	Balancer Kind = "lb"
	Service  Kind = "service"
	Cache    Kind = "cache"
	Database Kind = "db"
	Queue    Kind = "queue"
	CDN      Kind = "cdn"
	Blob     Kind = "blob"
	Search   Kind = "search"
	Worker   Kind = "worker"
)

// Status is what happened to a request on a link.
type Status string

// The statuses of a request.
const (
	OK      Status = "ok"
	Hit     Status = "hit"
	Miss    Status = "miss"
	Failed  Status = "error"
	Queued  Status = "queued"
	Dropped Status = "dropped"
)

// Architecture is a design drawn as components and the links between them,
// with one Frame per step of the simulation: the requests on each link and
// each component's load.
type Architecture struct {
	Components []Component `json:"components"`
	Links      []Link      `json:"links"`
	Frames     []Frame     `json:"frames"`
}

// Component is one box of an architecture.
type Component struct {
	ID    string `json:"id"`
	Kind  Kind   `json:"kind"`
	Label string `json:"label"`
	Notes string `json:"notes,omitempty"`
}

// Link is a connection that requests travel along, from one component to another.
type Link struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label,omitempty"`
}

// Request is one request on a link during a frame.
type Request struct {
	Link   int    `json:"link"`
	ID     int    `json:"id"`
	Status Status `json:"status"`
}

// Load is a component's state at the end of a frame. HitRate is a percent,
// and -1 when the component is not a cache.
type Load struct {
	Component string `json:"component"`
	Load      int    `json:"load"`
	Queue     int    `json:"queue"`
	HitRate   int    `json:"hitRate"`
}

// Frame is one step of an architecture's simulation.
type Frame struct {
	Note     string    `json:"note"`
	Requests []Request `json:"requests,omitempty"`
	Loads    []Load    `json:"loads,omitempty"`
}

// arch starts an architecture from its components and links, each link
// written "from>to" or "from>to:label".
func arch(cs []Component, links ...string) *Architecture {
	a := &Architecture{Components: cs}
	for _, l := range links {
		ends, label, _ := strings.Cut(l, ":")
		from, to, _ := strings.Cut(ends, ">")
		a.Links = append(a.Links, Link{From: from, To: to, Label: label})
	}
	return a
}

// link returns the index of the link from one component to another, or -1.
func (a *Architecture) link(from, to string) int {
	for i, l := range a.Links {
		if l.From == from && l.To == to {
			return i
		}
	}
	return -1
}

// frame adds a frame and returns it.
func (a *Architecture) frame(note string) *Frame {
	a.Frames = append(a.Frames, Frame{Note: note})
	return &a.Frames[len(a.Frames)-1]
}

// send puts request id on the link from one component to another in f.
func (a *Architecture) send(f *Frame, from, to string, id int, s Status) {
	if i := a.link(from, to); i >= 0 {
		f.Requests = append(f.Requests, Request{Link: i, ID: id, Status: s})
	}
}

// load records a component's load in f.
func (f *Frame) load(component string, load, queue, hitRate int) {
	f.Loads = append(f.Loads, Load{component, load, queue, hitRate})
}

// percent is part of whole as a whole-number percent, or 0 for an empty whole.
func percent(part, whole int) int {
	if whole == 0 {
		return 0
	}
	return part * 100 / whole
}
