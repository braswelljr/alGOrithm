package utils

import (
	"math/rand"
	"time"
)

var (
	src             = rand.NewSource(time.Now().UnixNano())                                               // random seed
	key             = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!@#$%^&*()_+{}|:<>?" // the key to be used for the hill cipher
	letterIndexBits = 6                                                                                   // 6 bits to represent a letter index
	letterIndexMask = 1<<letterIndexBits - 1                                                              // All 1-bits, as many as letterIndexBits
)

// GenerateRandomKey - generates a random key for the hill cipher
// @param n - the size of the key
// @return key - the key
func GenerateRandomKey(n int) string {
	// create a slice of bytes
	b := make([]byte, n)

	// loop through the key and add it to the key matrix
	for i, cache, remain := n-1, src.Int63(), letterIndexMask; i >= 0; {
		// if the cache is exhausted, get a new one
		if remain == 0 {
			cache, remain = src.Int63(), letterIndexMask
		} else {
			// get a random index from the key
			if idx := int(cache) & letterIndexMask; idx < len(key) {
				// add the letter to the key
				b[i] = key[idx]
				i--
			}
			// shift the cache and decrement the remain count
			cache >>= letterIndexBits
			remain--
		}

	}

	return string(b)
}
