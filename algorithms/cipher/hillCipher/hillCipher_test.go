package hillCipher

import "testing"

var (
	size      = 15                // the size of the key matrix
	keyMatrix = GenerateKey(size) // the key matrix
)

func TestGenerateKey(t *testing.T) {
	// `GenerateKey` should return a key matrix for the hill cipher given a key and the size of the key matrix`
	if len(keyMatrix) != size {
		t.Error("Key matrix should have a length of ", size)
		return
	}

	for i := 0; i < size; i++ {
		if len(keyMatrix[i]) != size {
			t.Error("Key matrix should have a length of ", size)
			return
		}
	}

	t.Log("Key matrix generated successfully!")
	t.Log(keyMatrix)
}
