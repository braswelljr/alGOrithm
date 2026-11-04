package processes

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestLifecycle(t *testing.T) {
	tr, err := Lifecycle("admit dispatch wait wake dispatch timeout dispatch exit")
	if err != nil || tr[len(tr)-1].To != Terminated {
		t.Fatalf("%v %v", tr, err)
	}
	if _, err := Lifecycle("admit dispatch wait dispatch"); err == nil {
		t.Error("a waiting process cannot be dispatched")
	}
}

func TestForkTreeZombiesAndOrphans(t *testing.T) {
	r, err := ForkTree("2 fork\n3 fork\n3 exit 0\n4 exit 1\n2 wait")
	if err != nil {
		t.Fatal(err)
	}
	state := map[int]string{}
	for _, p := range r.Procs {
		state[p.PID] = p.State
	}
	if state[3] != "reaped" || state[4] != "reaped" {
		t.Errorf("states %v: 3 reaped by the shell, 4 adopted and reaped by init", state)
	}
}

func TestSignalsAndThreadModels(t *testing.T) {
	r, _ := Signals("handle SIGINT\nblock SIGINT\nsend SIGINT\nsend SIGINT\nignore SIGKILL\nunblock SIGINT")
	if !r.Alive {
		t.Error("a handled SIGINT keeps the process alive")
	}
	r, _ = Signals("ignore SIGKILL\nsend SIGKILL")
	if r.Alive {
		t.Error("SIGKILL cannot be ignored")
	}
	end := func(tl Timeline) int { return tl[len(tl)-1].Tick }
	one, _ := ThreadModels(ManyToOne, 3, 1, 2, 3, 4)
	oto, _ := ThreadModels(OneToOne, 3, 3, 2, 3, 4)
	if end(oto) >= end(one) {
		t.Errorf("one-to-one finishes at %d, many-to-one at %d", end(oto), end(one))
	}
	m, _ := ThreadsVsProcesses(3, 5, true)
	p, _ := ThreadsVsProcesses(3, 5, false)
	if m.Parent != 15 || p.Parent != 0 {
		t.Errorf("threads %d, processes %d", m.Parent, p.Parent)
	}
}

func TestSystemCalls(t *testing.T) {
	r, err := SystemCalls("compute 3\nopen /etc/hosts\nread 3 100\nwrite 1 hello world\nclose 3\nread 3 10\n\ngetpid\nfork\nopen a\nopen b\nclose 3\nopen c")
	if err != nil {
		t.Fatal(err)
	}
	wantLog := []string{
		"open /etc/hosts -> fd 3 (trap to kernel, 3 ticks, return to user)",
		"read 3 100 -> 100 bytes from /etc/hosts (trap to kernel, 4 ticks, return to user)",
		"write 1 hello world -> 11 bytes to stdout (trap to kernel, 3 ticks, return to user)",
		"close 3 -> ok (trap to kernel, 1 ticks, return to user)",
		"read 3 10 -> -1 EBADF (bad file descriptor) (trap to kernel, 1 ticks, return to user)",
		"getpid -> 42 (trap to kernel, 1 ticks, return to user)",
		"fork -> child 43 (trap to kernel, 5 ticks, return to user)",
		"open a -> fd 3 (trap to kernel, 3 ticks, return to user)",
		"open b -> fd 4 (trap to kernel, 3 ticks, return to user)",
		"close 3 -> ok (trap to kernel, 1 ticks, return to user)",
		"open c -> fd 3 (trap to kernel, 3 ticks, return to user)",
	}
	if !reflect.DeepEqual(r.Log, wantLog) {
		t.Errorf("log\n%s\nwant\n%s", strings.Join(r.Log, "\n"), strings.Join(wantLog, "\n"))
	}
	if r.UserTicks != 3 || r.KernelTicks != 28 || r.Switches != 22 {
		t.Errorf("user %d, kernel %d, switches %d; want 3, 28, 22", r.UserTicks, r.KernelTicks, r.Switches)
	}
	// every tick is one event: compute, one trap per call, and the kernel's work
	if len(r.Timeline) != 3+11+28 {
		t.Fatalf("%d events", len(r.Timeline))
	}
	kernel := 0
	for i, e := range r.Timeline {
		if e.Tick != i {
			t.Fatalf("event %d at tick %d", i, e.Tick)
		}
		if e.Lane == "kernel" {
			kernel++
			if e.State != "running" {
				t.Errorf("kernel event %v", e)
			}
		}
	}
	if kernel != r.KernelTicks {
		t.Errorf("%d kernel events, %d kernel ticks", kernel, r.KernelTicks)
	}

	t.Run("an unknown descriptor fails with EBADF", func(t *testing.T) {
		r, _ := SystemCalls("write x hi\nclose 9")
		for _, l := range r.Log {
			if !strings.Contains(l, "EBADF") {
				t.Errorf("log %q", l)
			}
		}
		if r.KernelTicks != 2 {
			t.Errorf("a failed call costs one tick: %d", r.KernelTicks)
		}
	})

	for name, script := range map[string]string{
		"compute zero":    "compute 0",
		"compute too big": "compute 101",
		"compute word":    "compute x",
		"compute nothing": "compute",
		"unknown call":    "getpid\nexecve x",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := SystemCalls(script); !errors.Is(err, ErrInput) {
				t.Errorf("err %v, want ErrInput", err)
			}
		})
	}
	if _, err := SystemCalls("getpid\nsleep 1"); err == nil || !strings.Contains(err.Error(), "line 2") {
		t.Errorf("err %v, want line 2", err)
	}
}

func TestForkTreeErrors(t *testing.T) {
	r, err := ForkTree("2 fork\n3 exec ls\n2 wait\n3 exit 7\n2 fork\n4 fork\n4 exit")
	if err != nil {
		t.Fatal(err)
	}
	byPID := map[int]Proc{}
	for _, p := range r.Procs {
		byPID[p.PID] = p
	}
	if p := byPID[3]; p.Program != "ls" || p.State != "reaped" || p.Status != 7 {
		t.Errorf("3: %+v, want ls reaped with status 7 by the blocked wait", p)
	}
	if p := byPID[5]; p.Parent != 1 || p.State != "running" || p.Program != "sh" {
		t.Errorf("5: %+v, want an orphan adopted by init", p)
	}
	if p := byPID[4]; p.State != "zombie" {
		t.Errorf("4: %+v, want a zombie until 2 waits", p)
	}
	for name, script := range map[string]string{
		"not a process":  "9 fork",
		"not a number":   "x fork",
		"no call":        "2",
		"exec no prog":   "2 exec",
		"bad status":     "2 fork\n3 exit z",
		"init exits":     "1 exit 0",
		"unknown call":   "2 spawn",
		"zombie forking": "2 fork\n3 exit\n3 fork",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ForkTree(script); !errors.Is(err, ErrInput) {
				t.Errorf("err %v, want ErrInput", err)
			}
		})
	}
}

func TestSignalsRules(t *testing.T) {
	r, err := Signals("block INT\nsend SIGINT\nsend sigint\nblock SIGKILL\nhandle SIGSTOP\nsend SIGCHLD\nsend SIGCONT\nsend SIGSTOP\nblock SIGTERM\nsend SIGTERM")
	if err != nil {
		t.Fatal(err)
	}
	if !r.Alive || !reflect.DeepEqual(r.Pending, []string{"SIGINT", "SIGTERM"}) {
		t.Errorf("alive %v, pending %v", r.Alive, r.Pending)
	}
	log := strings.Join(r.Log, "\n")
	for _, s := range []string{"already pending", "cannot be blocked", "refused (EINVAL)", "delivered: ignored", "told to continue", "told to stop"} {
		if !strings.Contains(log, s) {
			t.Errorf("log lacks %q:\n%s", s, log)
		}
	}
	r, _ = Signals("ignore SIGTERM\nsend SIGTERM\ndefault SIGTERM\nsend SIGTERM")
	if r.Alive {
		t.Error("restoring the default lets SIGTERM end the process")
	}
	for name, script := range map[string]string{
		"one word":       "send",
		"unknown signal": "send SIGFOO",
		"unknown step":   "raise SIGINT",
		"after the end":  "send SIGKILL\nsend SIGINT",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Signals(script); !errors.Is(err, ErrInput) {
				t.Errorf("err %v, want ErrInput", err)
			}
		})
	}
}
