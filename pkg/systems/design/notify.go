package design

import (
	"fmt"
	"math/rand"
)

// NotifyConfig is a notification system's traffic and failures.
type NotifyConfig struct {
	Notifications    int   // notifications requested
	DuplicatePercent int   // chance a producer sends the same request twice (its own retry)
	FailPercent      int   // chance a provider (push, email, SMS) call fails
	MaxRetries       int   // retries before a notification goes to the dead-letter queue
	Seed             int64 // which requests duplicate and which calls fail
}

// Notifications models a notification system: producers put requests on a
// queue with an idempotency key; workers drop any key already sent (so a
// producer's retry never notifies twice), call the provider, and on failure
// retry with exponential backoff (1, 2, 4... ticks) up to MaxRetries, after
// which the request goes to a dead-letter queue for someone to look at. It
// leaves out user preferences, rate limits per user and templates.
//
// Time: O(Notifications * MaxRetries). Space: O(Notifications).
func Notifications(c NotifyConfig) (Report, error) {
	if err := check(c.Notifications >= 1 && c.Notifications <= 100_000 && c.DuplicatePercent >= 0 && c.DuplicatePercent <= 100 &&
		c.FailPercent >= 0 && c.FailPercent <= 100 && c.MaxRetries >= 0 && c.MaxRetries <= 10, "1 to 100,000 notifications, percents 0 to 100, retries 0 to 10"); err != nil {
		return Report{}, err
	}
	rng := rand.New(rand.NewSource(c.Seed))
	a := arch([]Component{
		{ID: "app", Kind: Service, Label: "Producer service"},
		{ID: "queue", Kind: Queue, Label: "Notification queue"},
		{ID: "worker", Kind: Worker, Label: "Sender workers"},
		{ID: "dedupe", Kind: Cache, Label: "Sent keys", Notes: "idempotency keys"},
		{ID: "provider", Kind: Service, Label: "Push / email / SMS provider"},
		{ID: "dlq", Kind: Queue, Label: "Dead-letter queue"},
	}, "app>queue", "queue>worker", "worker>dedupe:seen?", "worker>provider:send", "worker>dlq:gave up")
	sent := map[int]bool{}
	var delivered, deduped, retries, dead, backoff int
	id := 0
	for n := 0; n < c.Notifications; n++ {
		copies := 1
		if rng.Intn(100) < c.DuplicatePercent {
			copies = 2
		}
		for k := 0; k < copies; k++ {
			id++
			var f *Frame
			if a.recording() {
				f = a.frame("")
				a.send(f, "app", "queue", id, Queued)
				a.send(f, "queue", "worker", id, OK)
			}
			if sent[n] {
				deduped++
				if f != nil {
					f.Note = fmt.Sprintf("notification %d again (the producer retried): key already sent, dropped", n)
					a.send(f, "worker", "dedupe", id, Hit)
				}
				continue
			}
			if f != nil {
				a.send(f, "worker", "dedupe", id, Miss)
			}
			ok := false
			for try := 0; try <= c.MaxRetries; try++ {
				if try > 0 {
					retries++
					backoff += 1 << (try - 1)
				}
				if rng.Intn(100) >= c.FailPercent {
					ok = true
					break
				}
				if f != nil {
					a.send(f, "worker", "provider", id, Failed)
				}
			}
			if ok {
				delivered++
				sent[n] = true
				if f != nil {
					f.Note = fmt.Sprintf("notification %d sent", n)
					a.send(f, "worker", "provider", id, OK)
				}
			} else {
				dead++
				if f != nil {
					f.Note = fmt.Sprintf("notification %d failed %d times: dead-lettered", n, c.MaxRetries+1)
					a.send(f, "worker", "dlq", id, Dropped)
				}
			}
		}
	}
	last := a.frame("totals")
	last.load("dlq", dead, dead, -1)
	last.load("dedupe", len(sent), 0, percent(deduped, deduped+len(sent)))
	var r Report
	r.Arch = a
	r.say("%d notifications: %d delivered, %d duplicate requests dropped by key, %d retries (%d ticks of backoff), %d dead-lettered",
		c.Notifications, delivered, deduped, retries, backoff, dead)
	r.put("delivered", float64(delivered))
	r.put("duplicates dropped", float64(deduped))
	r.put("retries", float64(retries))
	r.put("dead-lettered", float64(dead))
	return r, nil
}
