package regression

import (
	"fmt"
	"math"
)

// Logistic is a yes-or-no model: P(yes | x) = sigmoid(w[0] + w[1]*x1 + ...).
type Logistic struct {
	Weights []float64 // bias first
	Loss    []float64 // the mean log loss after each epoch
}

// FitLogistic learns a Logistic model by gradient descent on the log loss
// (cross-entropy), from rows xs and labels ys of 0 or 1, starting from zero
// weights. The boundary it learns between yes and no is a straight line (a
// plane in more dimensions).
//
// Time: O(n * d * epochs).
// Space: O(d + epochs).
func FitLogistic(xs [][]float64, ys []int, rate float64, epochs int) (Logistic, error) {
	if len(xs) == 0 || len(xs) != len(ys) {
		return Logistic{}, fmt.Errorf("%w: %d rows and %d labels", ErrInput, len(xs), len(ys))
	}
	for i, y := range ys {
		if y != 0 && y != 1 {
			return Logistic{}, fmt.Errorf("%w: label %d is %d, want 0 or 1", ErrInput, i, y)
		}
		if len(xs[i]) != len(xs[0]) {
			return Logistic{}, fmt.Errorf("%w: row %d has %d features, the first has %d", ErrInput, i, len(xs[i]), len(xs[0]))
		}
	}
	if rate <= 0 || epochs < 1 || epochs > 1_000_000 {
		return Logistic{}, fmt.Errorf("%w: rate must be positive and epochs from 1 to 1,000,000", ErrInput)
	}
	m := Logistic{Weights: make([]float64, len(xs[0])+1)}
	n := float64(len(xs))
	for e := 0; e < epochs; e++ {
		grad := make([]float64, len(m.Weights))
		for i, row := range xs {
			d := m.Probability(row) - float64(ys[i])
			grad[0] += d / n
			for j, v := range row {
				grad[j+1] += d * v / n
			}
		}
		for j := range m.Weights {
			m.Weights[j] -= rate * grad[j]
		}
		loss := 0.0
		for i, row := range xs {
			p := math.Min(math.Max(m.Probability(row), 1e-12), 1-1e-12)
			if ys[i] == 1 {
				loss -= math.Log(p) / n
			} else {
				loss -= math.Log(1-p) / n
			}
		}
		m.Loss = append(m.Loss, loss)
	}
	return m, nil
}

// Probability is the model's P(yes) for the features row.
func (m Logistic) Probability(row []float64) float64 {
	z := m.Weights[0]
	for j, v := range row {
		z += m.Weights[j+1] * v
	}
	return 1 / (1 + math.Exp(-z))
}

// Predict is 1 when Probability(row) is at least one half, else 0.
func (m Logistic) Predict(row []float64) int {
	if m.Probability(row) >= 0.5 {
		return 1
	}
	return 0
}

// Binary turns targets of 0 and 1 (as Parse reads them) into the labels
// FitLogistic takes, or fails if any target is something else.
func Binary(ys []float64) ([]int, error) {
	out := make([]int, len(ys))
	for i, y := range ys {
		switch y {
		case 0, 1:
			out[i] = int(y)
		default:
			return nil, fmt.Errorf("%w: target %d is %v, want 0 or 1", ErrInput, i, y)
		}
	}
	return out, nil
}
