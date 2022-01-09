package tower_of_hanoi

// TowerOfHanoi returns the sequence of moves required to move n disks from the
// first peg to the last peg using the second and third pegs as temporary
// storage.
// The sequence of moves is returned as a slice of strings in the form
// "A -> B", where A and B are the names of the pegs.

func TowerOfHanoi(n int) []string {
	var moves []string
	towerOfHanoi(n, "A", "B", "C", &moves)
	return moves
}

func towerOfHanoi(n int, from string, to string, with string, moves *[]string) {
	if n == 1 {
		*moves = append(*moves, from+" -> "+to)
	} else {
		towerOfHanoi(n-1, from, with, to, moves)
		*moves = append(*moves, from+" -> "+to)
		towerOfHanoi(n-1, with, to, from, moves)
	}
}
