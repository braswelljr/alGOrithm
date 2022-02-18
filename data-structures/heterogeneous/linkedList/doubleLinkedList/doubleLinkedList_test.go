package doubleLinkedList

import "testing"

var (
	list = New()
	x    = []interface{}{4, 5, 6, 7, 0, 1, 2, 3}
)

func TestDoubleLinkedList(t *testing.T) {
	t.Run("New", TestListNew)
	t.Run("Insert", TestListInsert)
}

func TestListNew(t *testing.T) {
	if list == nil {
		t.Error("New() should not return nil")
		return
	}
	t.Log("New List initialized ", list)
}

func TestListInsert(t *testing.T) {
	for _, element := range x {
		// length before insert
		lb := len(list.elements)
		// Insert
		list.Insert(element)
		// length after insert
		la := len(list.elements)
		// check length after insert
		if la <= lb || list.elements[len(list.elements)-1].value != element {
			t.Error("Insert() should insert element at the end of the list")
			break
		}
		t.Log("Inserted ", element, " at the end of the list")
	}
	t.Log("List ", list.Values())
}
