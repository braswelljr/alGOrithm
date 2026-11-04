package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/braswelljr/alGOrithm/internal/core"
)

func TestLearn(t *testing.T) {
	out, _, status := runCLI(t, "", "learn")
	if status != ExitOK || !strings.Contains(out, "sorting-from-zero") || !strings.Contains(out, "how-an-os-works") {
		t.Fatalf("learn: status %d\n%s", status, out)
	}
	out, _, status = runCLI(t, "", "learn", "sorting-from-zero")
	if status != ExitOK || !strings.Contains(out, "Bubble Sort") || !strings.Contains(out, "$ algorithm run bubble-sort 12 11") {
		t.Fatalf("learn sorting-from-zero: status %d\n%s", status, out)
	}
	out, _, status = runCLI(t, "", "learn", "how-an-os-works", "--json")
	var p core.PathView
	if status != ExitOK || json.Unmarshal([]byte(out), &p) != nil || len(p.Steps) == 0 || p.Steps[0].Command == "" {
		t.Fatalf("learn --json: status %d\n%s", status, out)
	}
	if _, errOut, status := runCLI(t, "", "learn", "nope"); status != ExitUsage || !strings.Contains(errOut, "no learning path") {
		t.Fatalf("learn nope: status %d, %s", status, errOut)
	}
}
