package virtualmemory

import (
	"fmt"
	"strconv"
	"strings"
)

// InvertedEntry is one frame of an inverted page table.
type InvertedEntry struct {
	PID  int `json:"pid"`
	Page int `json:"page"`
}

// InvertedResult is an InvertedPageTable run: the table after each
// reference (PID -1 for a free frame), the probes each lookup took, and the
// faults.
type InvertedResult struct {
	Tables [][]InvertedEntry `json:"tables"`
	Probes []int             `json:"probes"`
	Faults int               `json:"faults"`
	Log    []string          `json:"log"`
}

// InvertedPageTable runs references "PID:PAGE" (separated by spaces) on
// the given number of frames with an inverted page table, as on the PowerPC and IA-64:
// one entry per physical frame saying which process's page is in it, rather
// than one table per process. That keeps the table the size of physical
// memory however large the address spaces are, but a lookup must search it;
// a hash of (PID, page) points at the first frame to probe, and collisions
// probe onward (linear probing). Faults load the page into the next frame
// in FIFO order.
//
// Time: O(n * frames) worst case, O(n) on average. Space: O(n * frames).
func InvertedPageTable(frames int, refs string) (InvertedResult, error) {
	if n := len(strings.Fields(refs)); frames < 1 || frames > 32 || n == 0 || n > 300 {
		return InvertedResult{}, inputErr("1-32 frames and 1-300 references")
	}
	table := make([]InvertedEntry, frames)
	for i := range table {
		table[i] = InvertedEntry{-1, -1}
	}
	var res InvertedResult
	victim := 0
	for _, r := range strings.Fields(refs) {
		ps, pg, ok := strings.Cut(r, ":")
		pid, err1 := strconv.Atoi(ps)
		page, err2 := strconv.Atoi(pg)
		if !ok || err1 != nil || err2 != nil || pid < 0 || page < 0 {
			return res, inputErr("reference %q: want PID:PAGE", r)
		}
		start := (pid*31 + page) % frames
		probes, found := 0, -1
		for k := 0; k < frames; k++ {
			probes++
			f := (start + k) % frames
			if table[f] == (InvertedEntry{pid, page}) {
				found = f
				break
			}
		}
		if found >= 0 {
			res.Log = append(res.Log, fmt.Sprintf("%s: hash says frame %d, found in frame %d after %d probes", r, start, found, probes))
		} else {
			res.Faults++
			old := table[victim]
			table[victim] = InvertedEntry{pid, page}
			msg := fmt.Sprintf("%s: not in the table after %d probes: page fault, load into frame %d", r, probes, victim)
			if old.PID >= 0 {
				msg += fmt.Sprintf(" (evicting %d:%d)", old.PID, old.Page)
			}
			res.Log = append(res.Log, msg)
			victim = (victim + 1) % frames
		}
		res.Probes = append(res.Probes, probes)
		res.Tables = append(res.Tables, append([]InvertedEntry(nil), table...))
	}
	return res, nil
}
