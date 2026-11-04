package design

import "fmt"

// SagaSteps are the steps of an order saga and what undoes each.
var SagaSteps = []struct{ Do, Undo string }{
	{"create order", "cancel order"},
	{"reserve inventory", "release inventory"},
	{"charge payment", "refund payment"},
	{"schedule shipping", "cancel shipment"},
	{"notify customer", ""},
}

// SagaConfig is an order saga and where it fails.
type SagaConfig struct {
	FailAt        int  // the step (1-based) that fails, 0 for none
	Orchestration bool // an orchestrator calls each service; otherwise services react to each other's events
}

// Saga runs an order through a saga: a business transaction across
// services, each with its own database, done as a sequence of local
// transactions instead of one distributed one. When step FailAt fails,
// every step already done is undone by its compensating action, in reverse
// order (refund the payment, then release the inventory, then cancel the
// order). With orchestration a central orchestrator calls each service and
// knows the state; with choreography each service listens for the previous
// one's event, which couples less but scatters the flow. It returns the
// actions in the order they ran. Compensations must themselves be
// retryable; the model leaves out their failures.
//
// Time: O(steps). Space: O(steps).
func Saga(c SagaConfig) ([]string, Report, error) {
	if err := check(c.FailAt >= 0 && c.FailAt <= len(SagaSteps), "fail at step 0 to %d", len(SagaSteps)); err != nil {
		return nil, Report{}, err
	}
	comps := []Component{{ID: "client", Kind: Client, Label: "Checkout"}}
	links := []string{}
	services := []string{"orders", "inventory", "payments", "shipping", "notify"}
	if c.Orchestration {
		comps = append(comps, Component{ID: "orch", Kind: Service, Label: "Saga orchestrator", Notes: "knows every step's state"})
		links = append(links, "client>orch")
	} else {
		comps = append(comps, Component{ID: "bus", Kind: Queue, Label: "Event bus"})
		links = append(links, "client>orders")
	}
	for _, s := range services {
		comps = append(comps, Component{ID: s, Kind: Service, Label: s + " service"})
		if c.Orchestration {
			links = append(links, "orch>"+s)
		} else {
			links = append(links, s+">bus", "bus>"+s)
		}
	}
	a := arch(comps, links...)
	var done, actions []string
	call := func(i int, status Status, note string) {
		f := a.frame(note)
		if c.Orchestration {
			a.send(f, "orch", services[i], i, status)
		} else {
			if i > 0 {
				a.send(f, services[i-1], "bus", i, OK)
			}
			a.send(f, "bus", services[i], i, status)
		}
	}
	f := a.frame("the order comes in")
	a.send(f, "client", map[bool]string{true: "orch", false: "orders"}[c.Orchestration], 0, OK)
	for i, s := range SagaSteps {
		if i+1 == c.FailAt {
			actions = append(actions, s.Do+" FAILED")
			call(i, Failed, s.Do+" failed: compensate what is done, newest first")
			for j := len(done) - 1; j >= 0; j-- {
				undo := SagaSteps[j].Undo
				actions = append(actions, undo)
				call(j, Queued, "compensate: "+undo)
			}
			break
		}
		actions = append(actions, s.Do)
		done = append(done, s.Do)
		call(i, OK, s.Do)
	}
	var r Report
	r.Arch = a
	for i, act := range actions {
		r.say("%d. %s", i+1, act)
	}
	style := map[bool]string{true: "orchestration: the orchestrator drives every step and every compensation", false: "choreography: each service reacts to the previous one's event"}[c.Orchestration]
	r.say("%s", style)
	if c.FailAt > 0 {
		r.say("step %d failed; %d compensations ran in reverse order, leaving no partial order behind", c.FailAt, len(done))
	}
	r.put("steps done", float64(len(done)))
	compensations := 0
	if c.FailAt > 0 {
		compensations = len(done)
	}
	r.put("compensations", float64(compensations))
	return actions, r, nil
}

// String names a saga config.
func (c SagaConfig) String() string {
	return fmt.Sprintf("saga failing at %d (orchestration %v)", c.FailAt, c.Orchestration)
}
