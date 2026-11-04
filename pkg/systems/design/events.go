package design

import (
	"fmt"
	"math/rand"
)

// CQRSConfig is an orders service with a separate read model.
type CQRSConfig struct {
	Commands int   // orders placed or changed (writes)
	Queries  int   // dashboard reads, spread among the commands
	LagMax   int   // most commands the projection trails the write model by
	Seed     int64 // which commands and when the projection catches up
}

// CQRS models command query responsibility segregation: commands change the
// write model (normalized, validated, one order at a time) and publish
// events; a projector consumes the events and updates a read model shaped
// for queries (here, counts of orders by status for a dashboard). Reads and
// writes then scale and evolve separately, but the read model trails the
// write model by the projection lag, so a query can be stale. It reports how
// often. It leaves out the event bus itself.
//
// Time: O(Commands + Queries). Space: O(Commands).
func CQRS(c CQRSConfig) (Report, error) {
	if err := check(c.Commands >= 1 && c.Commands <= 100_000 && c.Queries >= 0 && c.Queries <= 100_000 && c.LagMax >= 0, "1 to 100,000 commands, 0 to 100,000 queries, a lag of 0 or more"); err != nil {
		return Report{}, err
	}
	rng := rand.New(rand.NewSource(c.Seed))
	statuses := []string{"placed", "paid", "shipped"}
	status := make([]int, 0, c.Commands)
	var events []int // the order each command changed
	for i := 0; i < c.Commands; i++ {
		if len(status) == 0 || rng.Intn(3) == 0 {
			status = append(status, 0)
			events = append(events, len(status)-1)
		} else {
			o := rng.Intn(len(status))
			if status[o] < len(statuses)-1 {
				status[o]++
			}
			events = append(events, o)
		}
	}
	a := arch([]Component{
		{ID: "ui", Kind: Client, Label: "Shop"},
		{ID: "cmd", Kind: Service, Label: "Command side", Notes: "validates and writes"},
		{ID: "write", Kind: Database, Label: "Write model"},
		{ID: "bus", Kind: Queue, Label: "Order events"},
		{ID: "proj", Kind: Worker, Label: "Projector"},
		{ID: "read", Kind: Database, Label: "Read model", Notes: "counts by status"},
		{ID: "query", Kind: Service, Label: "Query side"},
		{ID: "dash", Kind: Client, Label: "Dashboard"},
	}, "ui>cmd", "cmd>write", "cmd>bus", "bus>proj", "proj>read", "dash>query", "query>read")
	stale := 0
	for q := 0; q < c.Queries; q++ {
		at := rng.Intn(c.Commands)
		lag := 0
		if c.LagMax > 0 {
			lag = rng.Intn(c.LagMax + 1)
		}
		behind := min(lag, at+1)
		if behind > 0 {
			stale++
		}
		if a.recording() {
			f := a.frame(fmt.Sprintf("query after command %d: the read model is %d events behind", at, behind))
			a.send(f, "ui", "cmd", at, OK)
			a.send(f, "cmd", "write", at, OK)
			a.send(f, "cmd", "bus", at, Queued)
			a.send(f, "bus", "proj", at, OK)
			a.send(f, "proj", "read", at, OK)
			a.send(f, "dash", "query", q, OK)
			a.send(f, "query", "read", q, map[bool]Status{true: Miss, false: Hit}[behind > 0])
			f.load("bus", behind, behind, -1)
		}
	}
	counts := make([]int, len(statuses))
	for _, s := range status {
		counts[s]++
	}
	var r Report
	r.Arch = a
	t := r.table("orders")
	for i, s := range statuses {
		t.add(s, counts[i])
	}
	r.say("%d commands produced %d events; the read model, once caught up, counts %d orders by status", c.Commands, len(events), len(status))
	r.say("%d of %d queries ran while the projection trailed by up to %d events: eventually consistent reads", stale, c.Queries, c.LagMax)
	r.put("events", float64(len(events)))
	r.put("stale queries", float64(stale))
	return r, nil
}

// Event is one change to a bank account in an event-sourced store.
type Event struct {
	Seq    int
	Kind   string // "deposit" or "withdraw"
	Amount int
}

// EventSourcingConfig is an account's history.
type EventSourcingConfig struct {
	Events        int   // events recorded
	SnapshotEvery int   // events between two snapshots of the balance (0: none)
	Seed          int64 // the amounts
}

// EventSourcing models event sourcing: the store keeps every change to an
// account as an append-only event (deposited 50, withdrew 20), never the
// balance itself; the current state is the replay of all events, and any
// past state the replay of a prefix, which gives an audit log for free.
// Replaying everything grows with the history, so a snapshot of the state
// every SnapshotEvery events lets a rebuild start from the last snapshot. It
// reports the events each rebuild applies and checks both give the same
// balance. Withdrawals that would overdraw are rejected and not recorded.
//
// Time: O(Events). Space: O(Events).
func EventSourcing(c EventSourcingConfig) ([]Event, Report, error) {
	if err := check(c.Events >= 1 && c.Events <= 1_000_000 && c.SnapshotEvery >= 0, "1 to 1,000,000 events, snapshots every 0 or more"); err != nil {
		return nil, Report{}, err
	}
	rng := rand.New(rand.NewSource(c.Seed))
	var log []Event
	balance, snapAt, snapBal, rejected := 0, 0, 0, 0
	for len(log) < c.Events {
		e := Event{Seq: len(log) + 1, Kind: "deposit", Amount: 10 + rng.Intn(90)}
		if rng.Intn(3) == 0 {
			e.Kind = "withdraw"
			if e.Amount > balance {
				rejected++
				continue
			}
		}
		log = append(log, e)
		balance = Apply(balance, e)
		if c.SnapshotEvery > 0 && len(log)%c.SnapshotEvery == 0 {
			snapAt, snapBal = len(log), balance
		}
	}
	full := Replay(0, log)
	fromSnap := Replay(snapBal, log[snapAt:])
	var r Report
	t := r.table("amount", "balance")
	b := 0
	for i, e := range log {
		b = Apply(b, e)
		if i < 10 {
			t.add(fmt.Sprintf("%d %s", e.Seq, e.Kind), e.Amount, b)
		}
	}
	r.say("%d events (%d overdrafts rejected, never recorded); the balance is the replay of all of them: %d", len(log), rejected, full)
	if c.SnapshotEvery > 0 {
		r.say("from the snapshot at event %d (balance %d), only %d events replay to the same %d", snapAt, snapBal, len(log)-snapAt, fromSnap)
	}
	r.put("balance", float64(full))
	r.put("events replayed (full)", float64(len(log)))
	r.put("events replayed (from snapshot)", float64(len(log)-snapAt))
	return log, r, nil
}

// Apply is the balance after one event.
//
// Time: O(1). Space: O(1).
func Apply(balance int, e Event) int {
	if e.Kind == "withdraw" {
		return balance - e.Amount
	}
	return balance + e.Amount
}

// Replay applies the events in order to a starting balance.
//
// Time: O(len(events)). Space: O(1).
func Replay(balance int, events []Event) int {
	for _, e := range events {
		balance = Apply(balance, e)
	}
	return balance
}

// OutboxConfig is an orders service publishing events about its orders.
type OutboxConfig struct {
	Orders       int   // orders created
	CrashPercent int   // chance the service crashes between its database write and its publish
	Outbox       bool  // write the event to an outbox table in the same transaction
	Seed         int64 // which steps crash
}

// TransactionalOutbox models the dual-write problem and its fix. Writing
// the order to the database and then publishing an event to the broker is
// two writes with no transaction around them: a crash between them leaves
// an order nobody hears about (a lost event). With an outbox, the order and
// its event row commit in one database transaction, and a relay reads the
// outbox and publishes; a relay crash after publishing but before marking
// the row sent republishes it, so delivery is at least once and consumers
// drop duplicates by event id. Change data capture is the same idea reading
// the database's own log.
//
// Time: O(Orders). Space: O(Orders).
func TransactionalOutbox(c OutboxConfig) (Report, error) {
	if err := check(c.Orders >= 1 && c.Orders <= 100_000 && c.CrashPercent >= 0 && c.CrashPercent <= 100, "1 to 100,000 orders, a crash percent of 0 to 100"); err != nil {
		return Report{}, err
	}
	rng := rand.New(rand.NewSource(c.Seed))
	a := arch([]Component{
		{ID: "api", Kind: Service, Label: "Orders service"},
		{ID: "db", Kind: Database, Label: "Orders database", Notes: "orders + outbox table"},
		{ID: "relay", Kind: Worker, Label: "Outbox relay"},
		{ID: "broker", Kind: Queue, Label: "Message broker"},
		{ID: "consumer", Kind: Service, Label: "Consumers", Notes: "dedupe by event id"},
	}, "api>db:commit", "api>broker:publish", "relay>db:read outbox", "relay>broker:publish", "broker>consumer")
	var lost, dupes, delivered int
	for o := 0; o < c.Orders; o++ {
		crash := rng.Intn(100) < c.CrashPercent
		var f *Frame
		if a.recording() {
			f = a.frame("")
			a.send(f, "api", "db", o, OK)
		}
		switch {
		case !c.Outbox && crash:
			lost++
			if f != nil {
				f.Note = fmt.Sprintf("order %d committed, then a crash before publishing: the event is lost", o)
				a.send(f, "api", "broker", o, Dropped)
			}
		case !c.Outbox:
			delivered++
			if f != nil {
				f.Note = fmt.Sprintf("order %d committed and published", o)
				a.send(f, "api", "broker", o, OK)
				a.send(f, "broker", "consumer", o, OK)
			}
		default:
			delivered++
			if crash {
				dupes++
			}
			if f != nil {
				f.Note = fmt.Sprintf("order %d and its event committed together; the relay publishes it%s", o, map[bool]string{true: " twice after a relay crash; the consumer drops the copy", false: ""}[crash])
				a.send(f, "relay", "db", o, OK)
				a.send(f, "relay", "broker", o, OK)
				a.send(f, "broker", "consumer", o, OK)
				if crash {
					a.send(f, "broker", "consumer", o, Hit)
				}
			}
		}
	}
	var r Report
	r.Arch = a
	if c.Outbox {
		r.say("outbox: all %d events delivered at least once; %d duplicates dropped by the consumers", delivered, dupes)
	} else {
		r.say("dual write: %d of %d events lost to a crash between the commit and the publish", lost, c.Orders)
	}
	r.put("lost events", float64(lost))
	r.put("duplicates", float64(dupes))
	r.put("delivered", float64(delivered))
	return r, nil
}

// CDCConfig is a database whose changes flow to downstream systems.
type CDCConfig struct {
	Changes  int   // rows changed
	Rows     int   // distinct rows
	LagTicks int   // ticks the slowest consumer trails the log by
	Seed     int64 // which rows change
}

// ChangeDataCapture models change data capture (Debezium, DynamoDB Streams):
// a connector tails the database's write-ahead log and publishes every
// change, in commit order, to a topic partitioned by row key; consumers keep
// a search index and a cache in step without the application writing to
// them twice. Ordering per key holds because each key's changes share a
// partition; each consumer tracks its offset in the log, so it resumes
// after a restart. It reports the offsets and lag.
//
// Time: O(Changes). Space: O(Rows).
func ChangeDataCapture(c CDCConfig) (Report, error) {
	if err := check(c.Changes >= 1 && c.Changes <= 100_000 && c.Rows >= 1 && c.LagTicks >= 0, "1 to 100,000 changes, at least 1 row, a lag of 0 or more"); err != nil {
		return Report{}, err
	}
	rng := rand.New(rand.NewSource(c.Seed))
	a := arch([]Component{
		{ID: "app", Kind: Service, Label: "Application"},
		{ID: "db", Kind: Database, Label: "Database", Notes: "write-ahead log"},
		{ID: "cdc", Kind: Worker, Label: "CDC connector", Notes: "tails the log"},
		{ID: "topic", Kind: Queue, Label: "Change topic", Notes: "partitioned by key"},
		{ID: "search", Kind: Search, Label: "Search index"},
		{ID: "cache", Kind: Cache, Label: "Cache invalidator"},
	}, "app>db:write once", "db>cdc:log", "cdc>topic", "topic>search", "topic>cache")
	version := map[int]int{}
	searchSeen := map[int]int{}
	outOfOrder := 0
	for i := 0; i < c.Changes; i++ {
		k := rng.Intn(c.Rows)
		version[k]++
		if searchSeen[k] > version[k] {
			outOfOrder++
		}
		searchSeen[k] = version[k]
		if a.recording() {
			f := a.frame(fmt.Sprintf("log offset %d: row %d version %d", i, k, version[k]))
			a.send(f, "app", "db", i, OK)
			a.send(f, "db", "cdc", i, OK)
			a.send(f, "cdc", "topic", i, Queued)
			a.send(f, "topic", "search", i, OK)
			a.send(f, "topic", "cache", i, OK)
		}
	}
	last := a.frame("totals")
	last.load("topic", c.Changes, min(c.LagTicks, c.Changes), -1)
	var r Report
	r.Arch = a
	r.say("%d changes streamed from the log in commit order; the application wrote each once", c.Changes)
	r.say("search consumer at offset %d, %d behind; %d changes applied out of order per key (none: one partition per key)", c.Changes-min(c.LagTicks, c.Changes), min(c.LagTicks, c.Changes), outOfOrder)
	r.put("changes", float64(c.Changes))
	r.put("consumer lag", float64(min(c.LagTicks, c.Changes)))
	r.put("out of order", float64(outOfOrder))
	return r, nil
}
