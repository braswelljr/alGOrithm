package memory

import "fmt"

// Partition is a block of memory after allocation: its size, and the process
// placed in it (-1 for none) with that process's size.
type Partition struct {
	Size    int `json:"size"`
	Process int `json:"process"`
	Used    int `json:"used"`
}

// Allocation is the result of placing processes in partitions.
type Allocation struct {
	// Placed[i] is the partition process i went in, or -1 when none fit.
	Placed []int `json:"placed"`
	// Steps[i] is every partition after process i was placed.
	Steps [][]Partition `json:"steps"`
	// Internal is the memory wasted inside used partitions; Unplaced the
	// processes that did not fit.
	Internal, Unplaced int
}

// FirstFit places each process, in order, in the first free partition big
// enough. Fast; leaves small leftovers at the start of memory.
//
// Time: O(n * m) for n processes and m partitions. Space: O(n * m).
func FirstFit(partitions, processes []int) (Allocation, error) {
	return allocate(partitions, processes, func(cands []int, sizes []int) int { return cands[0] })
}

// BestFit places each process in the smallest free partition big enough,
// wasting the least in that partition but leaving many tiny holes.
//
// Time: O(n * m). Space: O(n * m).
func BestFit(partitions, processes []int) (Allocation, error) {
	return allocate(partitions, processes, func(cands []int, sizes []int) int {
		best := cands[0]
		for _, c := range cands {
			if sizes[c] < sizes[best] {
				best = c
			}
		}
		return best
	})
}

// WorstFit places each process in the largest free partition, so what is left
// over stays big enough to be useful.
//
// Time: O(n * m). Space: O(n * m).
func WorstFit(partitions, processes []int) (Allocation, error) {
	return allocate(partitions, processes, func(cands []int, sizes []int) int {
		best := cands[0]
		for _, c := range cands {
			if sizes[c] > sizes[best] {
				best = c
			}
		}
		return best
	})
}

// allocate uses fixed partitions: each partition holds at most one process,
// and the rest of it is internal fragmentation.
func allocate(partitions, processes []int, choose func(cands, sizes []int) int) (Allocation, error) {
	if len(partitions) == 0 || len(processes) == 0 || len(partitions) > 1000 || len(processes) > 1000 {
		return Allocation{}, fmt.Errorf("%w: give 1 to 1000 partitions and processes", ErrInput)
	}
	mem := make([]Partition, len(partitions))
	for i, s := range partitions {
		if s < 1 {
			return Allocation{}, fmt.Errorf("%w: partition %d has size %d", ErrInput, i+1, s)
		}
		mem[i] = Partition{Size: s, Process: -1}
	}
	var a Allocation
	for p, need := range processes {
		var cands []int
		for i, part := range mem {
			if part.Process < 0 && part.Size >= need {
				cands = append(cands, i)
			}
		}
		at := -1
		if len(cands) > 0 {
			at = choose(cands, partitions)
			mem[at].Process, mem[at].Used = p, need
			a.Internal += partitions[at] - need
		} else {
			a.Unplaced++
		}
		a.Placed = append(a.Placed, at)
		a.Steps = append(a.Steps, append([]Partition(nil), mem...))
	}
	return a, nil
}
