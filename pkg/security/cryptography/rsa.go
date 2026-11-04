package cryptography

import (
	"fmt"
	"math/big"
	"math/rand"
)

// RSAKey is an RSA key pair: the public key is (N, E), the private key D.
type RSAKey struct {
	N, E, D *big.Int
	P, Q    *big.Int
}

// GenerateRSA makes an RSA key whose modulus has the given number of bits (64 to
// 4096) from two primes found by a generator seeded with seed, so the same
// seed gives the same key: good for a lesson, fatal in real use, where the
// primes must come from crypto/rand. E is 65537, D its inverse mod
// (P-1)(Q-1). For learning: use crypto/rsa in real code.
func GenerateRSA(bits int, seed int64) (*RSAKey, error) {
	if bits < 64 || bits > 4096 {
		return nil, fmt.Errorf("%w: RSA key size %d (64 to 4096 bits)", ErrInvalid, bits)
	}
	rng := rand.New(rand.NewSource(seed))
	e := big.NewInt(65537)
	one := big.NewInt(1)
	for {
		p, q := seededPrime(rng, bits/2), seededPrime(rng, bits-bits/2)
		if p.Cmp(q) == 0 {
			continue
		}
		phi := new(big.Int).Mul(new(big.Int).Sub(p, one), new(big.Int).Sub(q, one))
		d := new(big.Int).ModInverse(e, phi)
		if d == nil { // e shares a factor with phi
			continue
		}
		return &RSAKey{N: new(big.Int).Mul(p, q), E: e, D: d, P: p, Q: q}, nil
	}
}

// seededPrime returns a probable prime of exactly bits bits: random odd
// numbers with the top two bits set (so p*q has the full length), tested by
// Miller-Rabin.
func seededPrime(rng *rand.Rand, bits int) *big.Int {
	b := make([]byte, (bits+7)/8)
	for {
		_, _ = rng.Read(b) // math/rand.Read never fails
		p := new(big.Int).SetBytes(b)
		p.SetBit(p, bits-1, 1).SetBit(p, bits-2, 1).SetBit(p, 0, 1)
		for i := len(b) * 8; i >= bits; i-- {
			p.SetBit(p, i, 0)
		}
		if p.ProbablyPrime(20) {
			return p
		}
	}
}

// Encrypt returns m^E mod N, textbook RSA. Textbook RSA is deterministic (the
// same m always gives the same ciphertext, so guesses can be checked) and
// malleable (multiplying ciphertexts multiplies plaintexts), and a small m
// with a small E can be undone by an ordinary root. Real RSA pads the
// message first with random bytes (OAEP), which closes all three. For
// learning: use crypto/rsa.EncryptOAEP in real code.
func (k *RSAKey) Encrypt(m *big.Int) (*big.Int, error) {
	if m.Sign() < 0 || m.Cmp(k.N) >= 0 {
		return nil, fmt.Errorf("%w: the message must be from 0 to N-1", ErrInvalid)
	}
	return new(big.Int).Exp(m, k.E, k.N), nil
}

// Decrypt returns c^D mod N, undoing Encrypt because E*D = 1 mod (P-1)(Q-1)
// (Euler's theorem).
func (k *RSAKey) Decrypt(c *big.Int) *big.Int { return new(big.Int).Exp(c, k.D, k.N) }

// Sign returns the textbook signature m^D mod N: only the holder of D can
// make it, anyone with (N, E) can check it. Signing a raw m lets an attacker
// combine signatures into new ones; real signatures hash and pad the message
// (PSS). For learning: use crypto/rsa.SignPSS in real code.
func (k *RSAKey) Sign(m *big.Int) *big.Int { return k.Decrypt(m) }

// Verify reports whether sig is the textbook signature of m: sig^E mod N == m.
func (k *RSAKey) Verify(m, sig *big.Int) bool {
	return new(big.Int).Exp(sig, k.E, k.N).Cmp(new(big.Int).Mod(m, k.N)) == 0
}
