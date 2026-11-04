package core

import (
	"errors"
	"path"
	"strings"
	"testing"

	algorithm "github.com/braswelljr/alGOrithm"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func TestRegistryIsConsistent(t *testing.T) {
	cats := map[string]bool{}
	for _, c := range Categories {
		cats[c.ID] = true
		if algorithm.PackageDoc(c.ID) == "" {
			t.Errorf("category %s: package %s has no package doc", c.ID, c.ID)
		}
	}
	seen := map[string]bool{}
	for _, e := range Entries() {
		if e.ID == "" || seen[e.ID] || strings.ToLower(e.ID) != e.ID {
			t.Errorf("entry %q: ID empty, repeated or not lower case", e.ID)
		}
		seen[e.ID] = true
		if !cats[e.Category] {
			t.Errorf("%s: unknown category %q", e.ID, e.Category)
		}
		if e.Name == "" || e.Summary == "" || e.Time == "" || e.Space == "" || e.Run == nil || e.Input == "" || e.Example == nil {
			t.Errorf("%s: missing name, summary, cost, input, Run or Example", e.ID)
		}
		if _, ok := algorithm.File(e.File); !ok {
			t.Errorf("%s: source file %s is not embedded", e.ID, e.File)
		}
		if algorithm.Doc(e.File, e.Func) == "" {
			t.Errorf("%s: %s in %s has no doc comment", e.ID, e.Func, e.File)
		}
		if e.Import != Module+"/pkg/"+path.Dir(e.File) || path.Base(path.Dir(e.File)) != e.Package {
			t.Errorf("%s: import %s", e.ID, e.Import)
		}
		if (e.Input == Ops) != (len(e.Ops) > 0) {
			t.Errorf("%s: only structures list operations", e.ID)
		}
		if got, ok := Lookup(" " + strings.ToUpper(e.ID) + " "); !ok || got.ID != e.ID {
			t.Errorf("Lookup(%q) failed", e.ID)
		}
	}
	if len(seen) < 90 {
		t.Errorf("only %d entries", len(seen))
	}
}

func TestEveryEntryRunsItsExample(t *testing.T) {
	for _, e := range Entries() {
		t.Run(e.ID, func(t *testing.T) {
			res, err := Execute(e.ID, *e.Example, 0)
			if err != nil {
				t.Fatalf("Execute: %v", err)
			}
			if e.Traced && len(res.Events) == 0 {
				t.Error("a traced entry reported no events")
			}
			if e.Category == "sorting" {
				got := trace.Replay(append([]int(nil), res.Output.Initial...), res.Events)
				for i := range got {
					if got[i] != res.Output.Values[i] {
						t.Fatalf("replaying the events gives %v, want %v", got, res.Output.Values)
					}
				}
			}
			for _, st := range res.Output.Steps {
				if st.Error != "" {
					t.Errorf("example step %v failed: %s", st.Op, st.Error)
				}
			}
		})
	}
}

func TestInputErrors(t *testing.T) {
	cases := map[string]Input{
		"bubble-sort": {}, "binary-search": {}, "caesar-cipher": {}, "vigenere-cipher": {Text: "x"},
		"factorial": {Values: []int{21}}, "ackermann": {Values: []int{4, 1}}, "tower-of-hanoi": {Values: []int{13}},
		"a-star": {Text: "##\n##"}, "edit-distance": {}, "knuth-morris-pratt": {Text: "abc"}, "coin-change": {Values: []int{0}, Target: 3},
		"round-robin": {Text: "P1 0 0"}, "lru-page-replacement": {Values: []int{1}, Options: map[string]int{"frames": 0}},
		"scan-disk": {Values: []int{500}}, "bankers-algorithm": {Text: "nonsense"}, "race-condition": {Options: map[string]int{"threads": 99}},
		"stack": {Ops: make([]Op, maxOps+1)}, "sliding-window-max": {Values: []int{1}, Options: map[string]int{"window": 5}},
	}
	for id, in := range cases {
		if _, err := Execute(id, in, 0); !errors.Is(err, ErrInput) {
			t.Errorf("Execute(%s) error = %v, want an ErrInput", id, err)
		}
	}
	if _, err := Execute("nope", Input{}, 0); !errors.Is(err, ErrInput) {
		t.Errorf("unknown entry: %v", err)
	}
}

func TestUnknownOperationIsAStepError(t *testing.T) {
	res, err := Execute("stack", Input{Ops: ParseOps("fly")}, 0)
	if err != nil || !strings.Contains(res.Output.Steps[1].Error, `unknown operation "fly"`) {
		t.Errorf("result %+v, %v", res.Output.Steps, err)
	}
}

func TestCompareCountsWithoutKeeping(t *testing.T) {
	rs, err := Compare([]string{"bubble-sort", "merge-sort"}, Input{Values: []int{3, 2, 1}})
	if err != nil || len(rs) != 2 || len(rs[0].Events) != 0 || rs[0].Counts["compare"] != 3 {
		t.Fatalf("Compare = %+v, %v", rs, err)
	}
}

func TestStructuresReportTheirEvents(t *testing.T) {
	res, err := Execute("stack", Input{Ops: ParseOps("push 1, push 2, pop")}, 0)
	if err != nil || res.Counts["push"] != 2 || res.Counts["pop"] != 1 {
		t.Errorf("stack counts %v, %v", res.Counts, err)
	}
}
