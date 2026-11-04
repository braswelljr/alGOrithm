// Command leader-election runs a five-node Raft cluster that elects a
// leader and replicates log entries, as etcd (behind Kubernetes), Consul and
// CockroachDB do to agree on one copy of their data.
package main

import (
	"fmt"
	"log"

	"github.com/braswelljr/alGOrithm/pkg/systems/distributed"
)

func main() {
	res, err := distributed.Raft(distributed.RaftConfig{Nodes: 5, Ticks: 300, Seed: 3, Every: 20})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d election(s) in a 5-node cluster\n", len(res.Elections))
	for _, e := range res.Elections {
		fmt.Printf("  tick %3d: node %d leads term %d\n", e.Tick, e.Leader, e.Term)
	}
	fmt.Printf("committed entries: %d, messages sent: %d\n", res.Committed, res.Messages)
}
