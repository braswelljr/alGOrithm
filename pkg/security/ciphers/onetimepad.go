package ciphers

import (
	"encoding/hex"
	"errors"
)

// ErrPadTooShort is returned when a one-time pad is shorter than the message.
var ErrPadTooShort = errors.New("ciphers: the pad is shorter than the message")

// OneTimePad XORs message with pad byte by byte and returns the result as
// hex. With a truly random pad as long as the message, used once and kept
// secret, it is the only cipher proven unbreakable (Shannon, 1949); reusing
// the pad breaks it at once, since XORing two ciphertexts cancels the pad.
// Decrypt with OneTimePadDecrypt and the same pad.
//
// Time: O(n). Space: O(n).
func OneTimePad(message, pad []byte) (string, error) {
	out, err := xor(message, pad)
	return hex.EncodeToString(out), err
}

// OneTimePadDecrypt reverses OneTimePad: it decodes the hex and XORs it with
// pad again.
func OneTimePadDecrypt(cipherHex string, pad []byte) ([]byte, error) {
	c, err := hex.DecodeString(cipherHex)
	if err != nil {
		return nil, err
	}
	return xor(c, pad)
}

func xor(a, pad []byte) ([]byte, error) {
	if len(pad) < len(a) {
		return nil, ErrPadTooShort
	}
	out := make([]byte, len(a))
	for i := range a {
		out[i] = a[i] ^ pad[i]
	}
	return out, nil
}
