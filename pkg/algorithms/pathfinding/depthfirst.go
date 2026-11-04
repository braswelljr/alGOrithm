package pathfinding

import "github.com/braswelljr/alGOrithm/pkg/trace"

// DepthFirst finds a path from g.Start to g.End by depth-first search: follow
// one direction as far as it goes and back up only at a dead end. The path is
// usually not the shortest. It returns the path and its cost, or nil and -1.
//
// Time: O(V) for V cells.
// Space: O(V) for the stack.
func DepthFirst(g *Grid) ([]int, int) { return depthFirst(g, rec{trace.Nop{}}) }

// TraceDepthFirst is DepthFirst, reporting to r as the package describes.
func TraceDepthFirst(g *Grid, r trace.Recorder) ([]int, int) { return depthFirst(g, rec{r}) }

func depthFirst(g *Grid, x rec) ([]int, int) {
	prev := unreached(len(g.Cost))
	seen := make([]bool, len(g.Cost))
	stack := []int{g.Start}
	x.push(g.Start)
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[cur] {
			continue
		}
		seen[cur] = true
		x.visit(cur)
		if cur == g.End {
			return x.path(g, prev)
		}
		ns := g.Neighbors(cur)
		for i := len(ns) - 1; i >= 0; i-- { // reversed, so "up" is explored first
			if n := ns[i]; !seen[n] {
				prev[n] = cur
				stack = append(stack, n)
				x.push(n)
			}
		}
	}
	return nil, -1
}
