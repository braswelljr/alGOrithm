package neural

import (
	"math"
	"testing"
)

func TestXOR(t *testing.T) {
	m, err := NewMLP([]int{2, 4, 1}, 1)
	if err != nil {
		t.Fatal(err)
	}
	xs, ys := XOR()
	loss, err := m.Train(xs, ys, 0.5, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if loss[len(loss)-1] >= loss[0] || loss[len(loss)-1] > 0.02 {
		t.Errorf("loss %v -> %v", loss[0], loss[len(loss)-1])
	}
	for i, x := range xs {
		p, _ := m.Predict(x)
		if math.Abs(p[0]-ys[i][0]) > 0.3 {
			t.Errorf("%v -> %v", x, p)
		}
	}
	again, _ := NewMLP([]int{2, 4, 1}, 1)
	l2, _ := again.Train(xs, ys, 0.5, 5000)
	if l2[len(l2)-1] != loss[len(loss)-1] {
		t.Error("the same seed trained differently")
	}
}

func TestActivations(t *testing.T) {
	for _, a := range Activations {
		for _, x := range []float64{-2, -0.5, 0.5, 2} {
			h := 1e-6
			num := (a.F(x+h) - a.F(x-h)) / (2 * h)
			if math.Abs(num-a.DF(x)) > 1e-5 {
				t.Errorf("%s'(%v) = %v, numerically %v", a.Name, x, a.DF(x), num)
			}
		}
	}
}

func TestSoftmax(t *testing.T) {
	p, err := Softmax([]float64{1000, 1001, 1002})
	if err != nil {
		t.Fatal(err)
	}
	sum := p[0] + p[1] + p[2]
	if math.Abs(sum-1) > 1e-12 || !(p[2] > p[1] && p[1] > p[0]) {
		t.Errorf("%v", p)
	}
	ce, _ := CrossEntropy([]float64{0.25, 0.75}, 1)
	if math.Abs(ce+math.Log(0.75)) > 1e-12 {
		t.Errorf("ce %v", ce)
	}
	if _, err := CrossEntropy(p, 3); err == nil {
		t.Error("target out of range accepted")
	}
}
