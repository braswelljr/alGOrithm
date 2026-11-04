package ipc

import (
	"strings"
	"testing"
)

func TestPipeBlocksAndEnds(t *testing.T) {
	r, err := Pipe("writer: hello world\nwriter: close\nreader: sleep\nreader: sleep\nreader: 3\nreader: 8\nreader: 8\nreader: 8", 4)
	if err != nil {
		t.Fatal(err)
	}
	if r.Counts["bytes read"] != 11 || r.Counts["writer blocked"] == 0 {
		t.Errorf("counts %v", r.Counts)
	}
	if !strings.Contains(strings.Join(r.Log, "\n"), "end of file") {
		t.Error("the last read should see end of file")
	}
}

func TestMessageQueuePriority(t *testing.T) {
	r, _ := MessageQueue("send 1 a\nsend 5 b\nsend 1 c\nreceive\nreceive\nreceive", 4)
	got := strings.Join(r.Log[3:], "|")
	if !strings.Contains(r.Log[3], `"b"`) || !strings.Contains(r.Log[4], `"a"`) || !strings.Contains(r.Log[5], `"c"`) {
		t.Errorf("order: %s", got)
	}
}

func TestSharedMemoryNeedsTheSemaphore(t *testing.T) {
	torn := false
	for seed := int64(1); seed < 20; seed++ {
		if r, _ := SharedMemory(6, true, seed); r.Counts["torn"] != 0 {
			t.Fatalf("seed %d: a synchronized read was torn", seed)
		}
		if r, _ := SharedMemory(6, false, seed); r.Counts["torn"] > 0 {
			torn = true
		}
	}
	if !torn {
		t.Error("some seed should tear an unsynchronized read")
	}
}

func TestRPCDedupe(t *testing.T) {
	for seed := int64(1); seed < 30; seed++ {
		if r, _ := RPC("add 2 3", 40, 8, true, seed); r.Counts["executions"] > 1 {
			t.Fatalf("seed %d: deduped RPC ran %d times", seed, r.Counts["executions"])
		}
	}
	twice := false
	for seed := int64(1); seed < 30; seed++ {
		if r, _ := RPC("add 2 3", 40, 8, false, seed); r.Counts["executions"] > 1 {
			twice = true
		}
	}
	if !twice {
		t.Error("some seed should lose a reply and run the call twice")
	}
}

func TestSocketBacklog(t *testing.T) {
	r, _ := LocalSocket(10, 1, 5, 1)
	if r.Counts["accepted"]+r.Counts["refused"] != 10 || r.Counts["refused"] == 0 {
		t.Errorf("counts %v", r.Counts)
	}
}
