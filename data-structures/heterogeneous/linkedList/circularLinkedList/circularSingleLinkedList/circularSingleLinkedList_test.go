package circularSingleLinkedList

import "testing"

var (
  list = New()
  x    = []interface{}{4, 5, 6, 7, 0, 1, 2, 3}
)

func TestCircularSingleLinkedList(t *testing.T) {
  t.Run("New", TestListNew)
  t.Run("Insert", TestListInsert)
  t.Run("Next", TestNextValue)
  t.Run("Remove", TestListRemove)
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

func TestNextValue(t *testing.T) {
  // Insert
  for _, element := range x {
    list.Insert(element)
  }
  nextList := make([]interface{}, len(list.elements))
  for i, element := range list.elements {
    nextList[i] = element.next.value
  }
  t.Log("List -> ", list.Values())
  t.Log("Next list -> ", nextList)
  t.Log("Head (", list.head, ") - Tail (", list.tail, ")")
}

func TestListRemove(t *testing.T) {
  for _, element := range x {
    // Insert
    list.Insert(element)
  }
  element := 0
  // length before remove
  lb := len(list.elements)
  list.Remove(element)
  list.RemoveAt(4)
  list.Remove(7)
  list.RemoveAt(0)
  // length after remove
  la := len(list.elements)
  // check length after remove
  if lb < la {
    t.Error("Remove() should remove element from the list")
  } else {
    t.Log("Removed ", element, " from the list")
  }
  t.Log("List ", list.Values())
}
