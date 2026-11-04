package ciphers

import (
	"fmt"
	"strings"
)

// The Enigma I rotors and reflector B, as wired in the army machine.
var (
	enigmaRotors = map[string]struct {
		wiring string
		notch  byte
	}{
		"I":   {"EKMFLGDQVZNTOWYHXUSPAIBRCJ", 'Q'},
		"II":  {"AJDKSIRUXBLHWTMCQGZNPYFVOE", 'E'},
		"III": {"BDFHJLCPRTXVZNYEIWGAKMUSQO", 'V'},
		"IV":  {"ESOVPZJAYQUIRHXLNFTGKDCMWB", 'J'},
		"V":   {"VZBRGITYUPSDNHLXAWMJQOFECK", 'Z'},
	}
	enigmaReflectorB = "YRUHQSLDPXNGOKMIEBFZCWVJAT"
)

// EnigmaConfig sets up a three-rotor Enigma I: rotors left to right (of I to
// V), their starting positions and ring settings as three letters each, and
// plugboard pairs like "AB CD". Empty positions and rings mean "AAA".
type EnigmaConfig struct {
	Rotors    [3]string
	Positions string
	Rings     string
	Plugboard string
}

type rotor struct {
	fwd, bwd  [26]int
	notch     int
	pos, ring int
}

func (r *rotor) forward(c int) int {
	shift := r.pos - r.ring
	return ((r.fwd[(c+shift+26)%26]-shift)%26 + 26) % 26
}

func (r *rotor) backward(c int) int {
	shift := r.pos - r.ring
	return ((r.bwd[(c+shift+26)%26]-shift)%26 + 26) % 26
}

// Enigma encrypts text with the German Enigma machine: each key press steps
// the rotors (the right one every time, the middle one when the right passes
// its notch, and the middle and left together when the middle is at its own
// notch: the double step), then the signal goes through the plugboard, the
// rotors right to left, the reflector, back through the rotors and the
// plugboard. The reflector makes it its own inverse: with the same settings,
// Enigma(Enigma(s)) == s, and no letter ever becomes itself, the flaw Turing's
// bombe exploited. Letters only, upper-cased; other characters are dropped.
//
// Time: O(n). Space: O(n).
func Enigma(text string, c EnigmaConfig) (string, error) {
	rs, plug, err := c.build()
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, ch := range []byte(lettersOnly(text)) {
		// Step before enciphering, double step included.
		if rs[1].pos == rs[1].notch {
			rs[0].pos = (rs[0].pos + 1) % 26
			rs[1].pos = (rs[1].pos + 1) % 26
		} else if rs[2].pos == rs[2].notch {
			rs[1].pos = (rs[1].pos + 1) % 26
		}
		rs[2].pos = (rs[2].pos + 1) % 26
		x := plug[ch-'A']
		for i := 2; i >= 0; i-- {
			x = rs[i].forward(x)
		}
		x = int(enigmaReflectorB[x] - 'A')
		for i := 0; i < 3; i++ {
			x = rs[i].backward(x)
		}
		b.WriteByte(byte('A' + plug[x]))
	}
	return b.String(), nil
}

func (c EnigmaConfig) build() ([3]*rotor, [26]int, error) {
	var rs [3]*rotor
	var plug [26]int
	for i := range plug {
		plug[i] = i
	}
	pos, ring := strings.ToUpper(c.Positions), strings.ToUpper(c.Rings)
	if pos == "" {
		pos = "AAA"
	}
	if ring == "" {
		ring = "AAA"
	}
	if len(lettersOnly(pos)) != 3 || len(lettersOnly(ring)) != 3 {
		return rs, plug, fmt.Errorf("%w: positions and rings are three letters each", ErrBadKey)
	}
	for i, name := range c.Rotors {
		spec, ok := enigmaRotors[strings.ToUpper(name)]
		if !ok {
			return rs, plug, fmt.Errorf("%w: no rotor %q (I to V)", ErrBadKey, name)
		}
		r := &rotor{notch: int(spec.notch - 'A'), pos: int(pos[i] - 'A'), ring: int(ring[i] - 'A')}
		for j := 0; j < 26; j++ {
			w := int(spec.wiring[j] - 'A')
			r.fwd[j], r.bwd[w] = w, j
		}
		rs[i] = r
	}
	for _, pair := range strings.Fields(strings.ToUpper(c.Plugboard)) {
		p := lettersOnly(pair)
		if len(p) != 2 {
			return rs, plug, fmt.Errorf("%w: plugboard pair %q", ErrBadKey, pair)
		}
		a, b := int(p[0]-'A'), int(p[1]-'A')
		if plug[a] != a || plug[b] != b || a == b {
			return rs, plug, fmt.Errorf("%w: plugboard pair %q", ErrBadKey, pair)
		}
		plug[a], plug[b] = b, a
	}
	return rs, plug, nil
}
