package tuple

import "testing"

func TestTuple(t *testing.T) {
  tuple := New([]interface{}{6, 5, "hey"})
  t.Log("Tuple : ", tuple.ToSlice())
}

func TestTupleToString(t *testing.T) {
  tuple := New([]interface{}{6, 5})
  t.Log("Tuple : ", tuple.ToString())
}
