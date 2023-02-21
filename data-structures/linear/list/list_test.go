package list

import (
	"reflect"
	"testing"
)

var (
	x    = []interface{}{4, 5, 6, 7, 0, 1, 2, 3}
	list = New()
)

// TestList tests the list data structure.
func TestList(t *testing.T) {
	l := New()
	t.Logf("New List: %v", l)
	y := New(x)
	t.Logf("New List with List: %v", y)
	number := 7
	l.Push(1)
	t.Logf("With Push: %v", l)
	l.Pop()
	t.Logf("With Pop: %v", l)
	l.Push(7)
	t.Logf("With Push: %v", l)
	t.Logf("if %v then it contains %v\n", l.Contains(number), number)
	l.RemoveAt(2)
	t.Logf("With RemoveAt: %v", l)
	l.Empty()
	t.Logf("With Clear: %v", l)
	l.Join(x)
	t.Logf("With Join: %v", l)
	l.Insert(2, 3)
	t.Logf("With Insert: %v", l)
}

// test type of size of list
func TestListSize(t *testing.T) {
	if reflect.TypeOf(list.Size()).String() != "int" {
		t.Errorf("Size() should return an int")
		return
	}

	t.Logf("Passed Size() test")
}

// test empty list clear
func TestListClear(t *testing.T) {
	list.Empty()
	if list.Size() != 0 {
		t.Errorf("Clear() should return an empty list")
		return
	}

	t.Logf("Passed Clear() test")
}
