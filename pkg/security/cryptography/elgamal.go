package cryptography

import "fmt"

// ElGamalCiphertext is the pair (g^k, m * y^k) mod p.
type ElGamalCiphertext struct {
	C1 int64 `json:"c1"`
	C2 int64 `json:"c2"`
}

// ElGamalEncrypt encrypts m (1 to p-1) for the holder of the secret x whose
// public key is y = g^x mod p, with the one-time random k: the shared value
// y^k masks m, and g^k lets the holder rebuild it. Reusing k for two messages
// reveals their ratio. For learning: ElGamal survives as the basis of DSA
// and of elliptic-curve schemes; use crypto/ecdsa or crypto/ed25519 in real
// code.
func ElGamalEncrypt(p, g, y, m, k int64) (ElGamalCiphertext, error) {
	if err := checkGroup(p, g); err != nil {
		return ElGamalCiphertext{}, err
	}
	if m < 1 || m >= p || k < 1 || k >= p-1 || y < 1 || y >= p {
		return ElGamalCiphertext{}, fmt.Errorf("%w: m and y must be from 1 to p-1, k from 1 to p-2", ErrInvalid)
	}
	return ElGamalCiphertext{C1: ModPow(g, k, p), C2: m * ModPow(y, k, p) % p}, nil
}

// ElGamalDecrypt recovers m from c with the secret x: m = C2 / C1^x mod p,
// dividing by multiplying with the inverse s^(p-2) (Fermat's little theorem).
func ElGamalDecrypt(p, x int64, c ElGamalCiphertext) int64 {
	s := ModPow(c.C1, x, p)
	return c.C2 * ModPow(s, p-2, p) % p
}
