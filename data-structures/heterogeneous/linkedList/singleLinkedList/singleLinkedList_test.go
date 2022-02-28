package singleLinkedList

import "testing"

var (
  list = New()
  x    = []interface{}{4, 5, 6, 7, 0, 1, 2, 3}
)

func TestSingleLinkedList(t *testing.T) {
  t.Run("New", TestListNew)
  t.Run("Insert", TestListInsert)
  t.Run("Next, Head and Tail Values", TestNextValue)
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
    if element.next != nil {
      nextList[i] = element.next.value
    }
  }
  lv := list.Values()
  t.Log("List -> ", lv)
  t.Log("Head (", list.head.value, ") and Tail (", list.tail, ")")
  t.Log("Next list -> ", nextList)
  nlv, nnextList := lv[1:], nextList[:len(nextList)-1]

  for i, _ := range nlv {
    for j, _ := range nnextList {
      if i == j {
        if nlv[i] == nnextList[j] {
          t.Log("The next value of the node ", nlv[i], " is ", nnextList[j])
        } else {
          t.Error("Expected The next value of the node ", nlv[i], " to be ", nnextList[j], " but got ", nnextList[j])
          break
        }
      }
    }
  }
}
