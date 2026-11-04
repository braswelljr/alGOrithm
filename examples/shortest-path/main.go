// Command shortest-path finds the shortest distance from one town to every
// other on a small weighted road map with Dijkstra's algorithm.
package main

import (
	"fmt"
	"log"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/graphs"
)

func main() {
	g := graphs.New(false)
	g.AddEdge("Accra", "Kumasi", 250)
	g.AddEdge("Accra", "Ho", 160)
	g.AddEdge("Ho", "Kumasi", 300)
	g.AddEdge("Kumasi", "Tamale", 380)

	dist, _, err := graphs.Dijkstra(g, g.Node("Accra"))
	if err != nil {
		log.Fatal(err)
	}
	for i, d := range dist {
		fmt.Printf("%-7s %d km\n", g.Names[i], d)
	}
}
