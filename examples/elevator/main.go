// Command elevator serves floor requests with the SCAN (elevator) algorithm:
// keep moving in one direction, stopping at requested floors, then reverse.
// Real elevators and disk heads both move this way to avoid zigzagging.
package main

import (
	"fmt"
	"log"

	"github.com/braswelljr/alGOrithm/pkg/os/disk"
)

func main() {
	requests := []int{7, 2, 14, 5, 11, 1}
	const at, floors = 6, 16
	scan, err := disk.SCAN(requests, at, floors, false)
	if err != nil {
		log.Fatal(err)
	}
	naive, _ := disk.FCFS(requests, at, floors)
	fmt.Printf("SCAN stops: %v, %d floors traveled\n", scan.Order, scan.Movement)
	fmt.Printf("first-come stops: %v, %d floors traveled\n", naive.Order, naive.Movement)
}
