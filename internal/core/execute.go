package core

import (
	"fmt"
	"strconv"
	"time"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Result is one run of an entry: what it returned, the events it reported, and
// how long it took. The command line prints it and the wasm build returns it
// as JSON, so both show the same thing.
type Result struct {
	ID        string         `json:"id"`
	Output    Output         `json:"output"`
	Events    []trace.Event  `json:"events,omitempty"`
	Truncated bool           `json:"truncated,omitempty"`
	Counts    map[string]int `json:"counts,omitempty"`
	// Nanos is the time an untraced run took, in nanoseconds.
	Nanos int64 `json:"nanos"`
}

// Execute runs the entry id on in. limit caps the events kept (0 means
// trace.DefaultLimit; a negative limit keeps only the counts). The run is
// timed without tracing, then repeated with tracing for the events, so the
// time is the algorithm's and not the recorder's.
func Execute(id string, in Input, limit int) (Result, error) {
	e, ok := Lookup(id)
	if !ok {
		return Result{}, fmt.Errorf("%w: no entry %q (see `algorithm list`)", ErrInput, id)
	}
	return execute(e, in, limit)
}

func execute(e Entry, in Input, limit int) (res Result, err error) {
	defer func() {
		// An entry that panics is a bug, but one bad input must not take
		// the browser tab or the command down with it. An entry may panic
		// with an inputError from inside a shared helper; that is a
		// rejected input, not a bug.
		if p := recover(); p != nil {
			if ie, ok := p.(inputError); ok {
				err = ie
				return
			}
			err = fmt.Errorf("%s failed: %v", e.ID, p)
		}
	}()
	res.ID = e.ID

	start := time.Now()
	out, err := e.Run(in, trace.Nop{})
	res.Nanos = time.Since(start).Nanoseconds()
	if err != nil {
		return res, err
	}
	exactText(&out)
	res.Output = out
	if !e.Traced && e.Input != Ops { // structures report pushes and pops as they go
		return res, nil
	}

	log := &trace.Log{Limit: limit}
	traced, err := e.Run(in, log)
	if err != nil {
		return res, err
	}
	exactText(&traced)
	res.Output = traced
	res.Events = log.Events
	res.Truncated = log.Truncated
	res.Counts = map[string]int{}
	for k := trace.Kind(0); int(k) < trace.NumKinds; k++ {
		if n := log.Count(k); n > 0 {
			res.Counts[k.String()] = n
		}
	}
	return res, nil
}

// Compare runs every entry in ids on the same input, counting events without
// keeping them, and returns one Result each, in the order given.
func Compare(ids []string, in Input) ([]Result, error) {
	out := make([]Result, 0, len(ids))
	for _, id := range ids {
		e, ok := Lookup(id)
		if !ok {
			return nil, fmt.Errorf("%w: no entry %q", ErrInput, id)
		}
		r, err := execute(e, in, -1)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

// maxExact is the largest integer JavaScript's numbers hold exactly (2^53).
const maxExact = 1 << 53

// exactText also writes a Number beyond what JavaScript holds exactly into
// Text, so the visualizer can show every digit (20!, F(90)).
func exactText(o *Output) {
	if o.Number != nil && o.Text == "" && (*o.Number > maxExact || *o.Number < -maxExact) {
		o.Text = strconv.Itoa(*o.Number)
	}
}
