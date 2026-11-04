// Command large-file-sort sorts more numbers than fit in memory with an
// external merge sort: sort memory-sized runs, then merge k runs at a time.
// Databases sort big tables and Unix sort handles huge files this way.
package main

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/systems/databases"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func main() {
	values := make([]int, 40)
	x := 17
	for i := range values {
		x = (x*73 + 41) % 997
		values[i] = x
	}
	sorted, passes := databases.ExternalSort(values, 8, 3, trace.Nop{}) // 8 values fit in memory, 3-way merge
	for i, p := range passes {
		fmt.Printf("pass %d: sorted runs left: %d\n", i, len(p.Runs))
	}
	fmt.Printf("smallest 5: %v, largest: %d\n", sorted[:5], sorted[len(sorted)-1])
}
