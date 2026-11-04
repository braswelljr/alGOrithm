// Package neural builds tiny neural networks from scratch, to show how they
// learn: the activation functions and their derivatives, softmax with
// cross-entropy, and a multilayer perceptron trained by backpropagation that
// learns XOR, which no single straight line (and so no perceptron) can.
//
// Weights start from a seed, so training is repeatable. Nothing here changes
// its input.
package neural

import (
	"errors"
	"fmt"
	"math"
)

// ErrInput is returned for shapes that do not match or settings out of range.
var ErrInput = errors.New("neural: invalid input")

// Activation is a nonlinearity and its derivative, which backpropagation
// needs. Without a nonlinearity, any stack of layers is just one linear map.
type Activation struct {
	Name string
	F    func(float64) float64
	// DF is the derivative, written in terms of the input x.
	DF func(float64) float64
}

// Sigmoid squashes any number into (0, 1): 1 / (1 + e^-x). Its derivative
// is at most 1/4, so deep stacks of sigmoids learn slowly (vanishing
// gradients).
var Sigmoid = Activation{"sigmoid", sigmoid, func(x float64) float64 { s := sigmoid(x); return s * (1 - s) }}

// Tanh squashes into (-1, 1) and is centered on 0.
var Tanh = Activation{"tanh", math.Tanh, func(x float64) float64 { t := math.Tanh(x); return 1 - t*t }}

// ReLU is max(0, x): cheap, and its gradient does not vanish for x > 0,
// which is why most deep networks use it.
var ReLU = Activation{"relu", func(x float64) float64 { return math.Max(0, x) }, func(x float64) float64 {
	if x > 0 {
		return 1
	}
	return 0
}}

// Activations lists the activations by name.
var Activations = []Activation{Sigmoid, Tanh, ReLU}

func sigmoid(x float64) float64 { return 1 / (1 + math.Exp(-x)) }

// Softmax turns scores into probabilities that sum to 1: e^x[i] / sum e^x[j],
// computed after subtracting the largest score so no exponent overflows.
//
// Time: O(n).
// Space: O(n).
func Softmax(x []float64) ([]float64, error) {
	if len(x) == 0 {
		return nil, fmt.Errorf("%w: no scores", ErrInput)
	}
	top := x[0]
	for _, v := range x {
		top = math.Max(top, v)
	}
	out := make([]float64, len(x))
	sum := 0.0
	for i, v := range x {
		out[i] = math.Exp(v - top)
		sum += out[i]
	}
	for i := range out {
		out[i] /= sum
	}
	return out, nil
}

// CrossEntropy is -log(p[target]): the loss of predicting the probabilities p
// when the answer is target. It is 0 for certainty in the right answer and
// grows without bound for certainty in a wrong one.
//
// Time: O(1).
// Space: O(1).
func CrossEntropy(p []float64, target int) (float64, error) {
	if target < 0 || target >= len(p) {
		return 0, fmt.Errorf("%w: target %d out of %d classes", ErrInput, target, len(p))
	}
	return -math.Log(math.Max(p[target], 1e-15)), nil
}
