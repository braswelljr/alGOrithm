// Package filesystems shows how an operating system keeps files on a disk,
// one file per idea, as small deterministic simulations:
//
//	allocation.go  contiguous, linked (FAT) and indexed block allocation
//	inode.go       an inode's direct and indirect pointers, and the largest file
//	namespace.go   directories, hard and symbolic links, and path resolution
//	freespace.go   tracking free blocks: a bitmap against a free list
//	journal.go     journaling: crashing at every point, and recovering
//	pagecache.go   the page cache, write-through against write-back
//	raid.go        RAID 0, 1, 5 and 6: striping, mirroring and parity
//	lfs.go         a log-structured file system and its cleaner
package filesystems

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ErrInput is wrapped by every error about invalid input.
var ErrInput = errors.New("filesystems: invalid input")

func inputErr(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInput, fmt.Sprintf(format, args...))
}

// lines splits a script into its non-empty lines' fields, numbered from 1.
func lines(script string) (out [][]string, nums []int) {
	for n, l := range strings.Split(script, "\n") {
		if f := strings.Fields(l); len(f) > 0 {
			out = append(out, f)
			nums = append(nums, n+1)
		}
	}
	return out, nums
}

func atoi(s string, lo, hi int) (int, bool) {
	v, err := strconv.Atoi(s)
	return v, err == nil && v >= lo && v <= hi
}
