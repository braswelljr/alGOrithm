package ciphers

import (
	"fmt"
	"math"
	"sort"
)

// englishFreq is the share of each letter in English text, A to Z.
var englishFreq = [26]float64{
	0.08167, 0.01492, 0.02782, 0.04253, 0.12702, 0.02228, 0.02015, 0.06094, 0.06966, 0.00153, 0.00772, 0.04025, 0.02406,
	0.06749, 0.07507, 0.01929, 0.00095, 0.05987, 0.06327, 0.09056, 0.02758, 0.00978, 0.02360, 0.00150, 0.01974, 0.00074,
}

// LetterCounts counts each letter of text, case ignored.
func LetterCounts(text string) [26]int {
	var c [26]int
	for i := 0; i < len(text); i++ {
		if idx, _, ok := letterIndex(text[i]); ok {
			c[idx]++
		}
	}
	return c
}

// ChiSquared measures how far text's letter counts are from English: lower is
// more English-like. Text without letters scores +Inf.
func ChiSquared(text string) float64 {
	c := LetterCounts(text)
	n := 0
	for _, v := range c {
		n += v
	}
	if n == 0 {
		return math.Inf(1)
	}
	score := 0.0
	for i, v := range c {
		expected := englishFreq[i] * float64(n)
		d := float64(v) - expected
		score += d * d / expected
	}
	return score
}

// Guess is one candidate decryption with its English score (chi-squared:
// lower is better).
type Guess struct {
	Shift int     `json:"shift"`
	Text  string  `json:"text"`
	Score float64 `json:"score"`
}

// CaesarBruteForce tries all 26 shifts on ciphertext and returns them best
// first by ChiSquared: with only 26 keys, Caesar falls to trying them all.
//
// Time: O(26 * n). Space: O(26 * n).
func CaesarBruteForce(ciphertext string) []Guess {
	out := make([]Guess, 26)
	for k := 0; k < 26; k++ {
		p := Caesar(ciphertext, -k)
		out[k] = Guess{Shift: k, Text: p, Score: ChiSquared(p)}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Score < out[j].Score })
	return out
}

// FrequencyAnalysis returns the letter counts of ciphertext and the Caesar
// shift whose decryption looks most like English, found by comparing letter
// frequencies (al-Kindi's method, 9th century) rather than reading every
// candidate.
//
// Time: O(26 * n). Space: O(1).
func FrequencyAnalysis(ciphertext string) (counts [26]int, shift int) {
	return LetterCounts(ciphertext), CaesarBruteForce(ciphertext)[0].Shift
}

// IndexOfCoincidence is the chance that two letters drawn from text are the
// same: about 0.066 for English, 0.038 for uniformly random letters.
func IndexOfCoincidence(text string) float64 {
	c := LetterCounts(text)
	n, sum := 0, 0
	for _, v := range c {
		n += v
		sum += v * (v - 1)
	}
	if n < 2 {
		return 0
	}
	return float64(sum) / float64(n*(n-1))
}

// KeyLength is the evidence for one Vigenère key length.
type KeyLength struct {
	Length int `json:"length"`
	// Kasiski counts the repeated trigrams whose distance the length divides.
	Kasiski int `json:"kasiski"`
	// IC is the average index of coincidence of the columns the length
	// splits the text into: near 0.066 when the length is right.
	IC float64 `json:"ic"`
}

// VigenereKeyLength ranks key lengths 1 to maxLen for a Vigenère ciphertext,
// best first: by the columns' index of coincidence (Friedman, 1920), with the
// Kasiski examination (1863) of repeated trigrams as a second signal.
//
// Time: O(n * maxLen + n^2) for the trigram distances. Space: O(n).
func VigenereKeyLength(ciphertext string, maxLen int) []KeyLength {
	s := lettersOnly(ciphertext)
	maxLen = max(1, min(maxLen, len(s)/2))
	seen := map[string]int{}
	var dists []int
	for i := 0; i+3 <= len(s); i++ {
		tri := s[i : i+3]
		if j, ok := seen[tri]; ok {
			dists = append(dists, i-j)
		}
		seen[tri] = i
	}
	out := make([]KeyLength, 0, maxLen)
	for l := 1; l <= maxLen; l++ {
		kl := KeyLength{Length: l}
		for _, d := range dists {
			if d%l == 0 {
				kl.Kasiski++
			}
		}
		for _, col := range columns(s, l) {
			kl.IC += IndexOfCoincidence(col)
		}
		kl.IC /= float64(l)
		out = append(out, kl)
	}
	// Prefer the shortest length whose IC is close to English: a multiple of
	// the true length scores as well, so take the first that clears the bar.
	sort.SliceStable(out, func(i, j int) bool {
		gi, gj := out[i].IC > 0.06, out[j].IC > 0.06
		if gi != gj {
			return gi
		}
		if gi {
			return out[i].Length < out[j].Length
		}
		return out[i].IC > out[j].IC
	})
	return out
}

func columns(s string, l int) []string {
	cols := make([][]byte, l)
	for i := 0; i < len(s); i++ {
		cols[i%l] = append(cols[i%l], s[i])
	}
	out := make([]string, l)
	for i, c := range cols {
		out[i] = string(c)
	}
	return out
}

// VigenereCrack recovers the key and plaintext of a Vigenère ciphertext of
// English: it finds the key length with VigenereKeyLength, then each key
// letter as the Caesar shift that makes its column most English. It needs a
// few hundred letters to be reliable; maxLen bounds the key length tried.
//
// Time: O(n * maxLen). Space: O(n).
func VigenereCrack(ciphertext string, maxLen int) (key, plaintext string) {
	s := lettersOnly(ciphertext)
	if s == "" {
		return "", ciphertext
	}
	l := VigenereKeyLength(ciphertext, maxLen)[0].Length
	k := make([]byte, l)
	for i, col := range columns(s, l) {
		k[i] = byte('A' + CaesarBruteForce(col)[0].Shift)
	}
	return string(k), VigenereDecrypt(ciphertext, string(k))
}

// String describes a guess.
func (g Guess) String() string {
	return fmt.Sprintf("shift %2d  score %8.1f  %s", g.Shift, g.Score, g.Text)
}
