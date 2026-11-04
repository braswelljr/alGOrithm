package numerical

import (
	"fmt"
	"math"
)

// singularTolerance is how close to zero a pivot may be before the matrix
// counts as singular.
const singularTolerance = 1e-12

// EliminationStep is one row operation of GaussianEliminationSteps: what was
// done and the augmented matrix [A | b] afterward.
type EliminationStep struct {
	Note   string
	Matrix [][]float64
}

// GaussianElimination solves A x = b for a square A: eliminate below each
// pivot to reach an upper triangular system, then substitute back from the
// last row. Partial pivoting swaps in the row with the largest pivot first,
// which keeps rounding errors from growing. a and b are not modified.
// ErrShape when the sizes do not match, ErrSingular when there is no unique
// solution.
//
// Time: O(n^3). Space: O(n^2).
func GaussianElimination(a [][]float64, b []float64) ([]float64, error) {
	return gauss(a, b, nil)
}

// GaussianEliminationSteps is GaussianElimination, also returning each row
// swap and elimination with the matrix after it.
func GaussianEliminationSteps(a [][]float64, b []float64) ([]float64, []EliminationStep, error) {
	var steps []EliminationStep
	x, err := gauss(a, b, func(note string, m [][]float64) {
		c := make([][]float64, len(m))
		for i := range m {
			c[i] = append([]float64(nil), m[i]...)
		}
		steps = append(steps, EliminationStep{note, c})
	})
	return x, steps, err
}

func gauss(a [][]float64, b []float64, step func(string, [][]float64)) ([]float64, error) {
	n := len(a)
	if n == 0 || len(b) != n {
		return nil, ErrShape
	}
	m := make([][]float64, n)
	for i, row := range a {
		if len(row) != n {
			return nil, ErrShape
		}
		m[i] = append(append(make([]float64, 0, n+1), row...), b[i])
	}
	for col := 0; col < n; col++ {
		p := col
		for r := col + 1; r < n; r++ {
			if math.Abs(m[r][col]) > math.Abs(m[p][col]) {
				p = r
			}
		}
		if math.Abs(m[p][col]) < singularTolerance {
			return nil, ErrSingular
		}
		if p != col {
			m[p], m[col] = m[col], m[p]
			if step != nil {
				step(fmt.Sprintf("swap rows %d and %d", col+1, p+1), m)
			}
		}
		for r := col + 1; r < n; r++ {
			factor := m[r][col] / m[col][col]
			if factor == 0 {
				continue
			}
			for c := col; c <= n; c++ {
				m[r][c] -= factor * m[col][c]
			}
			if step != nil {
				step(fmt.Sprintf("row %d -= %.4g * row %d", r+1, factor, col+1), m)
			}
		}
	}
	x := make([]float64, n)
	for r := n - 1; r >= 0; r-- {
		sum := m[r][n]
		for c := r + 1; c < n; c++ {
			sum -= m[r][c] * x[c]
		}
		x[r] = sum / m[r][r]
	}
	return x, nil
}
