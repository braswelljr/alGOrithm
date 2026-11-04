// Command flight-routes finds the cheapest fare from one airport to every
// other with Dijkstra's algorithm, the core of route planners, network
// routing (OSPF) and map directions.
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/graphs"
)

func main() {
	g := graphs.New(true)
	fares := []struct {
		from, to string
		usd      int
	}{
		{"ACC", "LOS", 180}, {"ACC", "LHR", 620}, {"LOS", "LHR", 390}, {"LOS", "JFK", 700},
		{"LHR", "JFK", 310}, {"ACC", "DXB", 540}, {"DXB", "JFK", 480}, {"LHR", "DXB", 260},
	}
	for _, f := range fares {
		g.AddEdge(f.from, f.to, f.usd)
	}
	src := g.Node("ACC")
	dist, prev, err := graphs.Dijkstra(g, src)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("cheapest fares from ACC:")
	for _, to := range []string{"LOS", "LHR", "DXB", "JFK"} {
		t := g.Node(to)
		var hops []string
		for _, n := range graphs.PathTo(prev, src, t) {
			hops = append(hops, g.Names[n])
		}
		fmt.Printf("  %s $%-4d via %s\n", to, dist[t], strings.Join(hops, " -> "))
	}
}
