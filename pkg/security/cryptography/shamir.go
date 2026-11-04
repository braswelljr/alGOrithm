package cryptography

import (
	"fmt"
	"math/rand"
)

// ShamirPrime is the field Shamir shares live in: 2^31 - 1, a Mersenne prime.
const ShamirPrime = 1<<31 - 1

// Share is one point (X, Y) on the secret polynomial.
type Share struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
}

// ShamirSplit splits secret (0 to ShamirPrime-1) into n shares of which any k
// rebuild it (Shamir, 1979): the shares are points on a random polynomial of
// degree k-1 whose value at 0 is the secret. Any k-1 shares say nothing at
// all about it, since every secret fits them equally well. The coefficients
// come from a generator seeded with seed, for repeatable lessons. For
// learning: real shares need crypto/rand coefficients.
func ShamirSplit(secret int64, k, n int, seed int64) ([]Share, error) {
	if secret < 0 || secret >= ShamirPrime || k < 1 || n < k || n >= ShamirPrime {
		return nil, fmt.Errorf("%w: need 0 <= secret < 2^31-1 and 1 <= k <= n", ErrInvalid)
	}
	rng := rand.New(rand.NewSource(seed))
	coef := []int64{secret}
	for i := 1; i < k; i++ {
		coef = append(coef, rng.Int63n(ShamirPrime))
	}
	shares := make([]Share, n)
	for i := range shares {
		x, y := int64(i+1), int64(0)
		for j := len(coef) - 1; j >= 0; j-- { // Horner's rule
			y = (y*x + coef[j]) % ShamirPrime
		}
		shares[i] = Share{X: x, Y: y}
	}
	return shares, nil
}

// ShamirCombine rebuilds the secret from k or more shares by Lagrange
// interpolation at x = 0. With too few shares it returns a wrong number, not
// an error: nothing in the shares says how many are needed.
//
// Time: O(k^2 log p). Space: O(1).
func ShamirCombine(shares []Share) (int64, error) {
	seen := map[int64]bool{}
	for _, s := range shares {
		if s.X <= 0 || s.X >= ShamirPrime || seen[s.X] {
			return 0, fmt.Errorf("%w: share x values must be distinct and from 1 to 2^31-2", ErrInvalid)
		}
		seen[s.X] = true
	}
	if len(shares) == 0 {
		return 0, fmt.Errorf("%w: no shares", ErrInvalid)
	}
	var secret int64
	for i, si := range shares {
		num, den := int64(1), int64(1)
		for j, sj := range shares {
			if i != j {
				num = num * (ShamirPrime - sj.X) % ShamirPrime
				den = den * ((si.X - sj.X + ShamirPrime) % ShamirPrime) % ShamirPrime
			}
		}
		term := si.Y % ShamirPrime * num % ShamirPrime * ModPow(den, ShamirPrime-2, ShamirPrime) % ShamirPrime
		secret = (secret + term) % ShamirPrime
	}
	return secret, nil
}
