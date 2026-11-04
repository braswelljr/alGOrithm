package ciphers_test

import (
	"errors"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/security/ciphers"
)

func TestCaesarAndROT13(t *testing.T) {
	if got := ciphers.Caesar("Attack at dawn!", 3); got != "Dwwdfn dw gdzq!" {
		t.Errorf("Caesar = %q", got)
	}
	for _, k := range []int{0, 1, 13, 25, 26, -3, 100} {
		if got := ciphers.Caesar(ciphers.Caesar("Hello, World", k), -k); got != "Hello, World" {
			t.Errorf("round trip with %d = %q", k, got)
		}
	}
	if ciphers.ROT13(ciphers.ROT13("Why did the chicken")) != "Why did the chicken" || ciphers.ROT13("Hello") != "Uryyb" {
		t.Error("ROT13 wrong")
	}
}

func TestVigenere(t *testing.T) {
	if got := ciphers.Vigenere("ATTACKATDAWN", "LEMON"); got != "LXFOPVEFRNHR" {
		t.Errorf("Vigenere = %q", got)
	}
	if got := ciphers.Vigenere("Attack at dawn!", "lemon"); got != "Lxfopv ef rnhr!" {
		t.Errorf("Vigenere mixed = %q", got)
	}
	if ciphers.VigenereDecrypt("Lxfopv ef rnhr!", "LEMON") != "Attack at dawn!" || ciphers.Vigenere("abc", "123") != "abc" {
		t.Error("Vigenere decrypt or letterless key wrong")
	}
}

func TestHill(t *testing.T) {
	key := [][]int{{3, 3}, {2, 5}} // the textbook key
	got, err := ciphers.Hill("help", key)
	if err != nil || got != "HIAT" {
		t.Fatalf("Hill = %q, %v", got, err)
	}
	if back, _ := ciphers.HillDecrypt(got, key); back != "HELP" {
		t.Errorf("HillDecrypt = %q", back)
	}
	for n := 1; n <= 4; n++ {
		k := ciphers.HillKey(n, int64(n))
		enc, err := ciphers.Hill("The quick brown fox", k)
		if err != nil {
			t.Fatal(err)
		}
		dec, _ := ciphers.HillDecrypt(enc, k)
		want := "THEQUICKBROWNFOX"
		for len(want)%n != 0 {
			want += "X"
		}
		if dec != want {
			t.Errorf("n=%d: round trip %q, want %q", n, dec, want)
		}
	}
	if _, err := ciphers.Hill("x", [][]int{{2, 0}, {0, 2}}); !errors.Is(err, ciphers.ErrKey) {
		t.Error("a key with an even determinant was accepted")
	}
	if _, err := ciphers.Hill("x", [][]int{{1, 2}}); !errors.Is(err, ciphers.ErrKey) {
		t.Error("a non-square key was accepted")
	}
}
