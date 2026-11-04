package processes

import (
	"fmt"
	"strconv"
	"strings"
)

// syscallCost is how many ticks the kernel spends on each call.
var syscallCost = map[string]int{"open": 3, "read": 4, "write": 3, "close": 1, "getpid": 1, "fork": 5}

// SyscallResult is a SystemCalls run: the timeline of user and kernel mode,
// the log, and the totals.
type SyscallResult struct {
	Timeline    Timeline `json:"timeline"`
	Log         []string `json:"log"`
	UserTicks   int      `json:"userTicks"`
	KernelTicks int      `json:"kernelTicks"`
	Switches    int      `json:"switches"` // mode switches, two per call
}

// SystemCalls runs a program, one step a line:
//
//	compute N           that many ticks of ordinary work in user mode
//	open PATH           returns the lowest free file descriptor (from 3)
//	read FD N           reads N bytes from an open descriptor
//	write FD TEXT       writes to an open descriptor (1 is standard output)
//	close FD            frees the descriptor
//	getpid | fork
//
// A system call traps into the kernel (a mode switch), the kernel does the
// work on the program's behalf, then returns to user mode. A call on a
// descriptor that is not open fails with EBADF, as on Unix. The point is the
// cost: every crossing is extra time, which is why programs buffer their I/O.
//
// Time: O(total ticks). Space: O(total ticks).
func SystemCalls(script string) (SyscallResult, error) {
	var res SyscallResult
	fds := map[int]string{0: "stdin", 1: "stdout", 2: "stderr"}
	tick := 0
	run := func(lane, what, state string, n int) {
		for i := 0; i < n; i++ {
			res.Timeline = append(res.Timeline, Event{Tick: tick, Lane: lane, What: what, State: state})
			tick++
		}
	}
	for n, line := range strings.Split(script, "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		if f[0] == "compute" {
			k, err := strconv.Atoi(arg(f, 1))
			if err != nil || k < 1 || k > 100 {
				return res, inputErr(n+1, "compute needs 1-100 ticks")
			}
			run("user", "compute", "running", k)
			res.UserTicks += k
			continue
		}
		cost, ok := syscallCost[f[0]]
		if !ok {
			return res, inputErr(n+1, "unknown step %q", f[0])
		}
		result := "ok"
		switch f[0] {
		case "open":
			fd := 3
			for fds[fd] != "" {
				fd++
			}
			fds[fd] = arg(f, 1)
			result = "fd " + strconv.Itoa(fd)
		case "read", "write", "close":
			fd, err := strconv.Atoi(arg(f, 1))
			if err != nil || fds[fd] == "" {
				result = "-1 EBADF (bad file descriptor)"
				cost = 1
				break
			}
			switch f[0] {
			case "close":
				delete(fds, fd)
			case "read":
				result = arg(f, 2) + " bytes from " + fds[fd]
			default:
				result = strconv.Itoa(len(strings.Join(f[2:], " "))) + " bytes to " + fds[fd]
			}
		case "getpid":
			result = "42"
		case "fork":
			result = "child 43"
		}
		run("user", "trap: "+f[0], "blocked", 1)
		run("kernel", f[0], "running", cost)
		res.KernelTicks += cost
		res.Switches += 2
		res.Log = append(res.Log, fmt.Sprintf("%s -> %s (trap to kernel, %d ticks, return to user)", strings.Join(f, " "), result, cost))
	}
	return res, nil
}

func arg(f []string, i int) string {
	if i < len(f) {
		return f[i]
	}
	return ""
}
