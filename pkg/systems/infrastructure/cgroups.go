package infrastructure

import (
	"fmt"
	"strconv"
	"strings"
)

// Container is one container in a Cgroups run: its CPU shares and memory
// limit, what it wants, and what it got.
type Container struct {
	Name     string  `json:"name"`
	Shares   int     `json:"shares"`
	Want     float64 `json:"want"`     // CPUs it would use if alone
	Got      float64 `json:"got"`      // CPUs it gets
	MemLimit int     `json:"memLimit"` // MiB
	MemUse   int     `json:"memUse"`   // MiB it tries to use
	Killed   bool    `json:"killed"`   // by the OOM killer
}

// Cgroups shares cpus CPUs and enforces memory limits among containers given
// one a line, "NAME SHARES CPUS_WANTED MEM_LIMIT_MIB MEM_USE_MIB". A
// container over its memory limit is killed by the OOM killer, as the
// kernel does inside a cgroup, and uses no CPU. The CPUs are divided in
// proportion to shares among the containers that still want more (water
// filling): a container that wants less than its share gives the rest to
// the others, and a container's shares only matter under contention.
//
// Time: O(n^2) for n containers. Space: O(n).
func Cgroups(cpus int, spec string) ([]Container, error) {
	if cpus < 1 || cpus > 64 {
		return nil, inputErr("1-64 CPUs")
	}
	var cs []Container
	for n, line := range strings.Split(spec, "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		if len(f) != 5 {
			return nil, inputErr("line %d: want \"NAME SHARES CPUS LIMIT USE\"", n+1)
		}
		sh, e1 := strconv.Atoi(f[1])
		want, e2 := strconv.ParseFloat(f[2], 64)
		lim, e3 := strconv.Atoi(f[3])
		use, e4 := strconv.Atoi(f[4])
		if e1 != nil || e2 != nil || e3 != nil || e4 != nil || sh < 2 || sh > 262144 || want < 0 || want > 64 || lim < 1 || use < 0 {
			return nil, inputErr("line %d: shares 2-262144, CPUs 0-64, limit at least 1 MiB", n+1)
		}
		cs = append(cs, Container{Name: f[0], Shares: sh, Want: want, MemLimit: lim, MemUse: use, Killed: use > lim})
	}
	if len(cs) == 0 || len(cs) > 20 {
		return nil, inputErr("1-20 containers")
	}
	left := float64(cpus)
	active := map[int]bool{}
	for i, c := range cs {
		if !c.Killed && c.Want > 0 {
			active[i] = true
		}
	}
	for len(active) > 0 && left > 1e-9 {
		total := 0
		for i := range active {
			total += cs[i].Shares
		}
		satisfied := false
		for i := range active {
			share := left * float64(cs[i].Shares) / float64(total)
			if cs[i].Want-cs[i].Got <= share {
				satisfied = true
			}
		}
		if !satisfied { // everyone wants more than its share: split what is left
			for i := range active {
				cs[i].Got += left * float64(cs[i].Shares) / float64(total)
			}
			break
		}
		spent := 0.0
		for i := range active {
			share := left * float64(cs[i].Shares) / float64(total)
			if need := cs[i].Want - cs[i].Got; need <= share {
				cs[i].Got += need
				spent += need
				delete(active, i)
			}
		}
		left -= spent
	}
	return cs, nil
}

// String describes a container.
func (c Container) String() string {
	if c.Killed {
		return fmt.Sprintf("%s: OOM-killed (used %d MiB, limit %d)", c.Name, c.MemUse, c.MemLimit)
	}
	return fmt.Sprintf("%s: %.2f of %.2f CPUs wanted", c.Name, c.Got, c.Want)
}
