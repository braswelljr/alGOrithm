package virtualmemory

import (
	"fmt"
	"strconv"
	"strings"
)

// Walk is a PageWalk result: how the address splits, each step of the walk,
// the physical address (or -1 on a page fault), and what the tables cost.
type Walk struct {
	Indices  []int    `json:"indices"` // one per level, then the offset
	Steps    []string `json:"steps"`
	Physical int64    `json:"physical"`
	// Tables is how many page-table pages the mapped pages need, against the
	// entries a flat (one-level) table would need for the whole space.
	Tables      int   `json:"tables"`
	FlatEntries int64 `json:"flatEntries"`
	MemoryReads int   `json:"memoryReads"` // to translate, without a TLB
}

// PageWalk translates addr with a page table of len(bits) levels, where
// bits[i] is the number of index bits at level i, and offsetBits the bits of
// the page offset (bits {10, 10} with 12 offset bits is 32-bit x86; four
// levels of 9 is x86-64). The mapped pages are given one a line,
// "PAGE FRAME" (the virtual page number, then the frame). Every level costs
// a memory read, which is why the TLB matters; but only the tables for
// mapped regions exist, which is why a sparse 64-bit address space fits in
// memory at all.
//
// Time: O(levels + mappings). Space: O(mappings * levels).
func PageWalk(mappings string, addr int64, bits []int, offsetBits int) (Walk, error) {
	total := offsetBits
	for _, b := range bits {
		if b < 1 || b > 12 {
			return Walk{}, inputErr("each level has 1-12 index bits")
		}
		total += b
	}
	if len(bits) < 1 || len(bits) > 5 || offsetBits < 4 || offsetBits > 16 || total > 62 {
		return Walk{}, inputErr("1-5 levels, 4-16 offset bits, at most 62 bits in all")
	}
	if addr < 0 || addr >= int64(1)<<total {
		return Walk{}, inputErr("the address must fit in %d bits", total)
	}
	table := map[int64]int64{}
	for n, line := range strings.Split(mappings, "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		if len(f) != 2 {
			return Walk{}, inputErr("line %d: want \"PAGE FRAME\"", n+1)
		}
		p, err1 := strconv.ParseInt(f[0], 0, 64)
		fr, err2 := strconv.ParseInt(f[1], 0, 64)
		if err1 != nil || err2 != nil || p < 0 || fr < 0 || p >= int64(1)<<(total-offsetBits) {
			return Walk{}, inputErr("line %d: a page in range and a frame", n+1)
		}
		table[p] = fr
	}
	w := Walk{Physical: -1, FlatEntries: int64(1) << (total - offsetBits)}
	// Count the table pages: one root, plus one per distinct prefix at each level.
	w.Tables = 1
	for level := 1; level < len(bits); level++ {
		shift := 0
		for _, b := range bits[level:] {
			shift += b
		}
		prefixes := map[int64]bool{}
		for p := range table {
			prefixes[p>>shift] = true
		}
		w.Tables += len(prefixes)
	}
	vpn, offset := addr>>offsetBits, addr&(int64(1)<<offsetBits-1)
	shift := total - offsetBits
	for _, b := range bits {
		shift -= b
		w.Indices = append(w.Indices, int((vpn>>shift)&(int64(1)<<b-1)))
	}
	w.Indices = append(w.Indices, int(offset))
	w.Steps = append(w.Steps, fmt.Sprintf("address %#x = page %d, offset %d; index bits %v split it into %v", addr, vpn, offset, bits, w.Indices[:len(bits)]))
	// Walk: a level's entry exists if any mapped page shares the prefix so far.
	shift = total - offsetBits
	for i, b := range bits {
		shift -= b
		prefix := vpn >> shift
		w.MemoryReads++
		present := false
		for p := range table {
			if p>>shift == prefix {
				present = true
				break
			}
		}
		if !present {
			w.Steps = append(w.Steps, fmt.Sprintf("level %d: entry %d is not present: page fault", i+1, w.Indices[i]))
			return w, nil
		}
		if i < len(bits)-1 {
			w.Steps = append(w.Steps, fmt.Sprintf("level %d: read entry %d, which points to a level %d table", i+1, w.Indices[i], i+2))
		}
	}
	frame := table[vpn]
	w.Physical = frame<<offsetBits | offset
	w.Steps = append(w.Steps, fmt.Sprintf("level %d: entry %d holds frame %d: physical address %#x", len(bits), w.Indices[len(bits)-1], frame, w.Physical))
	w.MemoryReads++ // and the access itself
	return w, nil
}
