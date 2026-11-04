package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/braswelljr/alGOrithm/internal/core"
)

func TestBenchCLI(t *testing.T) {
	out, errOut, status := runCLI(t, "", "bench", "merge-sort", "--sizes", "64,128,256,512")
	if status != ExitOK {
		t.Fatalf("status %d: %s", status, errOut)
	}
	for _, want := range []string{"Merge Sort on random values", "fit      O(n log n)", "verdict  agrees", "# measured"} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout lacks %q:\n%s", want, out)
		}
	}

	out, _, status = runCLI(t, "", "bench", "bubble-sort", "--input", "sorted", "--json")
	var m core.Measurement
	if status != ExitOK || json.Unmarshal([]byte(out), &m) != nil || m.Best != "O(n)" {
		t.Errorf("bubble sort on sorted input: status %d, best %q\n%s", status, m.Best, out)
	}

	for _, args := range [][]string{{"bench"}, {"bench", "a", "b"}, {"bench", "quick-sort", "--sizes", "x"}} {
		if _, _, status := runCLI(t, "", args...); status != ExitUsage {
			t.Errorf("%v: status %d, want %d", args, status, ExitUsage)
		}
	}
	if _, errOut, status := runCLI(t, "", "bench", "stack"); status != ExitFailure || !strings.Contains(errOut, "cannot be measured") {
		t.Errorf("bench stack: status %d, %s", status, errOut)
	}

	out, _, _ = runCLI(t, "", "info", "merge-sort")
	if !strings.Contains(out, "why") || !strings.Contains(out, "algorithm bench merge-sort") {
		t.Errorf("info lacks why and measure:\n%s", out)
	}
}
