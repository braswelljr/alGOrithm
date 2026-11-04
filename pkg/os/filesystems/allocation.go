package filesystems

import (
	"fmt"
	"strings"
)

// Method is a way to allocate a file's blocks.
type Method string

// The three classic allocation methods.
const (
	Contiguous Method = "contiguous" // one run of adjacent blocks: fast, but fragments
	Linked     Method = "linked"     // any free blocks, chained in a table (FAT)
	Indexed    Method = "indexed"    // any free blocks, listed in an index block (an inode)
)

// AllocResult is an Allocate run: the owner of every block after each
// step ("" for free, "x*" for x's index block), the FAT for Linked (next
// block, -1 for the end of a file, -2 for free), each file's blocks, and the
// log.
type AllocResult struct {
	Maps   [][]string       `json:"maps"`
	FAT    []int            `json:"fat,omitempty"`
	Files  map[string][]int `json:"files"`
	Log    []string         `json:"log"`
	Failed int              `json:"failed"`
}

// Allocate runs a script on a disk of blocks blocks, one call a line:
// "create NAME SIZE" (SIZE blocks) and "delete NAME". Contiguous takes the
// first run of SIZE free blocks, so after some deletes a file can fail to
// fit although enough blocks are free (external fragmentation). Linked
// takes any free blocks and chains them in the file allocation table, so it
// always fits but reaching block k means following k links. Indexed takes
// one extra block for the index, then any free blocks, so any block is one
// lookup away.
//
// Time: O(n * blocks) for n calls. Space: O(n * blocks) for the maps.
func Allocate(method Method, blocks int, script string) (AllocResult, error) {
	if method != Contiguous && method != Linked && method != Indexed {
		return AllocResult{}, inputErr("unknown method %q", method)
	}
	if blocks < 4 || blocks > 64 {
		return AllocResult{}, inputErr("4-64 blocks")
	}
	fa := &fileAlloc{method: method, disk: make([]string, blocks), res: AllocResult{Files: map[string][]int{}}}
	if method == Linked {
		fa.res.FAT = make([]int, blocks)
		for i := range fa.res.FAT {
			fa.res.FAT[i] = -2
		}
	}
	calls, nums := lines(script)
	for k, f := range calls {
		var err error
		switch {
		case f[0] == "create" && len(f) == 3:
			err = fa.create(f[1], f[2], nums[k])
		case f[0] == "delete" && len(f) == 2:
			err = fa.remove(f[1], nums[k])
		default:
			err = inputErr("line %d: create NAME SIZE or delete NAME", nums[k])
		}
		if err != nil {
			return fa.res, err
		}
		fa.res.Maps = append(fa.res.Maps, append([]string(nil), fa.disk...))
	}
	return fa.res, nil
}

// fileAlloc is a disk being allocated: each block's owner ("" when free).
type fileAlloc struct {
	method Method
	disk   []string
	res    AllocResult
}

// create allocates a new file of size blocks, or logs why it does not fit.
func (fa *fileAlloc) create(name, sizeText string, line int) error {
	size, ok := atoi(sizeText, 1, len(fa.disk))
	if !ok || fa.res.Files[name] != nil || strings.Contains(name, "*") {
		return inputErr("line %d: create NAME SIZE, a new name and 1-%d blocks", line, len(fa.disk))
	}
	got := allocate(fa.method, fa.disk, name, size)
	if got == nil {
		fa.res.Failed++
		why := ""
		if free := fa.free(); free >= size+boolInt(fa.method == Indexed) {
			why = fmt.Sprintf(" although %d blocks are free: no run of %d is (external fragmentation)", free, size)
		}
		fa.res.Log = append(fa.res.Log, fmt.Sprintf("create %s (%d blocks): does not fit%s", name, size, why))
		return nil
	}
	fa.res.Files[name] = got
	if fa.method == Linked {
		for i, b := range got {
			fa.res.FAT[b] = -1
			if i+1 < len(got) {
				fa.res.FAT[b] = got[i+1]
			}
		}
	}
	fa.res.Log = append(fa.res.Log, fmt.Sprintf("create %s: blocks %v", name, got))
	return nil
}

// remove frees a file's blocks, and its index block ("NAME*") if it has one.
func (fa *fileAlloc) remove(name string, line int) error {
	got, ok := fa.res.Files[name]
	if !ok {
		return inputErr("line %d: no file %s", line, name)
	}
	for _, b := range got {
		fa.disk[b] = ""
		if fa.method == Linked {
			fa.res.FAT[b] = -2
		}
	}
	for i, o := range fa.disk {
		if o == name+"*" {
			fa.disk[i] = ""
		}
	}
	delete(fa.res.Files, name)
	fa.res.Log = append(fa.res.Log, fmt.Sprintf("delete %s: blocks %v are free again", name, got))
	return nil
}

func (fa *fileAlloc) free() int {
	n := 0
	for _, o := range fa.disk {
		if o == "" {
			n++
		}
	}
	return n
}

// allocate claims size blocks for name, or returns nil if they do not fit.
func allocate(method Method, disk []string, name string, size int) []int {
	if method == Contiguous {
		run := 0
		for i, o := range disk {
			if o != "" {
				run = 0
				continue
			}
			run++
			if run == size {
				var got []int
				for b := i - size + 1; b <= i; b++ {
					disk[b] = name
					got = append(got, b)
				}
				return got
			}
		}
		return nil
	}
	var free []int
	for i, o := range disk {
		if o == "" {
			free = append(free, i)
		}
	}
	need := size
	if method == Indexed {
		need++
	}
	if len(free) < need {
		return nil
	}
	if method == Indexed {
		disk[free[0]] = name + "*"
		free = free[1:]
	}
	for _, b := range free[:size] {
		disk[b] = name
	}
	return append([]int(nil), free[:size]...)
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
