package arrays_test

import (
	"reflect"
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

func TestMatrixArithmetic(t *testing.T) {
	a := arrays.MatrixFrom([][]int{{1, 2, 3}, {4, 5, 6}})
	b := arrays.MatrixFrom([][]int{{7, 8}, {9, 10}, {11, 12}})
	if got := a.Multiply(b); got.Rows != 2 || got.Cols != 2 || !reflect.DeepEqual(got.Data, [][]int{{58, 64}, {139, 154}}) {
		t.Errorf("2x3 times 3x2 = %v, want [[58 64] [139 154]]", got.Data)
	}
	if got := a.Transform(b); !reflect.DeepEqual(got.Data, [][]int{{58, 64}, {139, 154}}) {
		t.Errorf("Transform = %v", got.Data)
	}
	tr := a.Transpose()
	if tr.Rows != 3 || tr.Cols != 2 || !reflect.DeepEqual(tr.Data, [][]int{{1, 4}, {2, 5}, {3, 6}}) {
		t.Errorf("transpose %v", tr.Data)
	}
	if !reflect.DeepEqual(tr.Transpose().Data, a.Data) {
		t.Error("transposing twice is not the identity")
	}
	sum := a.Add(arrays.MatrixFrom([][]int{{1, 1, 1}, {2, 2, 2}}))
	if !reflect.DeepEqual(sum.Data, [][]int{{2, 3, 4}, {6, 7, 8}}) {
		t.Errorf("sum %v", sum.Data)
	}
	if diff := sum.Subtract(a); !reflect.DeepEqual(diff.Data, [][]int{{1, 1, 1}, {2, 2, 2}}) {
		t.Errorf("(a + b) - a = %v", diff.Data)
	}
	m := arrays.NewMatrix(2, 3)
	m.Set(1, 2, 9)
	if m.Data[1][2] != 9 || m.ConvertToString() != "0 0 0\n0 0 9" {
		t.Errorf("Set then String: %q", m.ConvertToString())
	}
	if a.ConvertToString() != "1 2 3\n4 5 6" {
		t.Errorf("String %q", a.ConvertToString())
	}
}

func TestMatrixDeterminant(t *testing.T) {
	cases := []struct {
		data [][]int
		want int
	}{
		{[][]int{{5}}, 5},
		{[][]int{{1, 2}, {3, 4}}, -2},
		{[][]int{{2, 3, 1}, {1, 2, 1}, {1, 1, 1}}, 1},
		{[][]int{{6, 1, 1}, {4, -2, 5}, {2, 8, 7}}, -306},
		{[][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}, 0},
		{[][]int{{1, 0, 2, -1}, {3, 0, 0, 5}, {2, 1, 4, -3}, {1, 0, 5, 0}}, 30},
	}
	for _, c := range cases {
		m := arrays.MatrixFrom(c.data)
		if got := m.Determinant(); got != c.want {
			t.Errorf("det %v = %d, want %d", c.data, got, c.want)
		}
		if got := m.Transpose().Determinant(); got != c.want {
			t.Errorf("det of the transpose of %v = %d, want %d", c.data, got, c.want)
		}
	}
	// det(AB) = det(A) det(B)
	a := arrays.MatrixFrom([][]int{{2, 3, 1}, {1, 2, 1}, {1, 1, 1}})
	b := arrays.MatrixFrom([][]int{{6, 1, 1}, {4, -2, 5}, {2, 8, 7}})
	if got := a.Multiply(b).Determinant(); got != 1*-306 {
		t.Errorf("det(AB) = %d, want det(A) det(B) = -306", got)
	}
	if !reflect.DeepEqual(a.Data, [][]int{{2, 3, 1}, {1, 2, 1}, {1, 1, 1}}) {
		t.Error("Determinant changed its matrix")
	}
}

func TestMatrixInverse(t *testing.T) {
	identity := func(n int) [][]int {
		m := arrays.NewMatrix(n, n)
		for i := 0; i < n; i++ {
			m.Data[i][i] = 1
		}
		return m.Data
	}
	for _, data := range [][][]int{
		{{2, 1}, {1, 1}},
		{{2, 3, 1}, {1, 2, 1}, {1, 1, 1}},
		{{1, 2}, {3, 5}}, // determinant -1
		{{-1}},
	} {
		a := arrays.MatrixFrom(data)
		inv := a.Inverse()
		if got := a.Multiply(inv); !reflect.DeepEqual(got.Data, identity(len(data))) {
			t.Errorf("%v times its inverse %v = %v, want the identity", data, inv.Data, got.Data)
		}
		if got := inv.Multiply(a); !reflect.DeepEqual(got.Data, identity(len(data))) {
			t.Errorf("the inverse %v times %v = %v", inv.Data, data, got.Data)
		}
	}
	if got := arrays.MatrixFrom([][]int{{2, 1}, {1, 1}}).Inverse(); !reflect.DeepEqual(got.Data, [][]int{{1, -1}, {-1, 2}}) {
		t.Errorf("inverse %v, want [[1 -1] [-1 2]]", got.Data)
	}
	singular := arrays.MatrixFrom([][]int{{1, 2}, {2, 4}}).Inverse()
	if !reflect.DeepEqual(singular.Data, [][]int{{0, 0}, {0, 0}}) {
		t.Errorf("a singular matrix's inverse %v, want the zero matrix", singular.Data)
	}
}
