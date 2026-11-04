package design

import (
	"fmt"
	"math/rand"
)

// PaymentConfig is a payment system's traffic and failures.
type PaymentConfig struct {
	Payments     int   // payments customers make
	RetryPercent int   // chance the client times out and sends the same payment again
	FailPercent  int   // chance the payment processor declines
	Idempotent   bool  // the service honors idempotency keys
	AmountCents  int   // each payment's amount
	Seed         int64 // which payments retry and fail
}

// Payments models a payment service: each payment carries an idempotency
// key; the service records the key with the result, so a client that times
// out and retries gets the first result back instead of being charged twice.
// A charge that succeeds is written to a double-entry ledger, a debit from
// the customer and an equal credit to the merchant, so every entry has its
// opposite and the ledger always sums to zero, which reconciliation checks.
// With Idempotent false the retries are charged again. It leaves out the
// processor's own webhooks, refunds and currencies.
//
// Time: O(Payments). Space: O(Payments).
func Payments(c PaymentConfig) (Report, error) {
	if err := check(c.Payments >= 1 && c.Payments <= 100_000 && c.RetryPercent >= 0 && c.RetryPercent <= 100 && c.FailPercent >= 0 && c.FailPercent <= 100 && c.AmountCents > 0,
		"1 to 100,000 payments, percents 0 to 100, a positive amount"); err != nil {
		return Report{}, err
	}
	rng := rand.New(rand.NewSource(c.Seed))
	a := arch([]Component{
		{ID: "client", Kind: Client, Label: "Checkout"},
		{ID: "pay", Kind: Service, Label: "Payment service"},
		{ID: "keys", Kind: Database, Label: "Idempotency keys", Notes: "key -> result"},
		{ID: "psp", Kind: Service, Label: "Payment processor"},
		{ID: "ledger", Kind: Database, Label: "Ledger", Notes: "double entry"},
	}, "client>pay", "pay>keys:seen key?", "pay>psp:charge", "pay>ledger:debit + credit")
	results := map[int]bool{}
	var charged, declined, prevented, doubled int
	var ledger []int // signed entries: debits negative, credits positive
	id := 0
	for p := 0; p < c.Payments; p++ {
		tries := 1
		if rng.Intn(100) < c.RetryPercent {
			tries = 2
		}
		declinedNow := rng.Intn(100) < c.FailPercent
		for t := 0; t < tries; t++ {
			id++
			var f *Frame
			if a.recording() {
				f = a.frame("")
				a.send(f, "client", "pay", id, OK)
			}
			if _, seen := results[p]; seen && c.Idempotent {
				prevented++
				if f != nil {
					f.Note = fmt.Sprintf("payment %d retried with the same key: the stored result is returned, no second charge", p)
					a.send(f, "pay", "keys", id, Hit)
				}
				continue
			}
			if f != nil && c.Idempotent {
				a.send(f, "pay", "keys", id, Miss)
			}
			if declinedNow {
				declined++
				results[p] = false
				if f != nil {
					f.Note = fmt.Sprintf("payment %d declined by the processor: nothing written to the ledger", p)
					a.send(f, "pay", "psp", id, Failed)
				}
				continue
			}
			if t > 0 {
				doubled++
			}
			results[p] = true
			charged++
			ledger = append(ledger, -c.AmountCents, c.AmountCents) // customer debit, merchant credit
			if f != nil {
				f.Note = fmt.Sprintf("payment %d charged: ledger debit customer %d, credit merchant %d", p, c.AmountCents, c.AmountCents)
				a.send(f, "pay", "psp", id, OK)
				a.send(f, "pay", "ledger", id, OK)
			}
		}
	}
	balance := 0
	for _, e := range ledger {
		balance += e
	}
	entries := len(ledger)
	last := a.frame("totals")
	last.load("ledger", entries, 0, -1)
	last.load("keys", len(results), 0, percent(prevented, prevented+len(results)))
	var r Report
	r.Arch = a
	r.say("%d payments: %d charges, %d declined, %d retries answered from the idempotency key, %d customers charged twice", c.Payments, charged, declined, prevented, doubled)
	r.say("ledger: %d entries summing to %d: every debit has its credit", entries, balance)
	r.put("charges", float64(charged))
	r.put("double charges", float64(doubled))
	r.put("retries deduplicated", float64(prevented))
	r.put("ledger balance", float64(balance))
	return r, nil
}
