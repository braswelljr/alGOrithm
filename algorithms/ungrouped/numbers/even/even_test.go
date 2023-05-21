package even

import "testing"

// TestEven tests the even function.
func TestEven(t *testing.T) {
	tests := []struct {
		n    int
		want bool
	}{
		{1, false},
		{2, true},
		{3, false},
		{4, true},
		{5, false},
		{6, true},
		{7, false},
		{8, true},
	}

	for _, test := range tests {
		if got := Even(test.n); got != test.want {
			t.Errorf("Even(%d) = %t; want %t", test.n, got, test.want)
		}
	}
}
