// Command trending-hashtags counts hashtag mentions in a stream with a
// Count-Min sketch, a fixed-size table that never undercounts: stream
// processors and network monitors find heavy hitters this way without a
// counter per distinct item.
package main

import (
	"fmt"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/structures/probabilistic"
)

func main() {
	posts := strings.Repeat("#golang #wasm ", 40) + strings.Repeat("#golang ", 25) +
		strings.Repeat("#rust ", 18) + "#zig #zig #elixir"
	sketch := probabilistic.NewCountMinSketch(256, 4)
	for _, tag := range strings.Fields(posts) {
		sketch.Add(tag, 1)
	}
	fmt.Printf("estimated mentions out of %d:\n", sketch.Total())
	for _, tag := range []string{"#golang", "#wasm", "#rust", "#zig", "#elixir", "#cobol"} {
		fmt.Printf("  %-8s %d\n", tag, sketch.Estimate(tag))
	}
}
