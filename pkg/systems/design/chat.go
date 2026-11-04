package design

import (
	"fmt"
	"math/rand"
	"sort"
)

// ChatConfig is a chat system's traffic.
type ChatConfig struct {
	Users          int   // people chatting, in one-to-one conversations with a neighbor
	Messages       int   // messages sent
	OfflinePercent int   // chance a recipient is offline when a message is sent
	MaxDelay       int   // most ticks the network may delay a message (reorders them)
	Seed           int64 // who talks, delays and presence
}

// Chat models one-to-one chat delivery: senders keep a WebSocket to a
// gateway, the chat service stamps each message with the next sequence
// number of its conversation, and delivers it to an online recipient or
// stores it in the recipient's offline inbox, delivered on reconnect. The
// network delays messages by up to MaxDelay ticks, so they can arrive out of
// order; the client shows them by sequence number, holding back any that
// arrive before an earlier one. Presence is a heartbeat per user. It leaves
// out groups, receipts and end-to-end encryption.
//
// Time: O(Messages log Messages). Space: O(Messages).
func Chat(c ChatConfig) (Report, error) {
	if err := check(c.Users >= 2 && c.Users <= 10_000 && c.Messages >= 1 && c.Messages <= 100_000 && c.OfflinePercent >= 0 &&
		c.OfflinePercent <= 100 && c.MaxDelay >= 0 && c.MaxDelay <= 100, "2 to 10,000 users, 1 to 100,000 messages, percents 0 to 100"); err != nil {
		return Report{}, err
	}
	rng := rand.New(rand.NewSource(c.Seed))
	a := arch([]Component{
		{ID: "alice", Kind: Client, Label: "Sender"},
		{ID: "gw", Kind: Balancer, Label: "WebSocket gateway"},
		{ID: "chat", Kind: Service, Label: "Chat service", Notes: "sequence per conversation"},
		{ID: "store", Kind: Database, Label: "Message store"},
		{ID: "presence", Kind: Cache, Label: "Presence", Notes: "heartbeats"},
		{ID: "inbox", Kind: Queue, Label: "Offline inbox"},
		{ID: "bob", Kind: Client, Label: "Recipient"},
	}, "alice>gw", "gw>chat", "chat>store", "chat>presence:online?", "chat>inbox:offline", "chat>bob:deliver", "inbox>bob:on reconnect")
	type msg struct {
		conv, seq, arrive int
	}
	seq := map[int]int{}
	var arrived []msg
	live, offline := 0, 0
	for i := 0; i < c.Messages; i++ {
		from := rng.Intn(c.Users)
		conv := from / 2
		seq[conv]++
		isOff := rng.Intn(100) < c.OfflinePercent
		delay := 0
		if c.MaxDelay > 0 {
			delay = rng.Intn(c.MaxDelay + 1)
		}
		arrive := i + delay
		if isOff {
			offline++
			arrive = c.Messages + c.MaxDelay + i // delivered after the recipient reconnects
		} else {
			live++
		}
		arrived = append(arrived, msg{conv, seq[conv], arrive})
		if a.recording() {
			f := a.frame(fmt.Sprintf("message %d in conversation %d: %s", seq[conv], conv, map[bool]string{true: "recipient offline, kept in the inbox", false: "delivered live"}[isOff]))
			a.send(f, "alice", "gw", i, OK)
			a.send(f, "gw", "chat", i, OK)
			a.send(f, "chat", "store", i, OK)
			if isOff {
				a.send(f, "chat", "presence", i, Miss)
				a.send(f, "chat", "inbox", i, Queued)
			} else {
				a.send(f, "chat", "presence", i, Hit)
				a.send(f, "chat", "bob", i, OK)
			}
		}
	}
	sort.SliceStable(arrived, func(i, j int) bool { return arrived[i].arrive < arrived[j].arrive })
	next := map[int]int{}
	held := map[int]map[int]bool{}
	reordered := 0
	for _, m := range arrived {
		if next[m.conv] == 0 {
			next[m.conv] = 1
		}
		if m.seq != next[m.conv] {
			reordered++
			if held[m.conv] == nil {
				held[m.conv] = map[int]bool{}
			}
			held[m.conv][m.seq] = true
			continue
		}
		next[m.conv]++
		for held[m.conv][next[m.conv]] {
			delete(held[m.conv], next[m.conv])
			next[m.conv]++
		}
	}
	last := a.frame("totals")
	last.load("inbox", offline, offline, -1)
	var r Report
	r.Arch = a
	r.say("%d messages: %d delivered live, %d kept in offline inboxes and delivered on reconnect", c.Messages, live, offline)
	r.say("%d arrived before an earlier message of their conversation and were held back: sequence numbers keep each conversation in order", reordered)
	r.put("live", float64(live))
	r.put("offline inbox", float64(offline))
	r.put("held back for order", float64(reordered))
	return r, nil
}
