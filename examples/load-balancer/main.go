// Command load-balancer sends each request to the backend with the fewest
// open connections (least connections, as in nginx, HAProxy and Envoy),
// keeping the backends in a min-heap ordered by connection count.
package main

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/structures/heaps"
)

type backend struct {
	name  string
	conns int
}

func main() {
	pool := heaps.NewPriorityHeap(func(a, b *backend) bool {
		return a.conns < b.conns || (a.conns == b.conns && a.name < b.name)
	}, &backend{"api-1", 0}, &backend{"api-2", 0}, &backend{"api-3", 0})
	// request cost in "connection-seconds": long uploads keep a connection open
	durations := []int{1, 5, 1, 1, 4, 1, 1, 1}
	fmt.Println("least-connections routing:")
	for i, d := range durations {
		b, _ := pool.Pop()
		b.conns += d
		fmt.Printf("  request %d (%ds) -> %s, now %d open\n", i+1, d, b.name, b.conns)
		pool.Push(b)
	}
}
