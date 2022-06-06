package orderedList

import "testing"

var (
  list = New()
  x    = []interface{}{4, 5, 6, 7, 0, 1, 2, 3}
)

func TestOrderedList(t *testing.T) {
  t.Run("New", TestListNew)
  //t.Run("Insert", TestListInsert)
  //t.Run("Next, Head and Tail Values", TestNextValue)
}

func TestListNew(t *testing.T) {
  if list == nil {
    t.Error("New() should not return nil")
    return
  }
  t.Log("New List initialized ", list)
}
