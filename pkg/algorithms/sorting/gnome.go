package sorting

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Gnome sorts s in ascending order with gnome sort: walk forward while
// neighbors are in order; when they are not, swap them and step back. It sorts
// in place and returns s. It is stable.
//
// Time: O(n^2), O(n) when s is already sorted.
// Space: O(1).
func Gnome[T cmp.Ordered](s []T) []T { return gnome(s, nop) }

// TraceGnome is Gnome, reporting each comparison and swap to r.
func TraceGnome[T cmp.Ordered](s []T, r trace.Recorder) []T { return gnome(s, rec{r}) }

func gnome[T cmp.Ordered](s []T, x rec) []T {
	for i := 1; i < len(s); {
		if i == 0 {
			i++
			continue
		}
		x.compare(i-1, i)
		if s[i-1] <= s[i] {
			i++
			continue
		}
		s[i-1], s[i] = s[i], s[i-1]
		x.swap(i-1, i)
		i--
	}
	return s
}
