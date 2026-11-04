package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/braswelljr/alGOrithm/internal/core"
)

// TestQuizCommand answers a quiz from a script: right, then a bad line and
// wrong, then the input ends before the third question.
func TestQuizCommand(t *testing.T) {
	e, _ := core.Lookup("bubble-sort")
	qs, err := core.Quiz(e.ID, *e.Example, 3, 4)
	if err != nil || len(qs) != 3 {
		t.Fatalf("%v, %d questions", err, len(qs))
	}
	wrong := (qs[1].Answer + 1) % len(qs[1].Choices)
	script := fmt.Sprintf("%d\nnine\n%d\n", qs[0].Answer+1, wrong+1)
	out, errOut, status := runCLI(t, script, "quiz", "bubble-sort", "--count", "3", "--seed", "4")
	if status != 0 || errOut != "" {
		t.Fatalf("status %d, stderr %q", status, errOut)
	}
	for _, want := range []string{"question 1 of 3", qs[0].Prompt, "Right.", "Type a number from 1 to", "Not quite", "Score: 1 of 2, 100 points, best streak 1.", "score 100  streak 1 (x1)", "[+x.]"} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
}

func TestQuizCommandJSON(t *testing.T) {
	out, _, status := runCLI(t, "", "quiz", "binary-search", "--json", "--count", "2")
	if status != 0 {
		t.Fatalf("status %d", status)
	}
	var qs []core.Question
	if err := json.Unmarshal([]byte(out), &qs); err != nil || len(qs) != 2 || qs[0].Kind != "probe-next" {
		t.Fatalf("%v: %s", err, out)
	}
	if _, _, status := runCLI(t, "", "quiz", "nope"); status != ExitUsage {
		t.Errorf("unknown entry: status %d", status)
	}
	if out, _, status := runCLI(t, "", "quiz", "quick-sort", "4", "2", "9", "1", "--count", "1"); status != 0 || !strings.Contains(out, "Score: 0 of 0, 0 points") {
		t.Errorf("own values with no answers: %d\n%s", status, out)
	}
}

func TestStreakMultiplier(t *testing.T) {
	for streak, want := range map[int]int{0: 1, 1: 1, 3: 1, 4: 2, 7: 3, 10: 4, 50: 4} {
		if got := streakMultiplier(streak); got != want {
			t.Errorf("streakMultiplier(%d) = %d, want %d", streak, got, want)
		}
	}
}
