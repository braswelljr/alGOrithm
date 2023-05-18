package palindrome

import "testing"

// Test Palindrome
func TestPalindrome(t *testing.T) {
	word := "racecar"
	if !Palindrome(word) {
		t.Errorf("%s is not a palindrome", word)
		return
	}
	t.Logf("%s is a palindrome", word)
}
