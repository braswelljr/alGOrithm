// Command inventory-index looks up products by SKU with a B+ tree, the index
// behind almost every relational database (PostgreSQL, MySQL's InnoDB,
// SQLite): point lookups touch one root-to-leaf path, and range scans walk
// the linked leaves in order.
package main

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/systems/databases"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func main() {
	index := databases.NewBPlusTree(4)
	stock := map[int]int{1001: 12, 1007: 0, 1013: 40, 1020: 7, 1042: 3, 1050: 19, 1066: 25, 1070: 1}
	for sku, qty := range stock {
		index.Insert(sku, qty)
	}
	fmt.Printf("indexed %d SKUs in a B+ tree of height %d\n", index.Len(), index.Height())
	if qty, ok := index.Search(1042, trace.Nop{}); ok {
		fmt.Printf("SKU 1042: %d in stock\n", qty)
	}
	fmt.Printf("SKUs 1010 to 1050: %v\n", index.Range(1010, 1050, trace.Nop{}))
}
