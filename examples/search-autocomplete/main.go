// Command search-autocomplete suggests completions as a user types in a
// search box, the way editors, shells and site search bars do: the words go
// into a trie once, and each keystroke walks only the typed prefix.
package main

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/structures/lookup"
)

func main() {
	products := lookup.NewTrie(
		"laptop", "laptop stand", "lamp", "lantern", "keyboard", "keycap",
		"key holder", "monitor", "monitor arm", "mouse", "mousepad",
	)
	for _, typed := range []string{"la", "key", "mon", "z"} {
		fmt.Printf("suggestions for %q: %v\n", typed, products.WithPrefix(typed))
	}
}
