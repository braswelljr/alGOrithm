package recursion

// Hanoi returns the moves that take n disks from peg A to peg C, never putting
// a larger disk on a smaller one: move n-1 disks out of the way to B, move the
// largest to C, and move the n-1 back on top of it. Each move is "A -> C".
//
// Time: O(2^n), the 2^n - 1 moves. Space: O(n) for the recursion.
func Hanoi(n int) []string {
	var moves []string
	var move func(n int, from, to, via string)
	move = func(n int, from, to, via string) {
		if n == 0 {
			return
		}
		move(n-1, from, via, to)
		moves = append(moves, from+" -> "+to)
		move(n-1, via, to, from)
	}
	move(n, "A", "C", "B")
	return moves
}
