// Command file-integrity checks that downloaded file chunks are intact with
// SHA-256 and a Merkle tree, as BitTorrent, Git, IPFS and Certificate
// Transparency do: one root hash vouches for every chunk, and one chunk is
// checked with a short proof instead of the whole file.
package main

import (
	"fmt"
	"log"

	"github.com/braswelljr/alGOrithm/pkg/security/cryptography"
)

func main() {
	chunks := [][]byte{
		[]byte("chunk 0: header"), []byte("chunk 1: textures"),
		[]byte("chunk 2: audio"), []byte("chunk 3: levels"),
	}
	tree, err := cryptography.NewMerkleTree(chunks)
	if err != nil {
		log.Fatal(err)
	}
	root := tree.Root()
	fmt.Printf("published root hash: %s\n", cryptography.Hex(root[:8]))
	proof, err := tree.Proof(2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("chunk 2 verifies with a %d-hash proof: %v\n", len(proof), cryptography.VerifyMerkle(root, chunks[2], proof))
	fmt.Printf("tampered chunk 2 verifies: %v\n", cryptography.VerifyMerkle(root, []byte("chunk 2: malware"), proof))
	sum := cryptography.SHA256(chunks[0])
	fmt.Printf("sha256(chunk 0) = %s...\n", cryptography.Hex(sum[:8]))
}
