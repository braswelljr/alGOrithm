package ipc

import (
	"fmt"
	"math/rand"
)

// SharedMemory has a writer process fill a shared four-word record with
// its round number, word by word, rounds times, while a reader process
// copies the record out whenever the seeded scheduler runs it. Without
// synchronization the reader can copy a record the writer is half way
// through ([3 3 2 2]: a torn read, Counts["torn"]). With a semaphore around
// both the write and the copy, every read sees one whole round. Shared
// memory is the fastest IPC (no copying through the kernel) precisely
// because the kernel does nothing to keep it consistent.
//
// Time: O(rounds * 4). Space: O(rounds).
func SharedMemory(rounds int, synchronized bool, seed int64) (Result, error) {
	if rounds < 1 || rounds > 30 {
		return Result{}, inputErr("1-30 rounds")
	}
	rng := rand.New(rand.NewSource(seed))
	var record [4]int
	res := Result{Counts: map[string]int{"torn": 0, "whole": 0}}
	locked := "" // who holds the semaphore
	round, word := 1, 0
	readerWord, reading := 0, [4]int{}
	tick := 0
	for ; tick < 2000 && round <= rounds; tick++ {
		if rng.Intn(2) == 0 { // the writer runs
			if synchronized && locked == "reader" {
				res.Timeline = append(res.Timeline, Event{tick, "writer", "waits on the semaphore", "blocked"})
				continue
			}
			if synchronized {
				locked = "writer"
			}
			record[word] = round
			res.Timeline = append(res.Timeline, Event{tick, "writer", fmt.Sprintf("writes word %d = %d", word, round), "running"})
			word++
			if word == 4 {
				word = 0
				round++
				if synchronized {
					locked = ""
				}
			}
			continue
		}
		if synchronized && locked == "writer" {
			res.Timeline = append(res.Timeline, Event{tick, "reader", "waits on the semaphore", "blocked"})
			continue
		}
		if synchronized {
			locked = "reader"
		}
		reading[readerWord] = record[readerWord]
		res.Timeline = append(res.Timeline, Event{tick, "reader", fmt.Sprintf("copies word %d = %d", readerWord, record[readerWord]), "running"})
		readerWord++
		if readerWord == 4 {
			readerWord = 0
			if synchronized {
				locked = ""
			}
			whole := reading[0] == reading[1] && reading[1] == reading[2] && reading[2] == reading[3]
			if whole {
				res.Counts["whole"]++
			} else {
				res.Counts["torn"]++
				res.Log = append(res.Log, fmt.Sprintf("tick %d: torn read %v: part from one round, part from another", tick, reading))
			}
		}
	}
	res.Log = append(res.Log, fmt.Sprintf("%d whole reads, %d torn", res.Counts["whole"], res.Counts["torn"]))
	return res, nil
}
