package orderedList

import "testing"

var (
	list = New()
	x    = []interface{}{4, 5, 6, 7, 0, 1, 2, 3}
)

func TestOrderedList(t *testing.T) {
	t.Run("New", TestListNew)
	t.Run("Insert", TestListInsert)
	//t.Run("Next, Head and Tail Values", TestNextValue)
}

func TestListNew(t *testing.T) {
	if list == nil {
		t.Error("New() should not return nil")
		return
	}
	t.Log("New List initialized ", list)
}

func TestListInsert(t *testing.T) {
	ol := len(list.list) // old length (length before insert)
	list.Insert(5)
	nl := len(list.list) // new length (length after insert)
	if nl <= ol {
		t.Error("Could not insert item into List")
		return
	}
	t.Log("Item ", list.list[len(list.list)-1].value, " was inserted into list successfully!")
}
