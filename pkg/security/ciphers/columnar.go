package ciphers

import (
	"sort"
	"strings"
)

// Columnar is the columnar transposition cipher: write the text in rows under
// the keyword's letters, then read the columns in the alphabetical order of
// those letters (ties left to right). The text is padded with X to fill the
// last row, so decryption returns it with that padding. A keyword with no
// letters returns ErrBadKey.
//
// Time: O(n log k) for a keyword of k letters. Space: O(n).
func Columnar(text, keyword string) (string, error) {
	order, err := columnOrder(keyword)
	if err != nil {
		return "", err
	}
	k := len(order)
	for len(text)%k != 0 {
		text += "X"
	}
	var b strings.Builder
	for _, col := range order {
		for i := col; i < len(text); i += k {
			b.WriteByte(text[i])
		}
	}
	return b.String(), nil
}

// ColumnarDecrypt reverses Columnar for the same keyword.
func ColumnarDecrypt(text, keyword string) (string, error) {
	order, err := columnOrder(keyword)
	if err != nil {
		return "", err
	}
	k := len(order)
	if len(text)%k != 0 {
		return "", ErrBadKey
	}
	rows := len(text) / k
	out := make([]byte, len(text))
	pos := 0
	for _, col := range order {
		for r := 0; r < rows; r++ {
			out[r*k+col] = text[pos]
			pos++
		}
	}
	return string(out), nil
}

// columnOrder returns the columns in the order they are read.
func columnOrder(keyword string) ([]int, error) {
	letters := lettersOnly(keyword)
	if letters == "" {
		return nil, ErrBadKey
	}
	order := make([]int, len(letters))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool { return letters[order[a]] < letters[order[b]] })
	return order, nil
}
