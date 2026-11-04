package ipc

import (
	"fmt"
	"strconv"
	"strings"
)

// Pipe runs a writer and a reader on the two ends of a pipe holding
// capacity bytes. The script gives each side's calls in order, one a line:
// "writer: TEXT", "writer: close", "reader: N" (read up to N bytes) and
// "reader: sleep" (the reader is busy for a tick).
// They take turns, one call each a tick. A write into a full pipe writes
// what fits and blocks until the reader makes room; a read from an empty
// pipe blocks until data arrives, and returns 0 (end of file) once the
// writer has closed and the pipe is empty. This is the shell's `a | b`.
//
// Time: O(total bytes + calls). Space: O(capacity).
func Pipe(script string, capacity int) (Result, error) {
	if capacity < 1 || capacity > 64 {
		return Result{}, inputErr("capacity 1-64 bytes")
	}
	var writes []string
	var reads []int
	for n, line := range strings.Split(script, "\n") {
		side, rest, ok := strings.Cut(strings.TrimSpace(line), ":")
		if strings.TrimSpace(line) == "" {
			continue
		}
		rest = strings.TrimSpace(rest)
		switch {
		case ok && side == "writer":
			writes = append(writes, rest)
		case ok && side == "reader" && rest == "sleep":
			reads = append(reads, 0)
		case ok && side == "reader":
			k, err := strconv.Atoi(rest)
			if err != nil || k < 1 || k > 64 {
				return Result{}, inputErr("line %d: a read takes 1-64 bytes", n+1)
			}
			reads = append(reads, k)
		default:
			return Result{}, inputErr("line %d: want \"writer: TEXT\", \"writer: close\", \"reader: N\" or \"reader: sleep\"", n+1)
		}
	}
	res := Result{Counts: map[string]int{}}
	var buf []byte
	closed := false
	pending := "" // what is left of the current write
	wi, ri := 0, 0
	for tick := 0; tick < 1000 && (wi < len(writes) || pending != "" || ri < len(reads)); tick++ {
		// writer
		switch {
		case pending == "" && wi < len(writes) && writes[wi] == "close":
			closed = true
			wi++
			res.Log = append(res.Log, fmt.Sprintf("tick %d: writer closes its end", tick))
			res.Timeline = append(res.Timeline, Event{tick, "writer", "close", "done"})
		case pending == "" && wi < len(writes):
			pending = writes[wi]
			wi++
			fallthrough
		case pending != "":
			room := capacity - len(buf)
			if room == 0 {
				res.Counts["writer blocked"]++
				res.Timeline = append(res.Timeline, Event{tick, "writer", "blocked: pipe full", "blocked"})
				break
			}
			k := min(room, len(pending))
			buf = append(buf, pending[:k]...)
			pending = pending[k:]
			res.Log = append(res.Log, fmt.Sprintf("tick %d: writer puts %q (%d/%d bytes in the pipe)", tick, buf[len(buf)-k:], len(buf), capacity))
			res.Timeline = append(res.Timeline, Event{tick, "writer", fmt.Sprintf("write %d bytes", k), "running"})
		}
		// reader
		if ri < len(reads) {
			switch {
			case reads[ri] == 0:
				ri++
				res.Timeline = append(res.Timeline, Event{tick, "reader", "busy elsewhere", "idle"})
			case len(buf) > 0:
				k := min(reads[ri], len(buf))
				res.Log = append(res.Log, fmt.Sprintf("tick %d: reader gets %q", tick, buf[:k]))
				buf = buf[k:]
				ri++
				res.Counts["bytes read"] += k
				res.Timeline = append(res.Timeline, Event{tick, "reader", fmt.Sprintf("read %d bytes", k), "running"})
			case closed && pending == "" && wi == len(writes):
				res.Log = append(res.Log, fmt.Sprintf("tick %d: reader gets 0 bytes: end of file", tick))
				ri++
				res.Timeline = append(res.Timeline, Event{tick, "reader", "end of file", "done"})
			default:
				res.Counts["reader blocked"]++
				res.Timeline = append(res.Timeline, Event{tick, "reader", "blocked: pipe empty", "blocked"})
			}
		}
		if ri < len(reads) && len(buf) == 0 && !closed && pending == "" && wi == len(writes) {
			res.Log = append(res.Log, "the reader would block forever: the writer never closed its end")
			break
		}
	}
	return res, nil
}
