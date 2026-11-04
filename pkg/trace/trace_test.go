package trace_test

import (
	"reflect"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func TestKindNames(t *testing.T) {
	for k := trace.Kind(0); int(k) < trace.NumKinds; k++ {
		got, ok := trace.ParseKind(k.String())
		if !ok || got != k {
			t.Errorf("ParseKind(%q) = %v, %v; want %v, true", k.String(), got, ok, k)
		}
	}
	if _, ok := trace.ParseKind("nope"); ok {
		t.Error(`ParseKind("nope") succeeded`)
	}
	if got := trace.Kind(200).String(); got != "kind(200)" {
		t.Errorf("Kind(200).String() = %q", got)
	}
}

func TestLogLimitAndCounts(t *testing.T) {
	l := &trace.Log{Limit: 2}
	for i := 0; i < 5; i++ {
		l.Record(trace.Event{Kind: trace.Compare})
	}
	l.Record(trace.Event{Kind: trace.Swap, Indices: []int{0, 1}})
	if len(l.Events) != 2 || !l.Truncated {
		t.Errorf("kept %d events, truncated %v; want 2, true", len(l.Events), l.Truncated)
	}
	if l.Count(trace.Compare) != 5 || l.Count(trace.Swap) != 1 || l.Total() != 6 {
		t.Errorf("counts compare=%d swap=%d total=%d; want 5 1 6", l.Count(trace.Compare), l.Count(trace.Swap), l.Total())
	}
}

func TestLogZeroValueAndCountsOnly(t *testing.T) {
	var l trace.Log
	l.Record(trace.Event{Kind: trace.Visit})
	if len(l.Events) != 1 || l.Truncated {
		t.Errorf("zero Log kept %d events, truncated %v", len(l.Events), l.Truncated)
	}
	counts := &trace.Log{Limit: -1}
	counts.Record(trace.Event{Kind: trace.Visit})
	if len(counts.Events) != 0 || counts.Count(trace.Visit) != 1 {
		t.Errorf("Limit -1 kept %d events, counted %d", len(counts.Events), counts.Count(trace.Visit))
	}
}

func TestReplay(t *testing.T) {
	got := trace.Replay([]int{3, 1, 2}, []trace.Event{
		{Kind: trace.Compare, Indices: []int{0, 1}},
		{Kind: trace.Swap, Indices: []int{0, 1}},
		{Kind: trace.Set, Indices: []int{2}, Values: []int{9}},
	})
	if want := []int{1, 3, 9}; !reflect.DeepEqual(got, want) {
		t.Errorf("Replay = %v, want %v", got, want)
	}
}

func TestLogRecord(t *testing.T) {
	t.Run("the zero log keeps every event under the default limit", func(t *testing.T) {
		var l trace.Log
		l.Record(trace.Event{Kind: trace.Compare, Indices: []int{0, 1}})
		l.Record(trace.Event{Kind: trace.Swap, Indices: []int{0, 1}})
		l.Record(trace.Event{Kind: trace.Compare, Indices: []int{1, 2}})
		if len(l.Events) != 3 || l.Truncated || l.Count(trace.Compare) != 2 || l.Count(trace.Swap) != 1 || l.Total() != 3 {
			t.Errorf("%+v", l)
		}
		if l.Events[1].Kind != trace.Swap {
			t.Errorf("events out of order: %v", l.Events)
		}
	})
	t.Run("a limit keeps the first events and counts all", func(t *testing.T) {
		l := trace.Log{Limit: 2}
		for i := 0; i < 5; i++ {
			l.Record(trace.Event{Kind: trace.Set, Indices: []int{i}, Values: []int{i}})
		}
		if len(l.Events) != 2 || !l.Truncated || l.Count(trace.Set) != 5 || l.Events[1].Indices[0] != 1 {
			t.Errorf("%+v", l)
		}
	})
	t.Run("a negative limit keeps only the counts", func(t *testing.T) {
		l := trace.Log{Limit: -1}
		l.Record(trace.Event{Kind: trace.Push, Values: []int{7}})
		if len(l.Events) != 0 || !l.Truncated || l.Count(trace.Push) != 1 {
			t.Errorf("%+v", l)
		}
	})
	t.Run("an unknown kind is kept but not counted", func(t *testing.T) {
		var l trace.Log
		l.Record(trace.Event{Kind: trace.Kind(99)})
		if len(l.Events) != 1 || l.Total() != 0 || l.Count(trace.Kind(99)) != 0 {
			t.Errorf("%+v", l)
		}
	})
	t.Run("the default limit truncates", func(t *testing.T) {
		var l trace.Log
		for i := 0; i <= trace.DefaultLimit; i++ {
			l.Record(trace.Event{Kind: trace.Visit})
		}
		if len(l.Events) != trace.DefaultLimit || !l.Truncated || l.Count(trace.Visit) != trace.DefaultLimit+1 {
			t.Errorf("kept %d, truncated %v, counted %d", len(l.Events), l.Truncated, l.Count(trace.Visit))
		}
	})
	trace.Nop{}.Record(trace.Event{}) // must not panic
}

func TestRow(t *testing.T) {
	t.Run("a linked row", func(t *testing.T) {
		s := trace.Row([]string{"a", "b", "c"}, "next", "head", "tail")
		want := trace.Snapshot{
			Layout: "row",
			Nodes: []trace.Node{
				{ID: 0, Label: "a", Role: "head", Parent: -1},
				{ID: 1, Label: "b", Parent: -1},
				{ID: 2, Label: "c", Role: "tail", Parent: -1},
			},
			Edges: []trace.Edge{{From: 0, To: 1, Label: "next"}, {From: 1, To: 2, Label: "next"}},
		}
		if !reflect.DeepEqual(s, want) {
			t.Errorf("got %+v\nwant %+v", s, want)
		}
	})
	t.Run("no next means no edges", func(t *testing.T) {
		s := trace.Row([]int{1, 2}, "", "front", "back")
		if len(s.Edges) != 0 || s.Nodes[0].Role != "front" || s.Nodes[1].Role != "back" || s.Nodes[1].Label != "2" {
			t.Errorf("%+v", s)
		}
	})
	t.Run("one node takes the first role", func(t *testing.T) {
		if s := trace.Row([]int{9}, "next", "top", "bottom"); s.Nodes[0].Role != "top" || len(s.Edges) != 0 {
			t.Errorf("%+v", s)
		}
	})
	t.Run("one node with no first role takes the last", func(t *testing.T) {
		if s := trace.Row([]int{9}, "next", "", "tail"); s.Nodes[0].Role != "tail" {
			t.Errorf("%+v", s)
		}
	})
	t.Run("an empty row", func(t *testing.T) {
		if s := trace.Row([]int(nil), "next", "head", "tail"); s.Layout != "row" || len(s.Nodes) != 0 || len(s.Edges) != 0 {
			t.Errorf("%+v", s)
		}
	})
}
