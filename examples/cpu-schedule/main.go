// Command cpu-schedule runs four processes through round-robin scheduling and
// prints the Gantt chart and each process's waiting time.
package main

import (
	"fmt"
	"log"

	"github.com/braswelljr/alGOrithm/pkg/os/scheduling"
)

func main() {
	ps := []scheduling.Process{
		{Name: "P1", Arrival: 0, Burst: 5},
		{Name: "P2", Arrival: 1, Burst: 3},
		{Name: "P3", Arrival: 2, Burst: 8},
		{Name: "P4", Arrival: 3, Burst: 6},
	}
	s, err := scheduling.RoundRobin(ps, 3)
	if err != nil {
		log.Fatal(err)
	}
	for _, sl := range s.Timeline {
		fmt.Printf("[%2d-%2d) %s\n", sl.Start, sl.End, sl.Process)
	}
	for _, st := range s.Stats {
		fmt.Printf("%s waited %d\n", st.Name, st.Waiting)
	}
	fmt.Printf("average waiting: %.2f\n", s.AvgWaiting)
}
