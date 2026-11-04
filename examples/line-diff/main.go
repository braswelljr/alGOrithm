// Command line-diff compares two versions of a config file line by line
// with Myers' diff, the algorithm behind git diff and GNU diff: it finds the
// shortest edit script, so unchanged lines stay unchanged.
package main

import (
	"fmt"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/text"
)

func main() {
	before := strings.Split(`port: 8080
host: localhost
debug: true
workers: 4`, "\n")
	after := strings.Split(`port: 8080
host: 0.0.0.0
workers: 8
timeout: 30s`, "\n")
	fmt.Println("--- config.yaml (before)")
	fmt.Println("+++ config.yaml (after)")
	for _, e := range text.Diff(before, after) {
		prefix := " "
		if e.Op != '=' {
			prefix = string(e.Op)
		}
		fmt.Println(prefix + e.Line)
	}
}
