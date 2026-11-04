// Command password-hashing stores and checks a password the way a login
// system should: never the password itself, only a slow salted hash
// (PBKDF2-HMAC-SHA256), compared in constant time.
//
// This is a teaching implementation. In production use a vetted library such
// as golang.org/x/crypto (argon2, bcrypt, scrypt or pbkdf2).
package main

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/security/cryptography"
)

func main() {
	salt := []byte("per-user-random-salt") // random and stored with the hash in practice
	const iterations = 10_000
	stored := cryptography.PBKDF2([]byte("correct horse battery staple"), salt, iterations, 32)
	fmt.Printf("stored hash: %s\n", cryptography.Hex(stored))
	for _, attempt := range []string{"correct horse battery staple", "password123"} {
		got := cryptography.PBKDF2([]byte(attempt), salt, iterations, 32)
		fmt.Printf("login with %q: %v\n", attempt, cryptography.ConstantTimeEqual(got, stored))
	}
}
