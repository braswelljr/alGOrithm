package circularDoubleLinkedList

import "testing"

var (
	list = New()
	x    = []interface{}{4, 5, 6, 7, 0, 1, 2, 3}
)

func TestDoubleLinkedList(t *testing.T) {
	t.Run("New", TestListNew)
	t.Run("Insert", TestListInsert)
	t.Run("Previous, Next, Head and Tail Values", TestNextPrevValue)
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

func TestNextPrevValue(t *testing.T) {
	// Insert
	for _, element := range x {
		list.Insert(element)
	}
	nextList := make([]interface{}, len(list.elements))
	prevList := make([]interface{}, len(list.elements))
	for i, element := range list.elements {
		if element.next != nil && element.prev != nil {
			nextList[i] = element.next.value
			prevList[i] = element.prev.value
		}
	}
	lv := list.Values()
	t.Log("List -> ", lv)
	t.Log("Head (", list.head, ") and Tail (", list.tail, ")")
	t.Log("Prev list -> ", prevList)
	t.Log("Next list -> ", nextList)

	for i, node := range lv {
		for j, next := range nextList {
			if i == j && (i+1 < len(lv)) && (i-1 >= 0) {
				if lv[0] == nextList[len(lv)-1] || lv[i+1] == next {
					t.Log("The next value of the node ", node, " is ", next)
				} else {
					t.Error("Expected The next value of the node ", node, " to be ", lv[i+1], " but got ", next)
					break
				}
			}
		}
	}

	for i, node := range lv {
		for j, previous := range prevList {
			if i == j && (i+1 < len(lv)) && (i-1 >= 0) {
				if lv[len(lv)-1] == prevList[0] || lv[i-1] == previous {
					t.Log("The previous value of the node ", node, " is ", previous)
				} else {
					t.Error("Expected The previous value of the node ", node, " to be ", lv[i-1], " but got ", previous)
					break
				}
			}
		}
	}
}
