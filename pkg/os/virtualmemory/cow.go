package virtualmemory

import (
	"fmt"
	"strconv"
	"strings"
)

// COWResult is a CopyOnWrite run: which frame each process's pages are in
// after each step, the pages copied, and what an eager fork would copy.
type COWResult struct {
	Parent [][]int  `json:"parent"`
	Child  [][]int  `json:"child"`
	Copies int      `json:"copies"`
	Eager  int      `json:"eager"`
	Log    []string `json:"log"`
}

// CopyOnWrite forks a process with the given number of pages and runs a script, one
// access a line, "parent write N", "child read N" and so on. After fork the
// child's page table points at the parent's frames, all marked read-only.
// Reads share; the first write to a shared page faults, and the kernel
// copies just that page into a new frame for the writer. A child that
// execs straight away (the usual case) copies nothing, where an eager fork
// would have copied every page.
//
// Time: O(n + pages). Space: O(n * pages).
func CopyOnWrite(pages int, script string) (COWResult, error) {
	if pages < 1 || pages > 32 {
		return COWResult{}, inputErr("1-32 pages")
	}
	parent, child := make([]int, pages), make([]int, pages)
	refs := map[int]int{}
	for i := range parent {
		parent[i], child[i] = i, i
		refs[i] = 2
	}
	next := pages
	res := COWResult{Eager: pages}
	res.Log = append(res.Log, fmt.Sprintf("fork: the child shares all %d frames, every page now read-only", pages))
	for n, line := range strings.Split(script, "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		if len(f) != 3 || (f[0] != "parent" && f[0] != "child") || (f[1] != "read" && f[1] != "write") {
			return res, inputErr("line %d: want \"parent write 2\" or \"child read 0\"", n+1)
		}
		pg, err := strconv.Atoi(f[2])
		if err != nil || pg < 0 || pg >= pages {
			return res, inputErr("line %d: page 0-%d", n+1, pages-1)
		}
		table := parent
		if f[0] == "child" {
			table = child
		}
		fr := table[pg]
		switch {
		case f[1] == "read":
			res.Log = append(res.Log, fmt.Sprintf("%s reads page %d: frame %d, no fault", f[0], pg, fr))
		case refs[fr] > 1:
			refs[fr]--
			table[pg] = next
			refs[next] = 1
			res.Copies++
			res.Log = append(res.Log, fmt.Sprintf("%s writes page %d: protection fault on shared frame %d; copy it to frame %d, now writable", f[0], pg, fr, next))
			next++
		default:
			res.Log = append(res.Log, fmt.Sprintf("%s writes page %d: frame %d is its own already, no copy", f[0], pg, fr))
		}
		res.Parent = append(res.Parent, append([]int(nil), parent...))
		res.Child = append(res.Child, append([]int(nil), child...))
	}
	res.Log = append(res.Log, fmt.Sprintf("copied %d pages, where an eager fork copies %d", res.Copies, res.Eager))
	return res, nil
}
