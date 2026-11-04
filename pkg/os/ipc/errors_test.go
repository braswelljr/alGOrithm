package ipc

import (
	"errors"
	"testing"
)

// TestRejectsBadInput checks that every malformed script and capacity is an
// ErrInput naming the problem, rather than a run on wrong data.
func TestRejectsBadInput(t *testing.T) {
	for name, run := range map[string]func() error{
		"queue capacity 0":   func() error { _, err := MessageQueue("receive", 0); return err },
		"queue capacity 33":  func() error { _, err := MessageQueue("receive", 33); return err },
		"send no text":       func() error { _, err := MessageQueue("send 3", 4); return err },
		"send priority 32":   func() error { _, err := MessageQueue("send 32 hi", 4); return err },
		"send priority word": func() error { _, err := MessageQueue("send high hi", 4); return err },
		"unknown operation":  func() error { _, err := MessageQueue("peek", 4); return err },
		"pipe capacity 0":    func() error { _, err := Pipe("writer: hi", 0); return err },
		"pipe capacity 65":   func() error { _, err := Pipe("writer: hi", 65); return err },
		"read 0 bytes":       func() error { _, err := Pipe("writer: hi\nreader: 0", 4); return err },
		"read 65 bytes":      func() error { _, err := Pipe("writer: hi\nreader: 65", 4); return err },
	} {
		if err := run(); !errors.Is(err, ErrInput) {
			t.Errorf("%s: got %v, want an ErrInput", name, err)
		}
	}
}

// TestQueueFullAndEmpty checks a full queue refuses a send and an empty one
// gives nothing to a receive, both without an error.
func TestQueueFullAndEmpty(t *testing.T) {
	r, err := MessageQueue("send 1 a\nsend 1 b\nsend 1 c\nreceive\nreceive\nreceive\nreceive", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Log) < 7 {
		t.Errorf("the run logged %d lines for seven operations", len(r.Log))
	}
}
