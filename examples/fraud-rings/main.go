// Command fraud-rings groups accounts that share a device, card or address
// with union-find: any shared attribute links two accounts, and each group
// is a possible fraud ring. Payment companies and social networks find
// connected accounts this way; Kruskal's algorithm uses the same structure.
package main

import (
	"fmt"
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/structures/lookup"
)

func main() {
	accounts := []string{"acct-1", "acct-2", "acct-3", "acct-4", "acct-5", "acct-6", "acct-7"}
	shared := [][2]string{ // pairs that used the same device, card or address
		{"acct-1", "acct-2"}, {"acct-2", "acct-5"}, {"acct-3", "acct-4"}, {"acct-6", "acct-5"},
	}
	uf := lookup.NewUnionFind(accounts...)
	for _, p := range shared {
		uf.Union(p[0], p[1])
	}
	rings := map[string][]string{}
	for _, a := range accounts {
		root, _ := uf.Find(a)
		rings[root] = append(rings[root], a)
	}
	var groups [][]string
	for _, g := range rings {
		groups = append(groups, g)
	}
	sort.Slice(groups, func(i, j int) bool { return len(groups[i]) > len(groups[j]) })
	fmt.Printf("%d groups of linked accounts:\n", uf.Groups())
	for _, g := range groups {
		flag := ""
		if len(g) >= 3 {
			flag = "  <- review: possible ring"
		}
		fmt.Printf("  %v%s\n", g, flag)
	}
}
