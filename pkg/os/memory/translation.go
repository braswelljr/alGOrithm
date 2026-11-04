package memory

import (
	"fmt"
	"strconv"
	"strings"
)

// Translation is one virtual address translated.
type Translation struct {
	Virtual  int  `json:"virtual"`
	Page     int  `json:"page"`
	Offset   int  `json:"offset"`
	Frame    int  `json:"frame"`    // -1 on a page fault
	Physical int  `json:"physical"` // -1 on a page fault
	Fault    bool `json:"fault"`
	TLBHit   bool `json:"tlbHit"`
}

// TranslationResult is a whole run of address translation.
type TranslationResult struct {
	Steps   []Translation `json:"steps"`
	Faults  int           `json:"faults"`
	TLBHits int           `json:"tlbHits"`
}

// ParsePageTable reads a page table, one mapping a line: "page frame". Pages
// not listed are not in memory, so touching them is a page fault.
func ParsePageTable(text string) (map[int]int, error) {
	table := map[int]int{}
	for i, line := range strings.Split(text, "\n") {
		f := strings.FieldsFunc(line, func(r rune) bool { return r == ' ' || r == '\t' || r == ',' || r == '\r' || r == ':' })
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		p, err1 := strconv.Atoi(f[0])
		fr, err2 := strconv.Atoi(f[len(f)-1])
		if len(f) != 2 || err1 != nil || err2 != nil || p < 0 || fr < 0 {
			return nil, fmt.Errorf("%w: line %d: want \"page frame\"", ErrInput, i+1)
		}
		table[p] = fr
	}
	return table, nil
}

// Translate turns virtual addresses into physical ones with a single-level
// page table: page = address / pageSize, offset = address % pageSize, physical
// = frame * pageSize + offset. A page not in the table is a page fault. With
// tlbSize > 0, a translation lookaside buffer of that many entries (LRU) is
// consulted first, and a hit skips the page table.
//
// Time: O(n * tlbSize). Space: O(n).
func Translate(pageSize int, table map[int]int, addrs []int, tlbSize int) (TranslationResult, error) {
	if !pow2(pageSize) || pageSize > 1<<20 {
		return TranslationResult{}, fmt.Errorf("%w: the page size must be a power of two up to 2^20", ErrInput)
	}
	if len(addrs) == 0 || len(addrs) > MaxReferences || tlbSize < 0 || tlbSize > MaxFrames {
		return TranslationResult{}, fmt.Errorf("%w: give 1 to %d addresses and a TLB of 0 to %d entries", ErrInput, MaxReferences, MaxFrames)
	}
	var tlb []int // pages, most recently used first
	var r TranslationResult
	for _, a := range addrs {
		if a < 0 {
			return TranslationResult{}, fmt.Errorf("%w: address %d is negative", ErrInput, a)
		}
		t := Translation{Virtual: a, Page: a / pageSize, Offset: a % pageSize, Frame: -1, Physical: -1}
		if i := index(tlb, t.Page); i >= 0 {
			t.TLBHit = true
			r.TLBHits++
			tlb = append(tlb[:i], tlb[i+1:]...)
		}
		if f, ok := table[t.Page]; ok {
			t.Frame, t.Physical = f, f*pageSize+t.Offset
			if tlbSize > 0 {
				tlb = append([]int{t.Page}, tlb...)
				if len(tlb) > tlbSize {
					tlb = tlb[:tlbSize]
				}
			}
		} else {
			t.Fault = true
			r.Faults++
		}
		r.Steps = append(r.Steps, t)
	}
	return r, nil
}
