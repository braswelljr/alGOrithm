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
