package filesystems

import "fmt"

// InodeLevel is one level of an inode's pointers: how many blocks it
// reaches and how many bytes that is.
type InodeLevel struct {
	Name   string `json:"name"`
	Blocks int64  `json:"blocks"`
	Bytes  int64  `json:"bytes"`
}

// InodeLayout is what Inode computes: each level, the largest file, and
// where the byte at Offset lives.
type InodeLayout struct {
	PointersPerBlock int64        `json:"pointersPerBlock"`
	Levels           []InodeLevel `json:"levels"`
	MaxBytes         int64        `json:"maxBytes"`
	Lookup           []string     `json:"lookup"`
	DiskReads        int          `json:"diskReads"` // to reach Offset, counting the inode
}

// Inode lays out a Unix inode with the given number of direct pointers and one single,
// one double and one triple indirect pointer, on blocks of blockSize bytes
// with pointers of pointerSize bytes. An indirect block holds
// p = blockSize/pointerSize pointers, so the largest file is
// (direct + p + p^2 + p^3) blocks: with 4 KiB blocks, 4-byte pointers and
// 12 direct pointers, about 4 TiB. It also walks to the block holding byte
// offset, counting the disk reads: small files are fast, large ones pay for
// the indirection.
//
// Time: O(1). Space: O(1).
func Inode(blockSize, pointerSize, direct int, offset int64) (InodeLayout, error) {
	if blockSize < 64 || blockSize > 65536 || blockSize&(blockSize-1) != 0 || (pointerSize != 4 && pointerSize != 8) || direct < 0 || direct > 32 {
		return InodeLayout{}, inputErr("block size a power of two 64-65536, pointer size 4 or 8, 0-32 direct pointers")
	}
	bs, p := int64(blockSize), int64(blockSize/pointerSize)
	l := InodeLayout{PointersPerBlock: p}
	counts := []int64{int64(direct), p, p * p, p * p * p}
	names := []string{"direct", "single indirect", "double indirect", "triple indirect"}
	for i, c := range counts {
		l.Levels = append(l.Levels, InodeLevel{Name: names[i], Blocks: c, Bytes: c * bs})
		l.MaxBytes += c * bs
	}
	if offset < 0 || offset >= l.MaxBytes {
		return l, inputErr("offset 0-%d", l.MaxBytes-1)
	}
	block := offset / bs
	l.Lookup = append(l.Lookup, fmt.Sprintf("byte %d is in file block %d (offset %d within it)", offset, block, offset%bs))
	l.DiskReads = 1
	for i, c := range counts {
		if block >= c {
			block -= c
			continue
		}
		if i == 0 {
			l.Lookup = append(l.Lookup, fmt.Sprintf("read the inode: direct pointer %d holds it", block))
		} else {
			l.Lookup = append(l.Lookup, fmt.Sprintf("read the inode: follow the %s pointer", names[i]))
			for level := i; level >= 1; level-- {
				span := int64(1)
				for k := 1; k < level; k++ {
					span *= p
				}
				l.DiskReads++
				l.Lookup = append(l.Lookup, fmt.Sprintf("read an indirect block: entry %d", block/span))
				block %= span
			}
		}
		l.DiskReads++
		l.Lookup = append(l.Lookup, "read the data block")
		break
	}
	return l, nil
}
