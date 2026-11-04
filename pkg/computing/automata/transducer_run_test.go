package automata

import (
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// TestTransducerRun checks Run against Translate and TraceRun: one step per
// character, the last step's output is the translation, and the traced run
// is the same run.
func TestTransducerRun(t *testing.T) {
	m, err := ParseMealy(VendingMealy)
	if err != nil {
		t.Fatal(err)
	}
	run, err := m.Run("dnnnd")
	if err != nil {
		t.Fatal(err)
	}
	if len(run.Steps) < 5 {
		t.Fatalf("Run(dnnnd) took %d steps, want one a character", len(run.Steps))
	}
	want, _ := m.Translate("dnnnd")
	if got := run.Steps[len(run.Steps)-1].Output; got != want {
		t.Errorf("Run's last output %q, Translate %q", got, want)
	}
	var log trace.Log
	traced, err := m.TraceRun("dnnnd", &log)
	if err != nil || len(traced.Steps) != len(run.Steps) || log.Total() == 0 {
		t.Errorf("TraceRun gave %d steps and %d events (%v), Run %d steps", len(traced.Steps), log.Total(), err, len(run.Steps))
	}
	if _, err := m.Run("x"); err == nil {
		t.Error("a character outside the alphabet ran")
	}
}
