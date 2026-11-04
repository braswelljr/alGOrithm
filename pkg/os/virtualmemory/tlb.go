package virtualmemory

import "fmt"

// TLBResult is a TLB run: hit or miss for each reference, the hit rate, the
// measured average access time, and the textbook formula's.
type TLBResult struct {
	Hits      []bool   `json:"hits"`
	Contents  [][]int  `json:"contents"` // the TLB's pages after each reference
	HitRate   float64  `json:"hitRate"`
	Measured  float64  `json:"measured"`  // ns per access, over the trace
	Effective float64  `json:"effective"` // h(t+m) + (1-h)(t+(levels+1)m)
	Log       []string `json:"log"`
}

// TLB runs a trace of page numbers through a fully associative TLB of size
// entries with LRU replacement. A hit costs tlbNs plus one memory access of
// memNs; a miss also walks a levels-level page table, one memory read per
// level. The effective access time is
//
//	EAT = h * (t + m) + (1 - h) * (t + (levels + 1) * m)
//
// so with four levels, a 99% hit rate still costs 4% more than a perfect
// one: programs with locality are fast because their pages stay in the TLB.
//
// Time: O(n * size). Space: O(n * size).
func TLB(refs []int, size, tlbNs, memNs, levels int) (TLBResult, error) {
	if len(refs) == 0 || len(refs) > 500 || size < 1 || size > 64 || tlbNs < 0 || tlbNs > 100 || memNs < 1 || memNs > 1000 || levels < 1 || levels > 5 {
		return TLBResult{}, inputErr("1-500 references, 1-64 entries, TLB 0-100 ns, memory 1-1000 ns, 1-5 levels")
	}
	var res TLBResult
	var tlb []int // most recent last
	total, hits := 0, 0
	for _, p := range refs {
		at := -1
		for i, q := range tlb {
			if q == p {
				at = i
			}
		}
		hit := at >= 0
		if hit {
			hits++
			tlb = append(append(tlb[:at:at], tlb[at+1:]...), p)
			total += tlbNs + memNs
		} else {
			if len(tlb) == size {
				tlb = tlb[1:]
			}
			tlb = append(tlb, p)
			total += tlbNs + (levels+1)*memNs
		}
		res.Hits = append(res.Hits, hit)
		res.Contents = append(res.Contents, append([]int(nil), tlb...))
	}
	h := float64(hits) / float64(len(refs))
	t, m := float64(tlbNs), float64(memNs)
	res.HitRate = h
	res.Measured = float64(total) / float64(len(refs))
	res.Effective = h*(t+m) + (1-h)*(t+float64(levels+1)*m)
	res.Log = append(res.Log, fmt.Sprintf("%d hits of %d (%.0f%%); effective access time %.1f ns, against %.0f ns with no paging", hits, len(refs), 100*h, res.Effective, m))
	return res, nil
}
