// Command algorithm runs the alGOrithm algorithms and data structures on your
// own data: sorts, searches, pathfinding, dynamic programming, data
// structures, CPU and disk scheduling, page replacement, concurrency
// simulations and more, with every step they take. It also serves them as a
// JSON API (algorithm serve) and as MCP tools (algorithm mcp).
//
// Install with: go install github.com/braswelljr/alGOrithm/cmd/algorithm@latest
package main

import (
	"os"

	"github.com/braswelljr/alGOrithm/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args, cli.Streams{In: os.Stdin, Out: os.Stdout, Err: os.Stderr}))
}
