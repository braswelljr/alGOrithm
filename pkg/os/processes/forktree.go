package processes

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Proc is one process in a ForkTree run.
type Proc struct {
	PID     int    `json:"pid"`
	Parent  int    `json:"parent"`
	Program string `json:"program"`
	// State is "running", "zombie" (exited, status not yet collected) or
	// "reaped" (gone from the process table).
	State  string `json:"state"`
	Status int    `json:"status"`
}

// ForkResult is the process table after a ForkTree script, and what
// happened.
type ForkResult struct {
	Procs []Proc   `json:"procs"`
	Log   []string `json:"log"`
}

// ForkTree runs a script of process-control calls, one a line:
//
//	PID fork          PID makes a child, a copy of itself, with the next free ID
//	PID exec PROGRAM  the process replaces its program but keeps its ID
//	PID exit STATUS   the process ends; it stays a zombie until its parent waits
//	PID wait          PID collects one exited child (or blocks if none has exited)
//
// PID 1 is init, which starts PID 2, a shell. When a process exits, its
// living children are orphans and are adopted by init, which waits for
// every child at once, so an orphan never stays a zombie. A blocked wait
// returns as soon as a child exits.
//
// Time: O(n * p) for n lines and p processes. Space: O(p).
func ForkTree(script string) (ForkResult, error) {
	procs := map[int]*Proc{1: {PID: 1, Program: "init", State: "running"}, 2: {PID: 2, Parent: 1, Program: "sh", State: "running"}}
	waiting := map[int]bool{}
	next := 3
	var log []string
	// reap collects one zombie child of parent, if any.
	reap := func(parent int) bool {
		for _, pid := range sortedPIDs(procs) {
			c := procs[pid]
			if c.Parent == parent && c.State == "zombie" {
				c.State = "reaped"
				log = append(log, fmt.Sprintf("  %d reaps %d (exit status %d); its process-table slot is freed", parent, pid, c.Status))
				return true
			}
		}
		return false
	}
	for n, line := range strings.Split(script, "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		pid, err := strconv.Atoi(f[0])
		p, ok := procs[pid]
		if err != nil || !ok || p.State != "running" {
			return ForkResult{}, inputErr(n+1, "%q is not a running process", f[0])
		}
		if len(f) < 2 {
			return ForkResult{}, inputErr(n+1, "missing a call: fork, exec, exit or wait")
		}
		switch f[1] {
		case "fork":
			procs[next] = &Proc{PID: next, Parent: pid, Program: p.Program, State: "running"}
			log = append(log, fmt.Sprintf("%d fork: child %d is a copy of %d (%s); fork returns %d to the parent and 0 to the child", pid, next, pid, p.Program, next))
			next++
		case "exec":
			if len(f) < 3 {
				return ForkResult{}, inputErr(n+1, "exec needs a program")
			}
			log = append(log, fmt.Sprintf("%d exec %s: same PID, new program; the old code and data are gone", pid, f[2]))
			p.Program = f[2]
		case "exit":
			status := 0
			if len(f) > 2 {
				if status, err = strconv.Atoi(f[2]); err != nil {
					return ForkResult{}, inputErr(n+1, "bad exit status %q", f[2])
				}
			}
			if pid == 1 {
				return ForkResult{}, inputErr(n+1, "init never exits")
			}
			p.State, p.Status = "zombie", status
			log = append(log, fmt.Sprintf("%d exit %d: now a zombie, holding its status until %d waits", pid, status, p.Parent))
			for _, c := range sortedPIDs(procs) {
				if procs[c].Parent == pid && procs[c].State != "reaped" {
					procs[c].Parent = 1
					log = append(log, fmt.Sprintf("  %d is an orphan: init (1) adopts it", c))
				}
			}
			if waiting[p.Parent] || p.Parent == 1 {
				waiting[p.Parent] = false
				reap(p.Parent)
			}
			for reap(1) { // init collects adopted zombies straight away
			}
		case "wait":
			if !reap(pid) {
				waiting[pid] = true
				log = append(log, fmt.Sprintf("%d wait: no child has exited yet, so %d blocks", pid, pid))
			}
		default:
			return ForkResult{}, inputErr(n+1, "unknown call %q (fork, exec, exit or wait)", f[1])
		}
	}
	var out []Proc
	for _, pid := range sortedPIDs(procs) {
		out = append(out, *procs[pid])
	}
	return ForkResult{Procs: out, Log: log}, nil
}

func sortedPIDs(m map[int]*Proc) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}
