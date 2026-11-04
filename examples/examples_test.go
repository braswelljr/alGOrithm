package examples_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestExamplesRun builds every example once and runs it, checking it exits
// cleanly and prints its expected first line, so the examples never fall
// behind the packages they use.
func TestExamplesRun(t *testing.T) {
	if testing.Short() {
		t.Skip("builds every example")
	}
	firstLines := map[string]string{
		"api-pagination":      "GET /items?page=1",
		"api-rate-limit":      "token bucket:",
		"build-order":         "build order: ",
		"cache-sharding":      "keys moved when cache-d joined:",
		"cpu-schedule":        "[ 0- 3) P1",
		"customer-segments":   "3 segments",
		"deadlock-check":      "safe: true",
		"delivery-routing":    "route found:",
		"elevator":            "SCAN stops:",
		"file-integrity":      "published root hash:",
		"flight-routes":       "cheapest fares from ACC:",
		"fraud-rings":         "3 groups of linked accounts:",
		"house-prices":        "price = ",
		"http-cache":          "serving 7 requests",
		"inventory-index":     "indexed 8 SKUs",
		"large-file-sort":     "pass 0:",
		"leader-election":     "",
		"leaderboard":         "top 3 players:",
		"line-diff":           "--- config.yaml (before)",
		"load-balancer":       "least-connections routing:",
		"log-compression":     "original: 960 bytes",
		"log-dedup":           "unique log lines:",
		"log-search":          "\"timeout\" at offsets [6 28 51]",
		"moving-average":      "hours  avg   peak",
		"office-subnets":      "engineering  10.20.0.0/24",
		"password-hashing":    "stored hash:",
		"print-queue":         "printer timeline (pages):",
		"search-autocomplete": "suggestions for \"la\":",
		"shortest-path":       "Accra   0 km",
		"sort-and-trace":      "sorted: [1 2 3 5 7 9]",
		"spam-filter":         "\"free prize inside\"",
		"spell-checker":       "algorthm: did you mean \"algorithm\"?",
		"text-editor-buffer":  "start:",
		"timestamp-lookup":    "4 entries between",
		"trending-hashtags":   "estimated mentions",
		"unique-visitors":     "estimated unique visitors:",
	}
	dirs, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	if out, err := exec.Command("go", "build", "-o", bin+string(filepath.Separator), "./...").CombinedOutput(); err != nil {
		t.Fatalf("go build ./examples/...: %v\n%s", err, out)
	}
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		name := d.Name()
		want, ok := firstLines[name]
		if !ok {
			t.Errorf("%s has no expected first line in this test", name)
			continue
		}
		t.Run(name, func(t *testing.T) {
			exe := filepath.Join(bin, name)
			if runtime.GOOS == "windows" {
				exe += ".exe"
			}
			out, err := exec.Command(exe).CombinedOutput()
			if err != nil {
				t.Fatalf("exit: %v\n%s", err, out)
			}
			if first, _, _ := strings.Cut(string(out), "\n"); !strings.HasPrefix(first, want) {
				t.Errorf("first line %q, want prefix %q", first, want)
			}
		})
	}
}
