package cryptography

import (
	"encoding/hex"
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// MerkleTree is a binary tree of hashes over a list of data blocks: each
// leaf is the hash of a block, each parent the hash of its two children, so
// the root commits to every block. A leaf with no partner moves up as it is.
// Leaves and parents hash with different prefixes (0 and 1, as in RFC 6962)
// so a parent can never pass for a leaf. For learning: Git, Bitcoin and
// Certificate Transparency use the same idea.
type MerkleTree struct {
	// Levels[0] are the leaf hashes; the last level holds just the root.
	Levels [][][32]byte
}

// ProofStep is one sibling hash on the way from a leaf to the root.
type ProofStep struct {
	Hash [32]byte `json:"hash"`
	// Left is true when the sibling is on the left.
	Left bool `json:"left"`
}

func leafHash(d []byte) [32]byte { return SHA256(append([]byte{0}, d...)) }

func nodeHash(l, r [32]byte) [32]byte {
	return SHA256(append(append([]byte{1}, l[:]...), r[:]...))
}

// NewMerkleTree builds the tree over blocks (at least one).
//
// Time: O(n). Space: O(n).
func NewMerkleTree(blocks [][]byte) (*MerkleTree, error) {
	if len(blocks) == 0 {
		return nil, fmt.Errorf("%w: a Merkle tree needs at least one block", ErrInvalid)
	}
	level := make([][32]byte, len(blocks))
	for i, b := range blocks {
		level[i] = leafHash(b)
	}
	t := &MerkleTree{Levels: [][][32]byte{level}}
	for len(level) > 1 {
		next := make([][32]byte, 0, (len(level)+1)/2)
		for i := 0; i < len(level); i += 2 {
			if i+1 == len(level) {
				next = append(next, level[i])
			} else {
				next = append(next, nodeHash(level[i], level[i+1]))
			}
		}
		t.Levels = append(t.Levels, next)
		level = next
	}
	return t, nil
}

// Root returns the root hash.
func (t *MerkleTree) Root() [32]byte { return t.Levels[len(t.Levels)-1][0] }

// Proof returns the sibling hashes from leaf i up to the root: O(log n) of
// them, enough for anyone holding only the root to check that block i is in
// the tree.
func (t *MerkleTree) Proof(i int) ([]ProofStep, error) {
	if i < 0 || i >= len(t.Levels[0]) {
		return nil, fmt.Errorf("%w: no leaf %d", ErrInvalid, i)
	}
	var proof []ProofStep
	for _, level := range t.Levels[:len(t.Levels)-1] {
		if sib := i ^ 1; sib < len(level) {
			proof = append(proof, ProofStep{Hash: level[sib], Left: sib < i})
		}
		i /= 2
	}
	return proof, nil
}

// VerifyMerkle reports whether block, hashed up through proof, gives root.
func VerifyMerkle(root [32]byte, block []byte, proof []ProofStep) bool {
	h := leafHash(block)
	for _, s := range proof {
		if s.Left {
			h = nodeHash(s.Hash, h)
		} else {
			h = nodeHash(h, s.Hash)
		}
	}
	return h == root
}

// Snapshot draws the tree with each node labeled by the first four bytes of
// its hash in hex.
func (t *MerkleTree) Snapshot() trace.Snapshot {
	s := trace.Snapshot{Layout: "tree"}
	top := len(t.Levels) - 1
	// IDs number the nodes level by level from the root down.
	ids := make([][]int, len(t.Levels))
	next := 0
	for l := top; l >= 0; l-- {
		ids[l] = make([]int, len(t.Levels[l]))
		for i := range t.Levels[l] {
			ids[l][i] = next
			next++
		}
	}
	for l := top; l >= 0; l-- {
		for i, h := range t.Levels[l] {
			n := trace.Node{ID: ids[l][i], Label: hex.EncodeToString(h[:4]), Parent: -1}
			switch l {
			case top:
				n.Role = "root"
			case 0:
				n.Role = "leaf"
			}
			if l < top {
				n.Parent = ids[l+1][i/2]
				s.Edges = append(s.Edges, trace.Edge{From: n.Parent, To: n.ID})
			}
			s.Nodes = append(s.Nodes, n)
		}
	}
	return s
}
