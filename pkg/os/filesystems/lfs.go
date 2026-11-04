package filesystems

import "fmt"

// LFSResult is an LFS run: every segment's slots at the end (a block number,
// or -1 for a dead or empty slot), the log, and the write amplification.
type LFSResult struct {
	Segments [][]int  `json:"segments"`
	Log      []string `json:"log"`
	Logical  int      `json:"logical"`  // blocks the program wrote
	Physical int      `json:"physical"` // blocks the disk wrote, cleaning included
	Cleaned  int      `json:"cleaned"`
}

// LFS runs a log-structured file system (Rosenblum and Ousterhout, 1991) on
// the given number of segments of size slots each: every write, new or an overwrite, is
// appended at the head of the log, and the old copy of an overwritten block
// becomes dead. When only one free segment is left, the cleaner picks the
// segment with the fewest live blocks (greedy), copies its live blocks to
// the free segment, and frees it. Writes are always sequential, which disks
// love, but hot data overwritten often leaves cold live blocks to copy
// again and again: Physical/Logical is the write amplification.
//
// Time: O(n * segments * size) for n writes. Space: O(segments * size).
func LFS(segments, size int, writes []int) (LFSResult, error) {
	if segments < 3 || segments > 16 || size < 2 || size > 16 || len(writes) == 0 || len(writes) > 500 {
		return LFSResult{}, inputErr("3-16 segments of 2-16 blocks, 1-500 writes")
	}
	l := newLFSDisk(segments, size)
	for _, b := range writes {
		if b < 0 || b > 999 {
			return l.res, inputErr("block numbers 0-999")
		}
		l.res.Logical++
		if err := l.write(b); err != nil {
			return l.res, err
		}
	}
	l.res.Segments = l.segs
	l.res.Log = append(l.res.Log, fmt.Sprintf("%d writes cost %d disk writes (%.2fx) and %d cleanings", l.res.Logical, l.res.Physical, float64(l.res.Physical)/float64(l.res.Logical), l.res.Cleaned))
	return l.res, nil
}

type lfsLoc struct{ seg, slot int }

// lfsDisk is a log-structured disk: segments of slots holding block numbers
// (-1 for a free or dead slot), where each block's live copy is, and the
// segment the log is writing.
type lfsDisk struct {
	size  int
	segs  [][]int
	where map[int]lfsLoc
	used  []int // slots written in each segment (live or dead)
	head  int
	res   LFSResult
}

func newLFSDisk(segments, size int) *lfsDisk {
	l := &lfsDisk{size: size, segs: make([][]int, segments), where: map[int]lfsLoc{}, used: make([]int, segments)}
	for i := range l.segs {
		l.segs[i] = make([]int, size)
		for j := range l.segs[i] {
			l.segs[i][j] = -1
		}
	}
	return l
}

// write appends block b at the head of the log, killing its old copy,
// moving to a free segment (or cleaning one) when the head is full.
func (l *lfsDisk) write(b int) error {
	if l.used[l.head] == l.size {
		if err := l.advance(); err != nil {
			return err
		}
		if l.used[l.head] == l.size { // cleaning filled the new head
			return l.write(b)
		}
	}
	if old, ok := l.where[b]; ok {
		l.segs[old.seg][old.slot] = -1 // the old copy is dead
	}
	l.put(b)
	return nil
}

// put writes b into the head's next slot.
func (l *lfsDisk) put(b int) {
	l.segs[l.head][l.used[l.head]] = b
	l.where[b] = lfsLoc{l.head, l.used[l.head]}
	l.used[l.head]++
	l.res.Physical++
}

// advance moves the head to a free segment, keeping one spare; with only
// the spare left, it cleans the segment with the fewest live blocks into it.
func (l *lfsDisk) advance() error {
	f := l.free()
	switch {
	case len(f) >= 2:
		l.head = f[0]
		return nil
	case len(f) == 1:
		return l.clean(f[0])
	default:
		return inputErr("no free segment left")
	}
}

// clean copies the greedy victim's live blocks to the spare segment, which
// becomes the head, and frees the victim.
func (l *lfsDisk) clean(spare int) error {
	victim, best := -1, l.size+1
	for s := range l.segs {
		if s != l.head && s != spare && l.live(s) < best {
			victim, best = s, l.live(s)
		}
	}
	if victim < 0 || best == l.size {
		return inputErr("the disk is full of live blocks")
	}
	l.res.Cleaned++
	l.res.Log = append(l.res.Log, fmt.Sprintf("clean segment %d (%d of %d blocks live): copy them to segment %d", victim, best, l.size, spare))
	l.head = spare
	for _, lb := range l.segs[victim] {
		if lb >= 0 {
			l.put(lb)
		}
	}
	for j := range l.segs[victim] {
		l.segs[victim][j] = -1
	}
	l.used[victim] = 0
	return nil
}

func (l *lfsDisk) live(s int) int {
	n := 0
	for _, b := range l.segs[s] {
		if b >= 0 {
			n++
		}
	}
	return n
}

// free lists the empty segments other than the head.
func (l *lfsDisk) free() []int {
	var out []int
	for s := range l.segs {
		if s != l.head && l.used[s] == 0 {
			out = append(out, s)
		}
	}
	return out
}
