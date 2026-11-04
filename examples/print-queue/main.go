// Command print-queue shares a printer between jobs with round-robin time
// slices, the way an OS shares a CPU: no job waits forever behind a long one.
package main

import (
	"fmt"
	"log"

	"github.com/braswelljr/alGOrithm/pkg/os/scheduling"
)

func main() {
	jobs := []scheduling.Process{
		{Name: "report.pdf", Arrival: 0, Burst: 6},
		{Name: "photo.png", Arrival: 1, Burst: 2},
		{Name: "thesis.pdf", Arrival: 2, Burst: 9},
		{Name: "label.txt", Arrival: 3, Burst: 1},
	}
	s, err := scheduling.RoundRobin(jobs, 2) // 2 pages per turn
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("printer timeline (pages):")
	for _, sl := range s.Timeline {
		fmt.Printf("  %2d-%2d %s\n", sl.Start, sl.End, sl.Process)
	}
	fmt.Printf("average wait: %.2f pages\n", s.AvgWaiting)
}
