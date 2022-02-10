package euclidean

import "testing"

// TestEuclidean ...
func TestEuclidean(t *testing.T) {
	// Test 1
	if Euclidean(10, 5) != 5 {
		t.Error("Euclidean(10, 5) != 5")
	}

	// Test 2
	if Euclidean(10, 0) != 10 {
		t.Error("Euclidean(10, 0) != 10")
	}

	t.Log("Euclidean() passed")
}
