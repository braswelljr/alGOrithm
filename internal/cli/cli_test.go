package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/braswelljr/alGOrithm/internal/core"
)

func runCLI(t *testing.T, stdin string, args ...string) (string, string, int) {
	t.Helper()
	var out, errb bytes.Buffer
	status := Main(append([]string{"algorithm"}, args...), Streams{In: strings.NewReader(stdin), Out: &out, Err: &errb})
	return out.String(), errb.String(), status
}

func TestCLI(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "scores.csv")
	if err := os.WriteFile(csvPath, []byte("name,score\nada,90\nalan,75\ngrace,82\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		stdin  string
		args   []string
		status int
		out    []string
		err    string
	}{
		{"help", "", nil, 0, []string{"Usage: algorithm", "serve", "mcp"}, ""},
		{"version", "", []string{"version"}, 0, []string{"algorithm "}, ""},
		{"list", "", []string{"list"}, 0, []string{"quick-sort", "round-robin", "dining-philosophers"}, ""},
		{"list category", "", []string{"list", "disk"}, 0, []string{"scan-disk"}, ""},
		{"list unknown", "", []string{"list", "nope"}, 2, nil, "unknown category"},
		{"info", "", []string{"info", "round-robin"}, 0, []string{"--option quantum=N", "scheduling/algorithms.go"}, ""},
		{"run values", "", []string{"run", "merge-sort", "5", "3", "9"}, 0, []string{"[3 5 9]"}, ""},
		{"run csv", "", []string{"run", "insertion-sort", "--file", csvPath, "--column", "score"}, 0, []string{"[75 82 90]"}, ""},
		{"run csv bad", "", []string{"run", "bubble-sort", "--file", csvPath, "--column", "1"}, 1, nil, `"ada" is not a number`},
		{"run stdin", "7 1 4", []string{"run", "selection-sort", "--file", "-"}, 0, []string{"[1 4 7]"}, ""},
		{"run search", "", []string{"run", "binary-search", "1", "5", "9", "--target", "5"}, 0, []string{"5 is at index 1"}, ""},
		{"run search no target", "", []string{"run", "binary-search", "1", "2"}, 2, nil, "needs --target"},
		{"run grid", "", []string{"run", "a-star", "--generate", "maze", "--rows", "9", "--cols", "15"}, 0, []string{"steps, cost"}, ""},
		{"run pattern", "", []string{"run", "knuth-morris-pratt", "abracadabra", "abra"}, 0, []string{"[abra]cad[abra]"}, ""},
		{"run table", "", []string{"run", "edit-distance", "kitten", "sitting"}, 0, []string{"result  3"}, ""},
		{"run processes", "", []string{"run", "srtf-scheduling", "--example"}, 0, []string{"average waiting"}, ""},
		{"run pages", "", []string{"run", "optimal-page-replacement", "7", "0", "1", "2", "0", "3", "--option", "frames=3"}, 0, []string{"faults"}, ""},
		{"run disk", "", []string{"run", "sstf-disk", "--example"}, 0, []string{"head movement  236"}, ""},
		{"run banker", "", []string{"run", "bankers-algorithm", "--example"}, 0, []string{"safe order: P1, P3, P4, P0, P2"}, ""},
		{"run simulation", "", []string{"run", "dining-philosophers", "--option", "ordered=1"}, 0, []string{"meals eaten"}, ""},
		{"run structure", "", []string{"run", "stack", "--ops", "push 1, push 2, pop"}, 0, []string{"popped 2, the last one in"}, ""},
		{"run lru", "", []string{"run", "lru-cache", "--ops", "put 1 1, put 2 2, put 3 3", "--option", "capacity=2"}, 0, []string{"evicted 1"}, ""},
		{"run bad option", "", []string{"run", "round-robin", "--example", "--option", "quantum"}, 2, nil, "NAME=NUMBER"},
		{"run trace", "", []string{"run", "bubble-sort", "2", "1", "--trace"}, 0, []string{"compare", "swap"}, ""},
		{"compare", "", []string{"compare", "bubble-sort", "merge-sort", "--generate", "random", "--size", "50"}, 0, []string{"50 values"}, ""},
		{"compare every sort skips the capped", "", []string{"compare", "--generate", "random", "--size", "500"}, 0, []string{"500 values", "merge-sort"}, "skipping bogo-sort: at most 8 values"},
		{"compare a named capped sort", "", []string{"compare", "bogo-sort", "--generate", "random", "--size", "50"}, 1, nil, "at most 8 values"},
		{"source", "", []string{"source", "dijkstra"}, 0, []string{"// algorithms/pathfinding/dijkstra.go", "func Dijkstra("}, ""},
		{"snippet", "", []string{"snippet", "quick-sort"}, 0, []string{"sorting.Quick(", "algorithm run quick-sort"}, ""},
		{"data", "", []string{"data", "--generate", "sorted", "--size", "3"}, 0, nil, ""},
		{"unknown command", "", []string{"fly"}, 2, nil, `unknown command "fly"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, errOut, status := runCLI(t, tc.stdin, tc.args...)
			if status != tc.status {
				t.Fatalf("status %d, want %d\nstdout: %s\nstderr: %s", status, tc.status, out, errOut)
			}
			for _, want := range tc.out {
				if !strings.Contains(out, want) {
					t.Errorf("stdout lacks %q:\n%s", want, out)
				}
			}
			if tc.err != "" && !strings.Contains(errOut, tc.err) {
				t.Errorf("stderr lacks %q:\n%s", tc.err, errOut)
			}
		})
	}
}

func TestServe(t *testing.T) {
	srv := httptest.NewServer(Handler())
	defer srv.Close()
	get := func(path string) (int, map[string]any) {
		resp, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()
		var v map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&v)
		return resp.StatusCode, v
	}
	post := func(path, body string) (int, map[string]any) {
		resp, err := http.Post(srv.URL+path, "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()
		var v map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&v)
		return resp.StatusCode, v
	}
	if code, v := get("/api/catalog"); code != 200 || len(v["entries"].([]any)) != len(core.Entries()) {
		t.Errorf("catalog: %d", code)
	}
	if code, v := post("/api/run/quick-sort?limit=lots", `{"values":[3,1,2]}`); code != http.StatusBadRequest || !strings.Contains(fmt.Sprint(v["error"]), `"lots"`) {
		t.Errorf("a limit that is not a number: %d %v", code, v)
	}
	if code, v := get("/api/entries/heap-sort"); code != 200 || !strings.Contains(v["code"].(string), "func Heap[") {
		t.Errorf("entry: %d", code)
	}
	if code, _ := get("/api/entries/nope"); code != 404 {
		t.Errorf("unknown entry: %d", code)
	}
	if code, v := post("/api/run/quick-sort", `{"values":[3,1,2]}`); code != 200 || v["output"].(map[string]any)["values"] == nil {
		t.Errorf("run: %d %v", code, v)
	}
	if code, _ := post("/api/run/quick-sort", `{"nope":1}`); code != 400 {
		t.Errorf("bad body: %d", code)
	}
	if code, _ := post("/api/compare", `{"ids":["bubble-sort","heap-sort"],"input":{"values":[3,2,1]}}`); code != 200 {
		t.Errorf("compare: %d", code)
	}
	if code, v := post("/api/snippet/binary-search", ``); code != 200 || !strings.Contains(v["go"].(string), "searching.Binary(") {
		t.Errorf("snippet: %d %v", code, v)
	}
}

func TestMCP(t *testing.T) {
	in := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"run_entry","arguments":{"id":"merge-sort","input":{"values":[3,1,2]}}}}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"read_source","arguments":{"id":"stack"}}}`,
		`{"jsonrpc":"2.0","id":5,"method":"nope"}`,
	}, "\n")
	out, _, status := runCLI(t, in, "mcp")
	if status != 0 {
		t.Fatalf("status %d", status)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 5 {
		t.Fatalf("%d responses, want 5 (the notification gets none):\n%s", len(lines), out)
	}
	for i, want := range []string{`"protocolVersion"`, `"run_entry"`, `\"id\": \"merge-sort\"`, `func NewStack`, `"code":-32601`} {
		if !strings.Contains(lines[i], strings.ReplaceAll(want, "\n", "\\n")) && !strings.Contains(lines[i], want) {
			t.Errorf("response %d lacks %s: %s", i+1, want, lines[i][:min(200, len(lines[i]))])
		}
	}
}

func TestNegativeNumbersAreValues(t *testing.T) {
	out, errOut, status := runCLI(t, "", "run", "merge-sort", "-3", "2", "-10", "5")
	if status != 0 || !strings.Contains(out, "[-10 -3 2 5]") {
		t.Errorf("status %d\n%s%s", status, out, errOut)
	}
	if _, errOut, status := runCLI(t, "", "data", "--generate", "random", "--size", "3", "--seed", "-5"); status != 0 {
		t.Errorf("a negative flag value failed: %s", errOut)
	}
}

func TestListFiltersAndTags(t *testing.T) {
	out, _, status := runCLI(t, "", "list", "--tag", "lifo")
	if status != 0 || !strings.Contains(out, "stack") || strings.Contains(out, "queue ") {
		t.Errorf("list --tag lifo:\n%s", out)
	}
	out, _, _ = runCLI(t, "", "list", "--kind", "simulation", "--level", "advanced")
	if !strings.Contains(out, "bankers-algorithm") || strings.Contains(out, "quick-sort") {
		t.Errorf("list --kind simulation --level advanced:\n%s", out)
	}
	out, _, status = runCLI(t, "", "tags")
	if status != 0 || !strings.Contains(out, "stable") || !strings.Contains(out, "MEANING") {
		t.Errorf("tags:\n%s", out)
	}
}
