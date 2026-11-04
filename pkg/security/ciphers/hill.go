package ciphers

import (
	"errors"
	"math/rand"
	"strings"
)

// ErrKey is returned for a Hill key that is not square or not invertible
// modulo 26 (its determinant shares a factor with 26), which could not be
// decrypted.
var ErrKey = errors.New("ciphers: the Hill key must be a square matrix invertible modulo 26")

// Hill encrypts text with the Hill cipher (Lester Hill, 1929): the letters,
// upper-cased with everything else dropped and padded with X to a multiple of
// n, are taken n at a time as a vector and multiplied by the n x n key,
// modulo 26.
//
// Time: O(len(text) * n). Space: O(len(text)).
func Hill(text string, key [][]int) (string, error) {
	if err := checkHillKey(key); err != nil {
		return "", err
	}
	return hillApply(hillLetters(text, len(key)), key), nil
}

// HillDecrypt reverses Hill: it multiplies by the inverse of the key modulo
// 26. The padding X letters stay.
func HillDecrypt(text string, key [][]int) (string, error) {
	inv, err := HillInverse(key)
	if err != nil {
		return "", err
	}
	return hillApply(hillLetters(text, len(key)), inv), nil
}

// HillKey returns a random n x n key that is invertible modulo 26, from seed.
func HillKey(n int, seed int64) [][]int {
	r := rand.New(rand.NewSource(seed))
	for {
		key := make([][]int, n)
		for i := range key {
			key[i] = make([]int, n)
			for j := range key[i] {
				key[i][j] = r.Intn(26)
			}
		}
		if checkHillKey(key) == nil {
			return key
		}
	}
}

// HillInverse returns the inverse of key modulo 26: the adjugate times the
// inverse of the determinant.
func HillInverse(key [][]int) ([][]int, error) {
	if err := checkHillKey(key); err != nil {
		return nil, err
	}
	n := len(key)
	dinv := modInverse(mod26(det(key)))
	inv := make([][]int, n)
	for i := range inv {
		inv[i] = make([]int, n)
		for j := range inv[i] {
			// adj[i][j] is the cofactor of key[j][i].
			c := det(minor(key, j, i))
			if (i+j)%2 == 1 {
				c = -c
			}
			inv[i][j] = mod26(c * dinv)
		}
	}
	return inv, nil
}

func checkHillKey(key [][]int) error {
	if len(key) == 0 {
		return ErrKey
	}
	for _, row := range key {
		if len(row) != len(key) {
			return ErrKey
		}
	}
	if modInverse(mod26(det(key))) < 0 {
		return ErrKey
	}
	return nil
}

func hillLetters(text string, n int) []int {
	var out []int
	for _, c := range strings.ToUpper(text) {
		if c >= 'A' && c <= 'Z' {
			out = append(out, int(c-'A'))
		}
	}
	for len(out)%n != 0 {
		out = append(out, 'X'-'A')
	}
	return out
}

func hillApply(letters []int, key [][]int) string {
	n := len(key)
	out := make([]byte, len(letters))
	for b := 0; b < len(letters); b += n {
		for i := 0; i < n; i++ {
			sum := 0
			for j := 0; j < n; j++ {
				sum += key[i][j] * letters[b+j]
			}
			out[b+i] = byte('A' + mod26(sum))
		}
	}
	return string(out)
}

func det(m [][]int) int {
	switch len(m) {
	case 0:
		return 1 // the empty product: what a 1x1 matrix's cofactor needs
	case 1:
		return m[0][0]
	case 2:
		return m[0][0]*m[1][1] - m[0][1]*m[1][0]
	}
	d := 0
	for j := range m {
		c := m[0][j] * det(minor(m, 0, j))
		if j%2 == 1 {
			c = -c
		}
		d += c
	}
	return d
}

func minor(m [][]int, r, c int) [][]int {
	var out [][]int
	for i := range m {
		if i == r {
			continue
		}
		var row []int
		for j := range m[i] {
			if j != c {
				row = append(row, m[i][j])
			}
		}
		out = append(out, row)
	}
	return out
}

func mod26(x int) int { return ((x % 26) + 26) % 26 }

// modInverse returns x's inverse modulo 26, or -1 when there is none.
func modInverse(x int) int {
	for i := 1; i < 26; i++ {
		if x*i%26 == 1 {
			return i
		}
	}
	return -1
}
