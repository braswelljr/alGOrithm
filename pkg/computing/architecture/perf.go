package architecture

import (
	"fmt"
	"strconv"
	"strings"
)

// Class is one kind of instruction in a mix: its share of the instructions
// run, in percent, and the cycles each takes.
type Class struct {
	Name    string `json:"name"`
	Percent int    `json:"percent"`
	Cycles  int    `json:"cycles"`
}

// ParseMix reads an instruction mix, one class a line: "name percent
// cycles" ("load 20 5"). The percents must add up to 100.
func ParseMix(text string) ([]Class, error) {
	var out []Class
	total := 0
	for n, raw := range strings.Split(text, "\n") {
		f := strings.Fields(raw)
		if len(f) == 0 {
			continue
		}
		if len(f) != 3 {
			return nil, fmt.Errorf("%w: line %d: want name percent cycles", ErrInput, n+1)
		}
		p, err1 := strconv.Atoi(strings.TrimSuffix(f[1], "%"))
		c, err2 := strconv.Atoi(f[2])
		if err1 != nil || err2 != nil || p < 0 || c < 1 {
			return nil, fmt.Errorf("%w: line %d: the percent and the cycles are whole numbers, cycles at least 1", ErrInput, n+1)
		}
		total += p
		out = append(out, Class{f[0], p, c})
	}
	if len(out) == 0 || total != 100 {
		return nil, fmt.Errorf("%w: the percents add up to %d, not 100", ErrInput, total)
	}
	return out, nil
}

// CPI is the average cycles per instruction of a mix: each class's cycles
// weighted by its share. With the clock rate, CPU time = instructions x
// CPI / clock rate, the "iron law" of processor performance.
//
// Time: O(c) for c classes.
func CPI(mix []Class) float64 {
	sum := 0
	for _, c := range mix {
		sum += c.Percent * c.Cycles
	}
	return float64(sum) / 100
}

// Amdahl is the speedup of a program whose parallel part (in percent) runs
// on n processors while the rest stays serial (Gene Amdahl, 1967):
// 1 / ((1 - p) + p / n). However many processors, the speedup stays below
// 1 / (1 - p): the serial part sets the limit.
//
// Time: O(1).
func Amdahl(parallelPercent, n int) float64 {
	p := float64(parallelPercent) / 100
	return 1 / ((1 - p) + p/float64(n))
}

// Level is one level of the memory hierarchy: its name, typical size and
// latency in nanoseconds.
type Level struct {
	Name    string  `json:"name"`
	Size    string  `json:"size"`
	Latency float64 `json:"latency"`
}

// Hierarchy is the memory hierarchy of a typical desktop computer, fastest
// and smallest first. The numbers are orders of magnitude, not one
// machine's.
func Hierarchy() []Level {
	return []Level{
		{"register", "about 1 KB", 0.3},
		{"L1 cache", "32 to 64 KB", 1},
		{"L2 cache", "256 KB to 2 MB", 4},
		{"L3 cache", "8 to 64 MB", 15},
		{"main memory (DRAM)", "8 to 64 GB", 80},
		{"SSD", "0.5 to 4 TB", 100_000},
		{"hard disk", "1 to 20 TB", 10_000_000},
	}
}

// HumanScale says how long a latency would feel if one CPU cycle (0.3 ns)
// took one second.
func HumanScale(ns float64) string {
	s := ns / 0.3
	switch {
	case s < 60:
		return fmt.Sprintf("%.0f seconds", s)
	case s < 3600:
		return fmt.Sprintf("%.0f minutes", s/60)
	case s < 86400:
		return fmt.Sprintf("%.0f hours", s/3600)
	case s < 86400*365:
		return fmt.Sprintf("%.0f days", s/86400)
	}
	return fmt.Sprintf("%.1f years", s/(86400*365))
}

// Buses compares the same run on two designs. A von Neumann machine keeps
// instructions and data in one memory behind one bus, so a load or store
// and the next fetch cannot happen in the same cycle; a Harvard machine has
// two memories and two buses, so they can. The model counts one bus cycle
// per memory access and nothing else: it shows the bottleneck, not a real
// machine's timing.
//
// Time: O(s) for s instructions run.
func Buses(p Program, maxInstrs int) (vonNeumann, harvard int, err error) {
	run, err := Simulate(p, 64, maxInstrs)
	if err != nil {
		return 0, 0, err
	}
	data := 0
	for _, ex := range run.Trace {
		if op := p.Code[ex.PC].Op; op == "LD" || op == "ST" || op == "CALL" || op == "RET" {
			data++
		}
	}
	return len(run.Trace) + data, len(run.Trace), nil
}
