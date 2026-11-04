package memory

import (
	"errors"
	"fmt"
	"math/bits"
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// ErrNoSpace is returned when no free block is big enough for a request.
var ErrNoSpace = errors.New("memory: no free block is big enough")

// Block is a block of a buddy arena: where it starts, its size, and who holds
// it ("" when free) with the size they asked for.
type Block struct {
	Offset, Size int
	Owner        string
	Used         int
}

// Buddy is a buddy allocator (Knowlton, 1965) over a power-of-two arena:
// a request is rounded up to a power of two, and a bigger free block is split
// in halves ("buddies") until one fits. Freeing a block merges it with its
// buddy whenever the buddy is free too, so free memory stays in large blocks.
// The internal fragmentation is the rounding up.
type Buddy struct {
	size, min int
	free      map[int][]int // block size -> free offsets
	owners    map[string]Block
}

// NewBuddy returns an empty arena of size bytes whose smallest block is
// minBlock; both must be powers of two, minBlock at most size, and size at most
// 2^30.
func NewBuddy(size, minBlock int) (*Buddy, error) {
	if !pow2(size) || !pow2(minBlock) || minBlock > size || size > 1<<30 {
		return nil, fmt.Errorf("%w: size and smallest block must be powers of two, the block no bigger", ErrInput)
	}
	return &Buddy{size: size, min: minBlock, free: map[int][]int{size: {0}}, owners: map[string]Block{}}, nil
}

func pow2(n int) bool { return n > 0 && n&(n-1) == 0 }

// Alloc gives name a block of at least size bytes and returns where it
// starts. Time: O(log(size / minBlock)).
func (b *Buddy) Alloc(name string, size int) (int, error) {
	if _, ok := b.owners[name]; ok || name == "" {
		return 0, fmt.Errorf("%w: %q needs a new, non-empty name", ErrInput, name)
	}
	if size < 1 || size > b.size {
		return 0, fmt.Errorf("%w: size %d is outside 1..%d", ErrInput, size, b.size)
	}
	want := max(b.min, 1<<bits.Len(uint(size-1)))
	have := want
	for have <= b.size && len(b.free[have]) == 0 {
		have *= 2
	}
	if have > b.size {
		return 0, ErrNoSpace
	}
	off := b.take(have)
	for have > want { // split, keeping the lower half and freeing the upper
		have /= 2
		b.free[have] = append(b.free[have], off+have)
	}
	b.owners[name] = Block{Offset: off, Size: want, Owner: name, Used: size}
	return off, nil
}

// take removes the lowest free offset of size.
func (b *Buddy) take(size int) int {
	list := b.free[size]
	sort.Ints(list)
	off := list[0]
	b.free[size] = list[1:]
	return off
}

// Free returns name's block, merging it with its buddy while the buddy is
// free. Time: O(log(size / minBlock) * free blocks of a size).
func (b *Buddy) Free(name string) error {
	blk, ok := b.owners[name]
	if !ok {
		return fmt.Errorf("%w: %q holds no block", ErrInput, name)
	}
	delete(b.owners, name)
	off, size := blk.Offset, blk.Size
	for size < b.size {
		buddy := off ^ size
		i := index(b.free[size], buddy)
		if i < 0 {
			break
		}
		b.free[size] = append(b.free[size][:i], b.free[size][i+1:]...)
		off, size = min(off, buddy), size*2
	}
	b.free[size] = append(b.free[size], off)
	return nil
}

func index(s []int, v int) int {
	for i, x := range s {
		if x == v {
			return i
		}
	}
	return -1
}

// Blocks returns every block, free and used, in address order.
func (b *Buddy) Blocks() []Block {
	var out []Block
	for size, offs := range b.free {
		for _, o := range offs {
			out = append(out, Block{Offset: o, Size: size})
		}
	}
	for _, blk := range b.owners {
		out = append(out, blk)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Offset < out[j].Offset })
	return out
}

// Fragmentation is the bytes held but not asked for (internal fragmentation).
func (b *Buddy) Fragmentation() int {
	n := 0
	for _, blk := range b.owners {
		n += blk.Size - blk.Used
	}
	return n
}

// Snapshot describes the arena as a row of blocks in address order, used
// blocks labeled "name: used/size", free blocks with the role "free".
func (b *Buddy) Snapshot() trace.Snapshot {
	blocks := b.Blocks()
	s := trace.Snapshot{Layout: "row", Nodes: make([]trace.Node, len(blocks))}
	for i, blk := range blocks {
		n := trace.Node{ID: i, Parent: -1, Label: fmt.Sprintf("%d: free %d", blk.Offset, blk.Size), Role: "free"}
		if blk.Owner != "" {
			n.Label, n.Role = fmt.Sprintf("%d: %s %d/%d", blk.Offset, blk.Owner, blk.Used, blk.Size), "used"
		}
		s.Nodes[i] = n
	}
	return s
}
