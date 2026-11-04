// Command office-subnets splits an office network block into equal subnets
// for departments with CIDR arithmetic, as network engineers and cloud VPC
// planners do, and checks which subnet an address belongs to.
package main

import (
	"fmt"
	"log"

	"github.com/braswelljr/alGOrithm/pkg/systems/networking"
)

func main() {
	block, err := networking.ParsePrefix("10.20.0.0/22")
	if err != nil {
		log.Fatal(err)
	}
	subnets, err := networking.Split(block, 4)
	if err != nil {
		log.Fatal(err)
	}
	depts := []string{"engineering", "sales", "guest wifi", "servers"}
	for i, s := range subnets {
		fmt.Printf("%-12s %v/%d (broadcast %v)\n", depts[i], s.Addr, s.Bits, s.Broadcast())
	}
	ip, _ := networking.ParseAddr("10.20.2.77")
	for i, s := range subnets {
		if s.Contains(ip) {
			fmt.Printf("10.20.2.77 is in %s\n", depts[i])
		}
	}
}
