// Command spell-checker suggests corrections for misspelled words by edit
// distance (Levenshtein): the fewest inserts, deletes and substitutions
// between two words. Spell checkers, "did you mean" prompts and fuzzy
// command-line completion rank candidates this way.
package main

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/text"
)

func main() {
	dictionary := []string{"algorithm", "allocate", "logarithm", "rhythm", "argument", "alignment", "binary", "library"}
	for _, typo := range []string{"algorthm", "librery", "argumant"} {
		s := text.Suggest(dictionary, typo, 2)
		if len(s) == 0 {
			fmt.Printf("%s: no suggestion\n", typo)
			continue
		}
		fmt.Printf("%s: did you mean %q? (edit distance %d)\n", typo, s[0].Word, s[0].Distance)
	}
}
