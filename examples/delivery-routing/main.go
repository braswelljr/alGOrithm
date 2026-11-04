// Command delivery-routing plans a courier's route across a city grid with
// A*, which steers the search toward the destination with a distance
// estimate. Games, robots and delivery apps route on grids this way.
package main

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/pathfinding"
)

func main() {
	// S is the depot, E the customer, # buildings, digits slow streets.
	city := `S...#.....
.##.#.###.
.#..9...#.
.#.####.#.
...#..5.#E`
	g, err := pathfinding.Parse(city)
	if err != nil {
		panic(err)
	}
	path, cost := pathfinding.AStar(g)
	fmt.Printf("route found: %d blocks, travel cost %d\n", len(path)-1, cost)
	cells := []byte(city)
	onPath := map[int]bool{}
	for _, c := range path {
		onPath[c] = true
	}
	i := 0
	for j, ch := range cells {
		if ch == '\n' {
			continue
		}
		if onPath[i] && ch != 'S' && ch != 'E' {
			cells[j] = '*'
		}
		i++
	}
	fmt.Println(string(cells))
}
