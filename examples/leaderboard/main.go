// Command leaderboard keeps the top 3 scores of a game in a min-heap of
// size 3: each new score is compared with the smallest kept one, so a stream
// of a million scores needs only O(log k) work per score and O(k) memory.
// Game backends and "top N" dashboards do exactly this.
package main

import (
	"fmt"
	"slices"

	"github.com/braswelljr/alGOrithm/pkg/structures/heaps"
)

type score struct {
	player string
	points int
}

func main() {
	scores := []score{
		{"ama", 420}, {"kofi", 980}, {"yaw", 310}, {"esi", 1250},
		{"kwame", 760}, {"abena", 1100}, {"kojo", 640}, {"afia", 990},
	}
	const k = 3
	top := heaps.NewPriorityHeap(func(a, b score) bool { return a.points < b.points })
	for _, s := range scores {
		top.Push(s)
		if top.Len() > k {
			top.Pop() // drop the smallest of the k+1
		}
	}
	var board []score
	for top.Len() > 0 {
		s, _ := top.Pop()
		board = append(board, s)
	}
	slices.Reverse(board)
	fmt.Println("top 3 players:")
	for i, s := range board {
		fmt.Printf("%d. %-6s %d\n", i+1, s.player, s.points)
	}
}
