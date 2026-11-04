package processes

import (
	"fmt"
	"sort"
	"strings"
)

// defaultAction is what a signal does when the process has not changed it.
var defaultAction = map[string]string{
	"SIGINT": "terminate", "SIGTERM": "terminate", "SIGKILL": "terminate", "SIGHUP": "terminate",
	"SIGUSR1": "terminate", "SIGSEGV": "terminate (core dump)", "SIGCHLD": "ignore", "SIGSTOP": "stop", "SIGCONT": "continue",
}

// SignalResult is a Signals run: the log, whether the process is still
// alive, and the signals still pending (blocked) at the end.
type SignalResult struct {
	Log     []string `json:"log"`
	Alive   bool     `json:"alive"`
	Pending []string `json:"pending"`
}

// Signals runs a script against one process, one step a line:
//
//	handle SIG   install a handler (like signal(SIG, f))
//	ignore SIG   set the disposition to ignore
//	default SIG  restore the default action
//	block SIG    add to the signal mask: delivery waits
//	unblock SIG  remove from the mask: a pending one is delivered now
//	send SIG     another process (or the kernel) sends it
//
// SIGKILL and SIGSTOP cannot be handled, ignored or blocked: that is how a
// stuck process can always be stopped. A blocked signal is pending until it
// is unblocked, and two of the same kind pending count as one.
//
// Time: O(n) for n lines. Space: O(signals).
func Signals(script string) (SignalResult, error) {
	disp := map[string]string{}
	blocked, pending := map[string]bool{}, map[string]bool{}
	res := SignalResult{Alive: true}
	deliver := func(sig string) {
		action := defaultAction[sig]
		if d := disp[sig]; d != "" && sig != "SIGKILL" && sig != "SIGSTOP" {
			action = d
		}
		switch action {
		case "handler":
			res.Log = append(res.Log, "  delivered: the handler runs, then the program carries on")
		case "ignore":
			res.Log = append(res.Log, "  delivered: ignored")
		case "stop", "continue":
			res.Log = append(res.Log, "  delivered: the process is told to "+action)
		default:
			res.Log = append(res.Log, "  delivered: default action, "+action+"; the process ends")
			res.Alive = false
		}
	}
	for n, line := range strings.Split(script, "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		if len(f) != 2 {
			return res, inputErr(n+1, "want a step and a signal, like \"send SIGINT\"")
		}
		step, sig := f[0], strings.ToUpper(f[1])
		if !strings.HasPrefix(sig, "SIG") {
			sig = "SIG" + sig
		}
		if _, ok := defaultAction[sig]; !ok {
			return res, inputErr(n+1, "unknown signal %s", sig)
		}
		if !res.Alive {
			return res, inputErr(n+1, "the process has already ended")
		}
		uncatchable := sig == "SIGKILL" || sig == "SIGSTOP"
		switch step {
		case "handle", "ignore", "default":
			if uncatchable {
				res.Log = append(res.Log, fmt.Sprintf("%s %s: refused (EINVAL), %s cannot be caught or ignored", step, sig, sig))
				continue
			}
			disp[sig] = map[string]string{"handle": "handler", "ignore": "ignore", "default": ""}[step]
			res.Log = append(res.Log, fmt.Sprintf("%s %s: disposition set", step, sig))
		case "block":
			if uncatchable {
				res.Log = append(res.Log, fmt.Sprintf("block %s: silently not blocked, %s cannot be blocked", sig, sig))
				continue
			}
			blocked[sig] = true
			res.Log = append(res.Log, fmt.Sprintf("block %s: added to the signal mask", sig))
		case "unblock":
			delete(blocked, sig)
			res.Log = append(res.Log, fmt.Sprintf("unblock %s: removed from the mask", sig))
			if pending[sig] {
				delete(pending, sig)
				res.Log = append(res.Log, "  the pending "+sig+" is delivered now")
				deliver(sig)
			}
		case "send":
			res.Log = append(res.Log, "send "+sig)
			if blocked[sig] {
				if pending[sig] {
					res.Log = append(res.Log, "  already pending: standard signals do not queue, so this one merges with it")
				} else {
					res.Log = append(res.Log, "  blocked: it stays pending")
				}
				pending[sig] = true
				continue
			}
			deliver(sig)
		default:
			return res, inputErr(n+1, "unknown step %q", step)
		}
	}
	for s := range pending {
		res.Pending = append(res.Pending, s)
	}
	sort.Strings(res.Pending)
	return res, nil
}
