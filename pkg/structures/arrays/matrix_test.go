package arrays_test

import (
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/structures/arrays"
)

func TestMatrix(t *testing.T) {
	a := arrays.MatrixFrom([][]int{{1, 2}, {3, 4}})
	if a.Determinant() != -2 {
		t.Errorf("det = %d", a.Determinant())
	}
	if p := a.Multiply(arrays.MatrixFrom([][]int{{1, 0}, {0, 1}})); p.Determinant() != -2 {
		t.Error("identity product changed the matrix")
	}
}
