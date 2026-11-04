// Command log-search finds every occurrence of an error string in a log with
// Knuth-Morris-Pratt, which never re-reads a character, and finds several
// patterns in one pass with Aho-Corasick, as grep -F and intrusion detection
// systems (Snort) do.
package main

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/strmatch"
)

func main() {
	logText := "ok ok timeout ok refused ok timeout ok ok panic ok timeout"
	fmt.Printf("\"timeout\" at offsets %v\n", strmatch.KMP(logText, "timeout"))
	patterns := []string{"timeout", "refused", "panic"}
	fmt.Println("all patterns in one pass:")
	for _, m := range strmatch.AhoCorasick(logText, patterns) {
		fmt.Printf("  %-7s at %d\n", patterns[m.Pattern], m.Start)
	}
}
