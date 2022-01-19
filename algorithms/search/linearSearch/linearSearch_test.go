package linearSearch

import "testing"

// Search for a value in a sorted array
func TestLinearSearch(t *testing.T) {
	// Test Case 1
	arr := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	value := 6
	expected := 5
	actual := LinearSearch(arr, value)
	if actual != expected {
		t.Errorf("Expected %d, actual %d", expected, actual)
		return
	}
	t.Log("LinearSearch passed ", actual, " == ", expected)
}
