package logic

import "fmt"

// Gate is one of the seven basic logic gates.
type Gate string

// The basic gates. NOT takes one input; the others take two.
const (
	AND  Gate = "AND"
	OR   Gate = "OR"
	NOT  Gate = "NOT"
	NAND Gate = "NAND"
	NOR  Gate = "NOR"
	XOR  Gate = "XOR"
	XNOR Gate = "XNOR"
)

// Gates lists every basic gate in the order textbooks introduce them.
var Gates = []Gate{NOT, AND, OR, NAND, NOR, XOR, XNOR}

// Apply returns the output of g for inputs a and b (b is ignored by NOT).
//
// Time: O(1). Space: O(1).
func Apply(g Gate, a, b bool) (bool, error) {
	switch g {
	case AND:
		return a && b, nil
	case OR:
		return a || b, nil
	case NOT:
		return !a, nil
	case NAND: // NOT of AND
		and := a && b
		return !and, nil
	case NOR: // NOT of OR
		or := a || b
		return !or, nil
	case XOR:
		return a != b, nil
	case XNOR:
		return a == b, nil
	}
	return false, fmt.Errorf("%w: no gate %q", ErrInput, g)
}

// GateTable returns the truth table of every basic gate: one row per input
// pair (A, B) = 00, 01, 10, 11, one column per gate, 1 for true. NOT's column
// is NOT A.
//
// Time: O(1). Space: O(1).
func GateTable() (rows [][]bool) {
	for i := 0; i < 4; i++ {
		a, b := i&2 != 0, i&1 != 0
		row := []bool{a, b}
		for _, g := range Gates {
			v, _ := Apply(g, a, b)
			row = append(row, v)
		}
		rows = append(rows, row)
	}
	return rows
}
