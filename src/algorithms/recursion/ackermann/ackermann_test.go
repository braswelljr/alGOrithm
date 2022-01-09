package ackermann

import "testing"

// Test Ackermann function
func TestAckermann(t *testing.T) {
  // Test cases
  testCases := []struct {
    m        int
    n        int
    expected int
  }{
    {0, 0, 1},
    {0, 1, 2},
    {0, 2, 3},
    {1, 0, 2},
    {1, 1, 3},
    {1, 2, 4},
    {2, 0, 3},
    {2, 1, 4},
    {2, 2, 5},
  }

  // Iterate over test cases
  for _, testCase := range testCases {
    // Check if Ackermann function returns expected value
    if actual := Ackermann(testCase.m, testCase.n); actual != testCase.expected {
      t.Errorf("Ackermann(%d, %d) = %d, expected %d", testCase.m, testCase.n, actual, testCase.expected)
    } else {
      t.Logf("Ackermann(%d, %d) = %d", testCase.m, testCase.n, actual)
    }
  }
}
