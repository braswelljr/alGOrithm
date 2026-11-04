package core

import (
	"github.com/braswelljr/alGOrithm/pkg/os/ipc"
)

// ipcEntries is every entry of the ipc category.
func ipcEntries() []Entry {
	return simEntries(ipcSims())
}

// ipcOutput turns an ipc result into an Output.
func ipcOutput(r ipc.Result, err error) (Output, error) {
	if err != nil {
		return Output{}, err
	}
	out := Output{Lines: r.Log, Summary: summary(r.Counts)}
	for _, e := range r.Timeline {
		out.Timeline = append(out.Timeline, tick(e.Lane, e.What, e.State, e.Tick))
	}
	return out, nil
}

//nolint:dupl // rows of one table share their shape; each is a different simulation
var ipcSimTable = []sim{
	{
		id: "pipe", name: "Pipe", file: "os/ipc/pipe.go", fn: "Pipe",
		summary: "A writer and a reader on a small pipe: writes block when it is full, reads when it is empty, and close means end of file.",
		time:    "O(bytes + calls)", room: "O(capacity)", input: Text,
		example: Input{Text: "writer: hello world\nwriter: close\nreader: sleep\nreader: sleep\nreader: 3\nreader: 8\nreader: 8\nreader: 8", Options: map[string]int{"capacity": 4}},
		options: []Option{{"capacity", "pipe capacity in bytes", 4, 1, 64}},
		run:     func(in Input) (Output, error) { return ipcOutput(ipc.Pipe(in.Text, in.opt("capacity", 4))) },
		args:    func(in Input) []any { return []any{in.Text, in.opt("capacity", 4)} },
	},
	{
		id: "message-queue", name: "Message Queue", file: "os/ipc/messagequeue.go", fn: "MessageQueue",
		summary: "Messages keep their boundaries and come out by priority, oldest first among equals; a full queue says EAGAIN.",
		time:    "O(n log c)", room: "O(c)", input: Text,
		example: Input{Text: "send 1 low\nsend 5 urgent\nsend 1 also low\nreceive\nreceive\nsend 3 normal\nreceive\nreceive\nreceive", Options: map[string]int{"capacity": 4}},
		options: []Option{{"capacity", "queue capacity", 4, 1, 32}},
		run:     func(in Input) (Output, error) { return ipcOutput(ipc.MessageQueue(in.Text, in.opt("capacity", 4))) },
		args:    func(in Input) []any { return []any{in.Text, in.opt("capacity", 4)} },
	},
	{
		id: "shared-memory", name: "Shared Memory and Torn Reads", file: "os/ipc/sharedmemory.go", fn: "SharedMemory",
		summary: "Two processes share a record: without a semaphore the reader can copy it half written.",
		time:    "O(rounds)", room: "O(rounds)", input: Simulation,
		example: Input{Options: map[string]int{"rounds": 5, "sync": 0, "seed": 2}},
		options: []Option{{"rounds", "records written", 5, 1, 30}, {"sync", "use a semaphore (1) or not (0)", 0, 0, 1}, seedOption},
		run: func(in Input) (Output, error) {
			return ipcOutput(ipc.SharedMemory(in.opt("rounds", 5), on(in, "sync", 0), int64(in.opt("seed", 2))))
		},
		args: func(in Input) []any { return []any{in.opt("rounds", 5), on(in, "sync", 0), in.opt("seed", 2)} },
	},
	{
		id: "local-socket", name: "Local Socket Server", file: "os/ipc/socket.go", fn: "LocalSocket",
		summary: "Clients connect to a Unix-domain socket: queued in the listen backlog until accepted, refused when it is full.",
		time:    "O(T + clients)", room: "O(clients)", input: Simulation,
		example: Input{Options: map[string]int{"clients": 8, "backlog": 2, "accept": 2, "seed": 1}},
		options: []Option{{"clients", "clients", 8, 1, 30}, {"backlog", "listen backlog", 2, 0, 30}, {"accept", "ticks per accept", 2, 1, 10}, seedOption},
		run: func(in Input) (Output, error) {
			return ipcOutput(ipc.LocalSocket(in.opt("clients", 8), in.opt("backlog", 2), in.opt("accept", 2), int64(in.opt("seed", 1))))
		},
		args: func(in Input) []any {
			return []any{in.opt("clients", 8), in.opt("backlog", 2), in.opt("accept", 2), in.opt("seed", 1)}
		},
	},
	{
		id: "remote-procedure-call", name: "Remote Procedure Call", file: "os/ipc/rpc.go", fn: "RPC",
		summary: "Stubs marshal a call over a lossy network; a lost reply makes a retry run it twice, unless the server dedupes by request id.",
		time:    "O(retries)", room: "O(retries)", input: Text,
		example: Input{Text: "add 2 3", Options: map[string]int{"loss": 40, "retries": 4, "dedupe": 0, "seed": 3}},
		options: []Option{{"loss", "loss percent", 40, 0, 90}, {"retries", "retries", 4, 0, 10}, {"dedupe", "server dedupes by request id", 0, 0, 1}, seedOption},
		run: func(in Input) (Output, error) {
			return ipcOutput(ipc.RPC(in.Text, in.opt("loss", 40), in.opt("retries", 4), on(in, "dedupe", 0), int64(in.opt("seed", 3))))
		},
		args: func(in Input) []any {
			return []any{in.Text, in.opt("loss", 40), in.opt("retries", 4), on(in, "dedupe", 0), in.opt("seed", 3)}
		},
	},
}

func ipcSims() []sim {
	return ipcSimTable
}
