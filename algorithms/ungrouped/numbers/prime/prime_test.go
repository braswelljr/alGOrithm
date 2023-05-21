package prime

import "testing"

// TestPrime tests the prime function.
func TestPrime(t *testing.T) {
	tests := []struct {
		n    int
		want bool
	}{
		{1, false},
		{2, true},
		{3, true},
		{4, false},
		{5, true},
		{6, false},
		{7, true},
		{8, false},
	}

	for _, test := range tests {
		if got := Prime(test.n); got != test.want {
			t.Errorf("Prime(%d) = %t; want %t", test.n, got, test.want)
		}
	}
}
