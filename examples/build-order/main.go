// Command build-order works out the order to build packages that depend on
// each other with a topological sort, as make, Bazel, npm and the Go
// toolchain do; a cycle in the dependencies is reported as an error.
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/graphs"
)

func main() {
	g := graphs.New(true)
	// an edge a -> b means a must be built before b
	deps := [][2]string{
		{"trace", "sorting"}, {"trace", "graphs"}, {"sorting", "core"}, {"graphs", "core"},
		{"dataset", "core"}, {"core", "cli"}, {"core", "wasm"}, {"cli", "algorithm"},
	}
	for _, d := range deps {
		g.AddEdge(d[0], d[1], 1)
	}
	order, err := graphs.TopologicalSort(g)
	if err != nil {
		log.Fatal(err)
	}
	var names []string
	for _, n := range order {
		names = append(names, g.Names[n])
	}
	fmt.Println("build order: " + strings.Join(names, " -> "))
}
