package anagram

import (
	"testing"
)

type Case struct {
	Word     string
	Anagram  string
	Expected bool
}

// TestAnagram ...
func TestAnagram(t *testing.T) {
	cases := []Case{
		{"", "", true},
		{"aaz", "zza", false},
		{"anagram", "nagaram", true},
		{"rat", "car", false},
		{"awesome", "awesom", false},
		{"qwerty", "qeywrt", true},
		{"listen", "silent", true},
		{"india", "india", true},
		{"inlet", "silent", false},
	}

	// loop through the cases
	for _, c := range cases {
		if Anagram(c.Word, c.Anagram) != c.Expected {
			t.Errorf("Anagram(%s, %s) != %t", c.Word, c.Anagram, c.Expected)
		} else {
			t.Logf("\nWord: "+c.Word+"\nAnagram: "+c.Anagram+"\nExpected: %t\n", c.Expected)
		}
	}
}
