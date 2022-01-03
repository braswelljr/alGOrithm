package queue

import (
  "reflect"
  "testing"
)

func TestQueueInit(t *testing.T) {
  q := New()
  if q.Size() != 0 {
    t.Errorf("Expected size 0, got %d", q.Size())
  }
}

// test type of size of list
func TestQueueSize(t *testing.T) {
  q := New()
  q.Enqueue(5)
  q.Enqueue(7)
  if reflect.TypeOf(q.Size()).String() != "int" {
    t.Errorf("Size() should return an int")
    return
  }

  t.Log("Passed Size() test. With size of ", q.Size(), " Queue : ", q)
}

func TestQueueEnqueue(t *testing.T) {
  q := New()
  fl := q.Size()
  q.Enqueue(7)
  ll := q.Size()

  if ll <= fl {
    t.Error("Expected ", ll, " to be greater than ", fl, ". With Queue : ", q)
    return
  }

  t.Log("Passed Enqueue() test. With size before enqueue : ", fl, " and size after enqueue : ", ll, " Queue : ", q)
}

// Test Dequeue
func TestQueueDequeue(t *testing.T) {
  q := New()
  q.Enqueue(7)
  q.Enqueue(6)
  q.Enqueue(4)
  q.Enqueue(10)
  q.Enqueue(100)
  fl := q.Size()
  q.Dequeue()
  ll := q.Size()
  q.Reverse()

  if ll >= fl {
    t.Error("Expected ", ll, " to be greater than ", fl, ". With Queue : ", q)
    return
  }

  t.Log("Passed Dequeue() test. With size before Dequeue : ", fl, " and size after Dequeue : ", ll, " Queue : ", q)
}
