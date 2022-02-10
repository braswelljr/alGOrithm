package ternarySearch

import "testing"

func TestTernarySearch(t *testing.T) {
	// Test Case 1
	arr := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	value := 6
	expected := 5
	actual := TernarySearch(arr, value)
	if actual != expected {
		t.Errorf("Expected %d, actual %d", expected, actual)
		return
	}
	t.Log("TernarySearch passed ", actual, " == ", expected)
}
