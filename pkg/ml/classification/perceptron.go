package classification

import "fmt"

// Perceptron is a linear classifier: the sign of w[0] + w[1]*x1 + ....
type Perceptron struct {
	Weights  []float64 // bias first
	Mistakes []int     // how many examples were misclassified in each epoch
}

// FitPerceptron learns a Perceptron (Frank Rosenblatt, 1958) from rows and
// labels of +1 or -1: go through the examples, and whenever one is on the
// wrong side, add it (times its label) to the weights. If a straight line
// separates the two classes it stops with no mistakes; if not (XOR, say) it
// never does, which Mistakes shows.
//
// Time: O(n * d * epochs).
// Space: O(d + epochs).
func FitPerceptron(xs [][]float64, ys []int, epochs int) (Perceptron, error) {
	if len(xs) == 0 || len(xs) != len(ys) {
		return Perceptron{}, fmt.Errorf("%w: %d rows and %d labels", ErrInput, len(xs), len(ys))
	}
	for i, y := range ys {
		if y != 1 && y != -1 {
			return Perceptron{}, fmt.Errorf("%w: label %d is %d, want 1 or -1", ErrInput, i, y)
		}
	}
	if epochs < 1 || epochs > 100_000 {
		return Perceptron{}, fmt.Errorf("%w: epochs must be from 1 to 100,000", ErrInput)
	}
	p := Perceptron{Weights: make([]float64, len(xs[0])+1)}
	for e := 0; e < epochs; e++ {
		wrong := 0
		for i, row := range xs {
			if p.Predict(row) != ys[i] {
				wrong++
				p.Weights[0] += float64(ys[i])
				for j, v := range row {
					p.Weights[j+1] += float64(ys[i]) * v
				}
			}
		}
		p.Mistakes = append(p.Mistakes, wrong)
		if wrong == 0 {
			break
		}
	}
	return p, nil
}

// Predict is 1 when the weighted sum is positive, else -1.
func (p Perceptron) Predict(row []float64) int {
	z := p.Weights[0]
	for j, v := range row {
		z += p.Weights[j+1] * v
	}
	if z > 0 {
		return 1
	}
	return -1
}

// Signs turns labels into the +1 and -1 a Perceptron learns: +1 for
// positive, -1 for anything else.
func Signs(labels []string, positive string) []int {
	out := make([]int, len(labels))
	for i, l := range labels {
		out[i] = -1
		if l == positive {
			out[i] = 1
		}
	}
	return out
}
