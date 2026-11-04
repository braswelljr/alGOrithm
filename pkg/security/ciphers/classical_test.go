package ciphers_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/security/ciphers"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

const english = `It was the best of times, it was the worst of times, it was the age of
wisdom, it was the age of foolishness, it was the epoch of belief, it was the
epoch of incredulity, it was the season of Light, it was the season of Darkness,
it was the spring of hope, it was the winter of despair, we had everything
before us, we had nothing before us, we were all going direct to Heaven, we
were all going direct the other way. In short, the period was so far like the
present period, that some of its noisiest authorities insisted on its being
received, for good or for evil, in the superlative degree of comparison only.`

func TestKnownVectors(t *testing.T) {
	must := func(s string, err error) string {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	pf, _ := ciphers.Playfair("hide the gold in the tree stump", "playfair example")
	tests := []struct{ name, got, want string }{
		{"atbash", ciphers.Atbash("Hello, World"), "Svool, Dliow"},
		{"affine", must(ciphers.Affine("AFFINE CIPHER", 5, 8)), "IHHWVC SWFRCP"},
		{"rail fence", ciphers.RailFence("WEAREDISCOVEREDFLEEATONCE", 3), "WECRLTEERDSOEEFEAOCAIVDEN"},
		{"playfair", pf, "BMODZBXDNABEKUDMUIXMMOUVIF"},
		{"autokey", must(ciphers.Autokey("attackatdawn", "QUEENLY")), "qnxepvytwtwp"},
		{"beaufort", must(ciphers.Beaufort("DEFENDTHEEASTWALLOFTHECASTLE", "FORTIFICATION")), "CKMPVCPVWPIWUJOGIUAPVWRIWUUK"},
		{"substitution", must(ciphers.Substitution("flee at once", "ZEBRASCDFGHIJKLMNOPQTUVWXY")), "siaa zq lkba"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
		}
	}
}

func TestRoundTrips(t *testing.T) {
	msg := "Meet me at the Old Bridge at nine, bring 3 maps!"
	letters := strings.ToUpper(strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
			return r
		}
		return -1
	}, msg))
	check := func(name, got, want string, err error) {
		t.Helper()
		if err != nil || got != want {
			t.Errorf("%s round trip = %q, %v; want %q", name, got, err, want)
		}
	}
	a, _ := ciphers.Affine(msg, 7, 3)
	d, err := ciphers.AffineDecrypt(a, 7, 3)
	check("affine", d, msg, err)
	check("atbash", ciphers.Atbash(ciphers.Atbash(msg)), msg, nil)
	for rails := 1; rails <= 8; rails++ {
		check("rail fence", ciphers.RailFenceDecrypt(ciphers.RailFence(msg, rails), rails), msg, nil)
	}
	c, _ := ciphers.Columnar(letters, "ZEBRAS")
	d, err = ciphers.ColumnarDecrypt(c, "ZEBRAS")
	check("columnar", strings.TrimRight(d, "X"), letters, err)
	c, _ = ciphers.Autokey(msg, "KEY")
	d, err = ciphers.AutokeyDecrypt(c, "KEY")
	check("autokey", d, msg, err)
	c, _ = ciphers.Beaufort(msg, "FORT")
	d, err = ciphers.Beaufort(c, "FORT")
	check("beaufort", d, msg, err)
	key := "QWERTYUIOPASDFGHJKLZXCVBNM"
	c, _ = ciphers.Substitution(msg, key)
	d, err = ciphers.SubstitutionDecrypt(c, key)
	check("substitution", d, msg, err)
	pf, _ := ciphers.Playfair("balloon", "monarchy")
	check("playfair", ciphers.PlayfairDecrypt(pf, "monarchy"), "BALXLOON", nil)

	pad := bytes.Repeat([]byte{0x5a, 0x13, 0xc4}, 30)
	h, err := ciphers.OneTimePad([]byte(msg), pad)
	if err != nil {
		t.Fatal(err)
	}
	p, err := ciphers.OneTimePadDecrypt(h, pad)
	check("one-time pad", string(p), msg, err)
	if _, err := ciphers.OneTimePad([]byte(msg), pad[:3]); !errors.Is(err, ciphers.ErrPadTooShort) {
		t.Errorf("short pad err = %v", err)
	}
}

func TestBadKeys(t *testing.T) {
	_, e1 := ciphers.Affine("x", 13, 1)
	_, e2 := ciphers.Substitution("x", "AABCDEFGHIJKLMNOPQRSTUVWXY")
	_, e3 := ciphers.Autokey("x", "123")
	_, e4 := ciphers.Enigma("x", ciphers.EnigmaConfig{Rotors: [3]string{"I", "II", "IX"}})
	_, e5 := ciphers.Enigma("x", ciphers.EnigmaConfig{Rotors: [3]string{"I", "II", "III"}, Plugboard: "AB AC"})
	_, e6 := ciphers.Enigma("x", ciphers.EnigmaConfig{Rotors: [3]string{"I", "II", "III"}, Plugboard: "A"})
	for i, err := range []error{e1, e2, e3, e4, e5, e6} {
		if !errors.Is(err, ciphers.ErrBadKey) {
			t.Errorf("case %d: err = %v, want ErrBadKey", i+1, err)
		}
	}
}

func TestTracesRecordEachLetter(t *testing.T) {
	var log trace.Log
	ciphers.TraceAtbash("ab c", &log)
	if log.Count(trace.Set) != 3 {
		t.Errorf("atbash sets = %d, want 3", log.Count(trace.Set))
	}
}

func TestEnigma(t *testing.T) {
	cfg := ciphers.EnigmaConfig{Rotors: [3]string{"I", "II", "III"}}
	if got, _ := ciphers.Enigma("AAAAA", cfg); got != "BDZGO" {
		t.Errorf("Enigma(AAAAA) = %q, want BDZGO", got)
	}
	cfg = ciphers.EnigmaConfig{Rotors: [3]string{"II", "IV", "V"}, Positions: "BLA", Rings: "BUL", Plugboard: "AV BS CG DL FU HZ IN KM OW RX"}
	msg := strings.Repeat("THEQUICKBROWNFOXJUMPSOVERTHELAZYDOG", 20) // long enough to double step
	c, err := ciphers.Enigma(msg, cfg)
	if err != nil {
		t.Fatal(err)
	}
	for i := range c {
		if c[i] == msg[i] {
			t.Fatalf("letter %d enciphered to itself", i)
		}
	}
	if back, _ := ciphers.Enigma(c, cfg); back != msg {
		t.Error("Enigma is not its own inverse")
	}
}

func TestAttacks(t *testing.T) {
	c := ciphers.Caesar(english, 11)
	if g := ciphers.CaesarBruteForce(c); g[0].Shift != 11 || g[0].Text != english || len(g) != 26 {
		t.Errorf("brute force best = %d", g[0].Shift)
	}
	if counts, shift := ciphers.FrequencyAnalysis(c); shift != 11 || counts['P'-'A'] == 0 {
		t.Errorf("frequency analysis shift = %d", shift)
	}
	if ic := ciphers.IndexOfCoincidence(english); ic < 0.06 || ic > 0.075 {
		t.Errorf("English IC = %.3f", ic)
	}

	v := ciphers.Vigenere(english, "DICKENS")
	if l := ciphers.VigenereKeyLength(v, 12)[0].Length; l != 7 {
		t.Errorf("key length = %d, want 7", l)
	}
	key, plain := ciphers.VigenereCrack(v, 12)
	if key != "DICKENS" || plain != english {
		t.Errorf("crack key = %q", key)
	}
	if key, plain := ciphers.VigenereCrack("123", 5); key != "" || plain != "123" {
		t.Error("letterless crack")
	}
}
