package infrastructure

import "fmt"

// Rollout is a deployment strategy.
type Rollout string

// The strategies Deploy knows.
const (
	BlueGreen Rollout = "blue-green" // switch all traffic to the new fleet at once; switch back on trouble
	Canary    Rollout = "canary"     // 5%, 25%, 50%, 100%, watching errors at each stage
	Rolling   Rollout = "rolling"    // replace one instance per step
)

// DeployStep is one step of a Deploy run.
type DeployStep struct {
	NewShare int    `json:"newShare"` // percent of traffic on the new version
	Errors   int    `json:"errors"`   // failed requests this step
	Note     string `json:"note"`
}

// DeployResult is a Deploy run.
type DeployResult struct {
	Steps      []DeployStep `json:"steps"`
	Errors     int          `json:"errors"`
	RolledBack bool         `json:"rolledBack"`
}

// Deploy rolls a new version out to the instances, serving rps
// requests a step. If bad, the new version fails errorRate percent of its
// requests, and the deployment is rolled back once a step's errors on the
// new version pass 1% of that step's requests, detected one step late (it
// takes a step to see the metrics). Blue/green risks all traffic for that
// step; a canary risks only its small slice; rolling sits between.
//
// Time: O(steps). Space: O(steps).
func Deploy(strategy Rollout, instances, rps, errorRate int, bad bool) (DeployResult, error) {
	if instances < 1 || instances > 50 || rps < 1 || rps > 100000 || errorRate < 0 || errorRate > 100 {
		return DeployResult{}, inputErr("1-50 instances, 1-100000 requests a step, error rate 0-100%%")
	}
	var plan []int
	switch strategy {
	case BlueGreen:
		plan = []int{100, 100, 100}
	case Canary:
		plan = []int{5, 5, 25, 25, 50, 50, 100}
	case Rolling:
		for i := 1; i <= instances; i++ {
			plan = append(plan, 100*i/instances)
		}
		plan = append(plan, 100)
	default:
		return DeployResult{}, inputErr("unknown strategy %q", strategy)
	}
	var res DeployResult
	alarm := false
	for _, share := range plan {
		if alarm {
			res.RolledBack = true
			res.Steps = append(res.Steps, DeployStep{0, 0, "the error alarm fired: roll back, all traffic on the old version"})
			break
		}
		errs := 0
		if bad {
			errs = rps * share / 100 * errorRate / 100
		}
		res.Errors += errs
		note := fmt.Sprintf("%d%% of traffic on the new version", share)
		if errs*100 > rps {
			alarm = true
			note += fmt.Sprintf(": %d errors, over 1%% of requests", errs)
		}
		res.Steps = append(res.Steps, DeployStep{share, errs, note})
	}
	return res, nil
}
