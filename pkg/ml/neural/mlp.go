package neural

import (
	"fmt"
	"math"
	"math/rand"
)

// MLP is a fully connected network: sigmoid hidden layers and a sigmoid
// output. W[l][j][i] is the weight from unit i of layer l to unit j of layer
// l+1, and B[l][j] that unit's bias.
type MLP struct {
	Sizes []int
	W     [][][]float64
	B     [][]float64
}

// NewMLP makes a network with the given layer sizes (inputs first, outputs
// last), its weights drawn uniformly from [-1, 1) with seed.
//
// Time: O(weights).
// Space: O(weights).
func NewMLP(sizes []int, seed int64) (*MLP, error) {
	if len(sizes) < 2 {
		return nil, fmt.Errorf("%w: want at least an input and an output layer", ErrInput)
	}
	for _, s := range sizes {
		if s < 1 || s > 64 {
			return nil, fmt.Errorf("%w: each layer needs 1 to 64 units", ErrInput)
		}
	}
	rng := rand.New(rand.NewSource(seed))
	m := &MLP{Sizes: append([]int(nil), sizes...)}
	for l := 0; l+1 < len(sizes); l++ {
		w := make([][]float64, sizes[l+1])
		b := make([]float64, sizes[l+1])
		for j := range w {
			w[j] = make([]float64, sizes[l])
			for i := range w[j] {
				w[j][i] = rng.Float64()*2 - 1
			}
			b[j] = rng.Float64()*2 - 1
		}
		m.W, m.B = append(m.W, w), append(m.B, b)
	}
	return m, nil
}

// forward returns every layer's outputs, the input first.
func (m *MLP) forward(x []float64) [][]float64 {
	acts := [][]float64{x}
	for l := range m.W {
		out := make([]float64, len(m.W[l]))
		for j, w := range m.W[l] {
			z := m.B[l][j]
			for i, v := range acts[l] {
				z += w[i] * v
			}
			out[j] = sigmoid(z)
		}
		acts = append(acts, out)
	}
	return acts
}

// Predict runs x through the network.
func (m *MLP) Predict(x []float64) ([]float64, error) {
	if len(x) != m.Sizes[0] {
		return nil, fmt.Errorf("%w: %d inputs for a network of %d", ErrInput, len(x), m.Sizes[0])
	}
	a := m.forward(x)
	return a[len(a)-1], nil
}

// Train fits the network to the examples xs and targets ys by stochastic
// gradient descent with backpropagation (Rumelhart, Hinton and Williams,
// 1986): run an example forward, measure the squared error at the output,
// pass each unit's share of the blame back through the weights by the chain
// rule, and nudge every weight against its gradient times rate. It returns
// the mean squared error after each epoch.
//
// Time: O(epochs * n * weights).
// Space: O(units + epochs).
func (m *MLP) Train(xs, ys [][]float64, rate float64, epochs int) ([]float64, error) {
	if len(xs) == 0 || len(xs) != len(ys) {
		return nil, fmt.Errorf("%w: %d inputs and %d targets", ErrInput, len(xs), len(ys))
	}
	for i := range xs {
		if len(xs[i]) != m.Sizes[0] || len(ys[i]) != m.Sizes[len(m.Sizes)-1] {
			return nil, fmt.Errorf("%w: example %d does not match the network's shape", ErrInput, i)
		}
	}
	if rate <= 0 || epochs < 1 || epochs > 1_000_000 {
		return nil, fmt.Errorf("%w: rate must be positive and epochs from 1 to 1,000,000", ErrInput)
	}
	loss := make([]float64, 0, epochs)
	for e := 0; e < epochs; e++ {
		for k, x := range xs {
			acts := m.forward(x)
			out := acts[len(acts)-1]
			delta := make([]float64, len(out))
			for j, o := range out {
				delta[j] = (o - ys[k][j]) * o * (1 - o)
			}
			for l := len(m.W) - 1; l >= 0; l-- {
				prev := acts[l]
				var back []float64
				if l > 0 {
					back = make([]float64, len(prev))
					for i := range prev {
						s := 0.0
						for j := range delta {
							s += m.W[l][j][i] * delta[j]
						}
						back[i] = s * prev[i] * (1 - prev[i])
					}
				}
				for j := range delta {
					for i, v := range prev {
						m.W[l][j][i] -= rate * delta[j] * v
					}
					m.B[l][j] -= rate * delta[j]
				}
				delta = back
			}
		}
		sum := 0.0
		for k, x := range xs {
			a := m.forward(x)
			for j, o := range a[len(a)-1] {
				sum += (o - ys[k][j]) * (o - ys[k][j])
			}
		}
		loss = append(loss, sum/float64(len(xs)))
		if math.IsNaN(sum) {
			break
		}
	}
	return loss, nil
}

// XOR is the four examples of exclusive or: 1 when exactly one input is 1.
func XOR() (xs, ys [][]float64) {
	return [][]float64{{0, 0}, {0, 1}, {1, 0}, {1, 1}}, [][]float64{{0}, {1}, {1}, {0}}
}
