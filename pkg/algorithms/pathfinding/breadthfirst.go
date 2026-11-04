package pathfinding

import "github.com/braswelljr/alGOrithm/pkg/trace"

// BreadthFirst finds the path from g.Start to g.End with the fewest steps
// (ignoring cell costs) by breadth-first search: explore every cell one step
// away, then two, and so on. It returns the path and its cost, or nil and -1.
//
// Time: O(V) for V cells (each has at most four neighbors).
// Space: O(V).
func BreadthFirst(g *Grid) ([]int, int) { return breadthFirst(g, rec{trace.Nop{}}) }

// TraceBreadthFirst is BreadthFirst, reporting to r as the package describes.
func TraceBreadthFirst(g *Grid, r trace.Recorder) ([]int, int) { return breadthFirst(g, rec{r}) }

func breadthFirst(g *Grid, x rec) ([]int, int) {
	prev := unreached(len(g.Cost))
	seen := make([]bool, len(g.Cost))
	seen[g.Start] = true
	queue := []int{g.Start}
	x.push(g.Start)
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		x.visit(cur)
		if cur == g.End {
			return x.path(g, prev)
		}
		for _, n := range g.Neighbors(cur) {
			if !seen[n] {
				seen[n], prev[n] = true, cur
				queue = append(queue, n)
				x.push(n)
			}
		}
	}
	return nil, -1
}
