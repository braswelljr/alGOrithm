package filesystems

import (
	"fmt"
	"strings"
)

// Extent is a run of free blocks in a free list.
type Extent struct {
	Start  int `json:"start"`
	Length int `json:"length"`
}

// FreeSpaceStep is the state after one call: the bitmap ("1" used) and the
// same free space as a list of extents.
type FreeSpaceStep struct {
	Call   string   `json:"call"`
	Bitmap string   `json:"bitmap"`
	List   []Extent `json:"list"`
}

// FreeSpaceResult is a FreeSpace run and the cost of each way of tracking
// free blocks.
type FreeSpaceResult struct {
	Steps []FreeSpaceStep `json:"steps"`
	// BitmapBytes is one bit per block; ListBytes is 8 bytes per extent
	// (a start and a length).
	BitmapBytes int `json:"bitmapBytes"`
	ListBytes   int `json:"listBytes"`
	// Scanned is how many bitmap bits the first-fit searches read.
	Scanned int `json:"scanned"`
}

// FreeSpace runs "alloc NAME N" (N contiguous blocks, first fit) and
// "free NAME" on a disk of blocks blocks, and after each call shows the
// free space two ways: a bitmap, one bit per block, fixed in size and easy
// to scan for runs; and a free list of extents, small when free space is in
// a few big runs and large when it is fragmented.
//
// Time: O(n * blocks). Space: O(n * blocks).
func FreeSpace(blocks int, script string) (FreeSpaceResult, error) {
	if blocks < 8 || blocks > 256 {
		return FreeSpaceResult{}, inputErr("8-256 blocks")
	}
	used := make([]bool, blocks)
	files := map[string]Extent{}
	var res FreeSpaceResult
	calls, nums := lines(script)
	for k, f := range calls {
		switch {
		case f[0] == "alloc" && len(f) == 3:
			n, ok := atoi(f[2], 1, blocks)
			if !ok {
				return res, inputErr("line %d: alloc NAME 1-%d", nums[k], blocks)
			}
			start, run := -1, 0
			for i := 0; i < blocks && start < 0; i++ {
				res.Scanned++
				if used[i] {
					run = 0
					continue
				}
				run++
				if run == n {
					start = i - n + 1
				}
			}
			if start >= 0 {
				for i := start; i < start+n; i++ {
					used[i] = true
				}
				files[f[1]] = Extent{start, n}
			}
		case f[0] == "free" && len(f) == 2:
			e, ok := files[f[1]]
			if !ok {
				return res, inputErr("line %d: no allocation named %s", nums[k], f[1])
			}
			for i := e.Start; i < e.Start+e.Length; i++ {
				used[i] = false
			}
			delete(files, f[1])
		default:
			return res, inputErr("line %d: alloc NAME N or free NAME", nums[k])
		}
		var b strings.Builder
		var list []Extent
		for i, u := range used {
			if u {
				b.WriteByte('1')
				continue
			}
			b.WriteByte('0')
			if i > 0 && !used[i-1] && len(list) > 0 {
				list[len(list)-1].Length++
			} else {
				list = append(list, Extent{i, 1})
			}
		}
		res.Steps = append(res.Steps, FreeSpaceStep{Call: strings.Join(f, " "), Bitmap: b.String(), List: list})
		res.ListBytes = 8 * len(list)
	}
	res.BitmapBytes = (blocks + 7) / 8
	return res, nil
}

// String shows the last step.
func (r FreeSpaceResult) String() string {
	if len(r.Steps) == 0 {
		return "no steps"
	}
	s := r.Steps[len(r.Steps)-1]
	return fmt.Sprintf("bitmap %s (%d bytes), free list %v (%d bytes)", s.Bitmap, r.BitmapBytes, s.List, r.ListBytes)
}
