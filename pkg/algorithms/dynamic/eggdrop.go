package dynamic

import "github.com/braswelljr/alGOrithm/pkg/trace"

// EggDrop returns the fewest drops that always find the highest safe floor of
// a building with the given number of floors and eggs, when an egg breaks from any
// floor above it and survives from any floor at or below. It turns the
// question around: table[t][e] is the most floors t drops with e eggs can
// cover, table[t][e] = table[t-1][e-1] + table[t-1][e] + 1 (the floor dropped
// from, the floors below if it breaks, the floors above if not); the answer is
// the first t reaching floors. Rows are drops 0..t, columns eggs 0..eggs.
//
// Time: O(eggs * answer). Space: O(eggs * answer).
func EggDrop(eggs, floors int) (int, [][]int, error) { return eggDrop(eggs, floors, trace.Nop{}) }

// TraceEggDrop is EggDrop, reporting each cell written (Set), the cells read
// for it (Compare), and the row that reached floors (Mark "path").
func TraceEggDrop(eggs, floors int, r trace.Recorder) (int, [][]int, error) {
	return eggDrop(eggs, floors, r)
}

func eggDrop(eggs, floors int, r trace.Recorder) (int, [][]int, error) {
	if eggs < 1 || eggs > 64 || floors < 0 || floors > 1_000_000 {
		return 0, nil, ErrInput
	}
	tr := &table{cells: [][]int{make([]int, eggs+1)}, cols: eggs + 1, r: r}
	for e := 0; e <= eggs; e++ {
		tr.set(0, e, 0)
	}
	t := 0
	for tr.cells[t][eggs] < floors {
		t++
		tr.cells = append(tr.cells, make([]int, eggs+1))
		tr.set(t, 0, 0)
		for e := 1; e <= eggs; e++ {
			tr.read([2]int{t - 1, e - 1}, [2]int{t - 1, e})
			tr.set(t, e, min(tr.cells[t-1][e-1]+tr.cells[t-1][e]+1, floors))
		}
	}
	tr.path([][2]int{{t, eggs}})
	return t, tr.cells, nil
}
