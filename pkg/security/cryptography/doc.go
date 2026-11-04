// Package cryptography shows how modern cryptography works inside, written
// from scratch with the standard library only: hash functions (SHA-256,
// MD5), message authentication (HMAC) and key stretching (PBKDF2), a block
// cipher (AES-128) and a stream cipher (ChaCha20), public-key systems (RSA,
// Diffie-Hellman, ElGamal, elliptic curves), Shamir secret sharing, Merkle
// trees, a toy Feistel network, an LFSR, the Base64, Base32 and hex
// encodings, and constant-time comparison.
//
// Every function here is for learning. The tests check each one against the
// standard library or the official test vectors, but none of them resists
// timing attacks, and some (MD5, textbook RSA, the small groups) are broken
// on purpose. Use crypto/... and golang.org/x/crypto in real code.
package cryptography

import "errors"

// ErrInvalid is returned for input an algorithm cannot take: a key of the
// wrong length, a message too large for the modulus, bad hex.
var ErrInvalid = errors.New("cryptography: invalid input")
