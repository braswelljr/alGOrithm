package virtualmemory

import "fmt"

// MmapResult compares reading a file with read() against mapping it.
type MmapResult struct {
	ReadCopies   int      `json:"readCopies"`   // pages copied from the page cache to the buffer
	ReadCalls    int      `json:"readCalls"`    // system calls made
	MmapFaults   int      `json:"mmapFaults"`   // page faults, one per page first touched
	MmapCopies   int      `json:"mmapCopies"`   // always 0: the pages are shared
	DiskReads    int      `json:"diskReads"`    // the same either way
	DirtyWritten int      `json:"dirtyWritten"` // pages msync writes back
	Log          []string `json:"log"`
}

// MemoryMapped reads the pages in accesses (repeats allowed) of a file of
// filePages pages, writing to the pages in writes, two ways. With read(),
// every access is a system call that copies the page from the page cache
// into the program's buffer. With mmap(), the file's page-cache pages are
// mapped into the address space: the first touch of each page faults it in,
// later touches are plain memory accesses with no copy and no system call,
// and msync writes back only the dirty pages. Both read each page from disk
// once.
//
// Time: O(len(accesses) + len(writes)). Space: O(filePages).
func MemoryMapped(filePages int, accesses, writes []int) (MmapResult, error) {
	if filePages < 1 || filePages > 256 || len(accesses) == 0 || len(accesses) > 500 {
		return MmapResult{}, inputErr("1-256 file pages, 1-500 accesses")
	}
	var res MmapResult
	cached, mapped, dirty := map[int]bool{}, map[int]bool{}, map[int]bool{}
	for _, p := range append(append([]int(nil), accesses...), writes...) {
		if p < 0 || p >= filePages {
			return res, inputErr("pages are 0-%d", filePages-1)
		}
	}
	for _, p := range accesses {
		if !cached[p] {
			cached[p] = true
			res.DiskReads++
		}
		res.ReadCalls++
		res.ReadCopies++
		if !mapped[p] {
			mapped[p] = true
			res.MmapFaults++
		}
	}
	for _, p := range writes {
		if !mapped[p] {
			mapped[p] = true
			res.MmapFaults++
			if !cached[p] {
				cached[p] = true
				res.DiskReads++
			}
		}
		dirty[p] = true
	}
	res.DirtyWritten = len(dirty)
	res.Log = append(res.Log,
		fmt.Sprintf("read(): %d system calls, %d page copies into the buffer", res.ReadCalls, res.ReadCopies),
		fmt.Sprintf("mmap(): %d page faults (first touches), no copies, no system calls per access", res.MmapFaults),
		fmt.Sprintf("both read %d pages from disk; msync writes back %d dirty pages", res.DiskReads, res.DirtyWritten))
	return res, nil
}
