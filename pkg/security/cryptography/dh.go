package cryptography

import "fmt"

// ModPow returns base^exp mod m by square-and-multiply: one squaring per bit
// of exp, and a multiplication for each 1 bit. m must be from 2 to 2^31 so
// the products fit an int64.
//
// Time: O(log exp). Space: O(1).
func ModPow(base, exp, m int64) int64 {
	result, b := int64(1), (base%m+m)%m
	for ; exp > 0; exp >>= 1 {
		if exp&1 == 1 {
			result = result * b % m
		}
		b = b * b % m
	}
	return result
}

// DHExchange is what each side of a Diffie-Hellman exchange sees.
type DHExchange struct {
	// A and B are the public values g^a and g^b mod p, sent in the open.
	A, B int64
	// Shared is g^(ab) mod p, which both sides compute and an eavesdropper,
	// who sees only p, g, A and B, cannot without solving a discrete log.
	Shared int64
}

// DiffieHellman runs the 1976 key exchange in the small group mod p with
// generator g and secrets a and b. Small groups fall to brute force at once;
// real exchanges use 2048-bit groups or elliptic curves (X25519), and need
// authentication against a man in the middle. For learning: use crypto/ecdh
// in real code.
func DiffieHellman(p, g, a, b int64) (DHExchange, error) {
	if err := checkGroup(p, g); err != nil {
		return DHExchange{}, err
	}
	if a < 1 || b < 1 {
		return DHExchange{}, fmt.Errorf("%w: the secrets must be positive", ErrInvalid)
	}
	pa, pb := ModPow(g, a, p), ModPow(g, b, p)
	// Each side raises the other's public value to its own secret:
	// (g^b)^a = (g^a)^b.
	return DHExchange{A: pa, B: pb, Shared: ModPow(pb, a, p)}, nil
}

func checkGroup(p, g int64) error {
	if p < 3 || p > 1<<31 || !isPrime(p) {
		return fmt.Errorf("%w: p must be a prime from 3 to 2^31", ErrInvalid)
	}
	if g < 2 || g >= p {
		return fmt.Errorf("%w: g must be from 2 to p-1", ErrInvalid)
	}
	return nil
}

func isPrime(n int64) bool {
	if n < 2 {
		return false
	}
	for d := int64(2); d*d <= n; d++ {
		if n%d == 0 {
			return false
		}
	}
	return true
}
