// Command deadlock-check decides whether granting connections from a shared
// pool could deadlock, with the Banker's algorithm: a state is safe if some
// order lets every client finish with what is left.
package main

import (
	"fmt"
	"log"

	"github.com/braswelljr/alGOrithm/pkg/os/deadlock"
)

func main() {
	// resources: database connections, worker threads, file handles
	available := []int{3, 3, 2}
	maximum := [][]int{{7, 5, 3}, {3, 2, 2}, {9, 0, 2}, {2, 2, 2}, {4, 3, 3}}
	allocation := [][]int{{0, 1, 0}, {2, 0, 0}, {3, 0, 2}, {2, 1, 1}, {0, 0, 2}}
	ok, order, _, err := deadlock.Safe(available, maximum, allocation)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("safe: %v, clients can finish in order %v\n", ok, order)
	grant, err := deadlock.Request(1, []int{1, 0, 2}, available, maximum, allocation)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("grant client 1 (1,0,2) now: %v\n", grant)
}
