package infrastructure

import "fmt"

// ScaleStep is one tick of an Autoscale run.
type ScaleStep struct {
	Load      int `json:"load"`
	Instances int `json:"instances"` // serving
	Booting   int `json:"booting"`
	Dropped   int `json:"dropped"` // requests over capacity
}

// AutoscaleResult is an Autoscale run.
type AutoscaleResult struct {
	Steps   []ScaleStep `json:"steps"`
	Dropped int         `json:"dropped"`
	Cost    int         `json:"cost"` // instance-ticks paid for, booting included
	Log     []string    `json:"log"`
}

// Autoscale follows a load curve (requests per tick) with instances that
// each serve capacity requests a tick. Every tick it computes the instances
// needed to keep utilization at target percent; it adds them (they take
// boot ticks to start serving) and removes surplus ones, but waits cooldown
// ticks after any change before scaling down again, so a brief dip does not
// throw away warm capacity. Requests over the serving capacity are dropped:
// a sudden spike always outruns the boot time, which is why services keep
// headroom.
//
// Time: O(len(load)). Space: O(len(load)).
func Autoscale(load []int, capacity, target, boot, cooldown, minimum, maximum int) (AutoscaleResult, error) {
	if len(load) == 0 || len(load) > 500 || capacity < 1 || capacity > 10000 || target < 10 || target > 100 || boot < 0 || boot > 50 || cooldown < 0 || cooldown > 100 || minimum < 1 || maximum < minimum || maximum > 1000 {
		return AutoscaleResult{}, inputErr("1-500 load points, capacity 1-10000, target 10-100%%, boot 0-50, cooldown 0-100, 1 <= min <= max <= 1000")
	}
	serving, lastChange := minimum, -cooldown
	var booting []int // ticks at which booting instances start serving
	var res AutoscaleResult
	for t, l := range load {
		if l < 0 {
			return res, inputErr("load is never negative")
		}
		ready := booting[:0]
		for _, at := range booting {
			if at <= t {
				serving++
			} else {
				ready = append(ready, at)
			}
		}
		booting = ready
		need := (l*100 + capacity*target - 1) / (capacity * target)
		need = min(max(need, minimum), maximum)
		have := serving + len(booting)
		switch {
		case need > have:
			for i := 0; i < need-have; i++ {
				booting = append(booting, t+boot)
			}
			res.Log = append(res.Log, fmt.Sprintf("tick %d: load %d needs %d instances, have %d: start %d", t, l, need, have, need-have))
			lastChange = t
		case need < serving && t-lastChange >= cooldown:
			res.Log = append(res.Log, fmt.Sprintf("tick %d: load %d needs %d, scale in by %d", t, l, need, serving-need))
			serving = need
			lastChange = t
		}
		if boot == 0 {
			serving += len(booting)
			booting = nil
		}
		dropped := max(0, l-serving*capacity)
		res.Dropped += dropped
		res.Cost += serving + len(booting)
		res.Steps = append(res.Steps, ScaleStep{l, serving, len(booting), dropped})
	}
	return res, nil
}
