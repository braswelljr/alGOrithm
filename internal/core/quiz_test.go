package core

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// TestQuizOnEveryExample asks questions about every entry's example and
// checks each one: 3 or 4 unique choices, the answer among them, and the
// answer matching what the run really did.
func TestQuizOnEveryExample(t *testing.T) {
	specific := 0
	for _, e := range Entries() {
		qs, err := Quiz(e.ID, *e.Example, 8, 1)
		if err != nil {
			t.Errorf("%s: %v", e.ID, err)
			continue
		}
		if len(qs) > 0 && qs[0].Kind != "result" {
			specific++
		}
		res, err := execute(e, *e.Example, 0)
		if err != nil {
			t.Fatalf("%s: %v", e.ID, err)
		}
		for _, q := range qs {
			checkQuestion(t, e, res, q)
		}
	}
	t.Logf("%d of %d entries get step-by-step questions", specific, len(Entries()))
	if specific < 50 {
		t.Errorf("only %d entries get step-by-step questions", specific)
	}
}

func checkQuestion(t *testing.T, e Entry, res Result, q Question) {
	t.Helper()
	if len(q.Choices) < 3 || len(q.Choices) > 4 {
		t.Errorf("%s %s: %d choices", e.ID, q.Kind, len(q.Choices))
		return
	}
	seen := map[string]bool{}
	for _, c := range q.Choices {
		if seen[c] {
			t.Errorf("%s %s: choice %q twice", e.ID, q.Kind, c)
		}
		seen[c] = true
	}
	if q.Answer < 0 || q.Answer >= len(q.Choices) || q.Prompt == "" || q.Explanation == "" {
		t.Errorf("%s %s: bad answer, prompt or explanation: %+v", e.ID, q.Kind, q)
		return
	}
	got := q.Choices[q.Answer]
	o, ev := res.Output, res.Events
	event := func() trace.Event {
		if q.Step < 0 || q.Step >= len(ev) {
			t.Fatalf("%s %s: step %d of %d events", e.ID, q.Kind, q.Step, len(ev))
		}
		return ev[q.Step]
	}
	want := ""
	switch q.Kind {
	case "compare-next", "swap-next":
		x := event()
		want = pairName(x.Indices[0], x.Indices[1])
		if (q.Kind == "swap-next") != (x.Kind == trace.Swap) {
			t.Errorf("%s: %s at a %s event", e.ID, q.Kind, x.Kind)
		}
		if !reflect.DeepEqual(q.Shown, arrayAt(o.Initial, ev, q.Step)) {
			t.Errorf("%s: shown array is not the array at step %d", e.ID, q.Step)
		}
	case "probe-next":
		want = posName(event().Indices[0])
	case "visit-next":
		x := event()
		if x.Kind != trace.Visit {
			t.Errorf("%s: visit-next at a %s event", e.ID, x.Kind)
		}
		if o.Grid != nil {
			want = fmtCell(o.Grid, x.Indices[0])
		} else {
			want = o.Graph.Nodes[x.Indices[0]]
		}
	case "runs-at":
		for _, s := range o.Timeline {
			if s.Start == q.Step && (s.State == "" || s.State == "running") {
				want = s.Lane
			}
		}
	case "head-next":
		want = trackName(o.Track.Path[q.Step+1])
	case "page-reference":
		want = got // checked through pageOutcome below
		if !strings.HasPrefix(got, "hit") && !strings.HasPrefix(got, "fault") {
			t.Errorf("%s: page answer %q", e.ID, got)
		}
	case "op-result":
		st := o.Steps[q.Step+1]
		if a, _ := opAnswer(st); a != got || !strings.Contains(st.Result, got) {
			t.Errorf("%s: op answer %q, result %q", e.ID, got, st.Result)
		}
		want = got
	case "result":
		want = got
	default:
		t.Errorf("%s: unknown kind %q", e.ID, q.Kind)
	}
	if want != got {
		t.Errorf("%s %s at %d: answer %q, the run says %q", e.ID, q.Kind, q.Step, got, want)
	}
}

func fmtCell(g *GridOutput, i int) string {
	q := quiz{o: Output{Grid: g}}
	return q.cellName(i)
}

func TestQuizIsDeterministic(t *testing.T) {
	for _, id := range []string{"bubble-sort", "binary-search", "dijkstra", "round-robin", "lru-page-replacement", "stack"} {
		e, _ := Lookup(id)
		a, err := Quiz(id, *e.Example, 5, 7)
		if err != nil {
			t.Fatal(err)
		}
		b, _ := Quiz(id, *e.Example, 5, 7)
		if !reflect.DeepEqual(a, b) {
			t.Errorf("%s: the same seed gave different questions", id)
		}
		if len(a) == 0 || a[0].Kind == "result" {
			t.Errorf("%s: no step-by-step questions: %+v", id, a)
		}
		for i := 1; i < len(a); i++ {
			if a[i].Step < a[i-1].Step {
				t.Errorf("%s: questions out of order", id)
			}
		}
	}
}

func TestQuizKinds(t *testing.T) {
	cases := map[string]string{
		"bubble-sort":          "compare-next|swap-next",
		"binary-search":        "probe-next",
		"a-star":               "visit-next",
		"graph-bfs":            "visit-next",
		"round-robin":          "runs-at",
		"sstf-disk":            "head-next",
		"lru-page-replacement": "page-reference",
		"queue":                "op-result",
	}
	for id, kinds := range cases {
		e, ok := Lookup(id)
		if !ok {
			t.Errorf("no entry %s", id)
			continue
		}
		qs, err := Quiz(id, *e.Example, 3, 1)
		if err != nil || len(qs) == 0 {
			t.Errorf("%s: %v, %d questions", id, err, len(qs))
			continue
		}
		if !strings.Contains(kinds, qs[0].Kind) {
			t.Errorf("%s: kind %s, want %s", id, qs[0].Kind, kinds)
		}
	}
}

func TestQuizRejects(t *testing.T) {
	if _, err := Quiz("nope", Input{}, 3, 1); !errors.Is(err, ErrInput) {
		t.Errorf("unknown entry: %v", err)
	}
	if _, err := Quiz("bubble-sort", Input{}, 3, 1); err == nil {
		t.Error("an empty input to a sort should fail as the run does")
	}
}

func TestPageOutcome(t *testing.T) {
	for _, c := range []struct {
		page          int
		before, after []int
		want          string
		ok            bool
	}{
		{1, []int{1, 2}, []int{1, 2}, "hit", true},
		{3, []int{1}, []int{1, 3}, "fault, into an empty frame", true},
		{3, []int{1, 2}, []int{3, 2}, "fault, evicting page 1", true},
		{3, []int{1, 2}, []int{1, 2}, "", false},
	} {
		got, _, ok := pageOutcome(c.page, c.before, c.after, 2)
		if got != c.want || ok != c.ok {
			t.Errorf("pageOutcome(%d, %v, %v) = %q, %v", c.page, c.before, c.after, got, ok)
		}
	}
}
