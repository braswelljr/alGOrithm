package set

import (
  "reflect"
  "testing"
)

func TestSetInit(t *testing.T) {
  s := New()
  t.Log("Set:", s)
}

func TestSetAdd(t *testing.T) {
  s := New()
  s.Add(1)
  s.Add(3)
  s.Add([...]int{1, 2, 3})
  s.Add(4)
  s.Add(5)
  t.Log("Set:", s)
}

func TestSetSize(t *testing.T) {
  s := New()
  s.Add(1)
  s.Add(3)
  s.Add([...]int{1, 2, 3})
  s.Add(4)
  s.Add(5)
  t.Log("Set:", s)
  t.Log("Size:", s.Size())
}

func TestSetContains(t *testing.T) {
  s := New()
  s.Add(1)
  s.Add("hello")
  t.Log("Set:", s)
  if x := s.Contains(1); x {
    t.Log("Contains 1:", x)
  } else {
    t.Error("Contains 1:", x)
  }
}

func TestSetRemove(t *testing.T) {
  s := New()
  s.Add(1)
  s.Add(3)
  s.Add(4)
  s.Add(5)
  t.Log("Set:", s)
  s.Remove(3)
  if x := s.Contains(3); x {
    t.Error("Expected no value but returned 3")
  } else {
    t.Log("Remove() function test passed. returns", x)
  }
}

// TestSetSizeType
func TestSetSizeType(t *testing.T) {
  s := New()
  s.Add(1)
  s.Add(3)
  s.Add(4)
  if reflect.TypeOf(s.Size()).String() != "int" {
    t.Errorf("Size() should return an int")
    return
  }

  t.Log("Passed Size() test with", s, " having a size of ", s.Size())
}

// TestSetUnion tests the union function
func TestSetUnion(t *testing.T) {
  x := New()
  y := New()
  x.Add(1)

  y.Add(2)
  y.Add(3)
  y.Add(4)

  z := x.Union(y)
  t.Log("Set x:", x)
  t.Log("Set y:", y)
  t.Log("Set z:", z)
}

// TestSetIntersection tests the intersection function
func TestSetIntersection(t *testing.T) {
  x := New()
  y := New()
  x.Add(1)
  x.Add(4)

  y.Add(2)
  y.Add(3)
  y.Add(4)

  z := x.Intersection(y)
  t.Log("Set x:", x)
  t.Log("Set y:", y)
  t.Log("Set z:", z)
}

// TestSetDifference tests the difference function
func TestSetDifference(t *testing.T) {
  x := New()
  y := New()
  x.Add(1)
  x.Add(4)

  y.Add(2)
  y.Add(3)
  y.Add(4)

  z := x.Difference(y)
  t.Log("Set x:", x)
  t.Log("Set y:", y)
  t.Log("Set z:", z)
}

// TestSetSymmetricDifference tests the symmetric difference function
func TestSetSymmetricDifference(t *testing.T) {
  x := New()
  y := New()
  x.Add(1)
  x.Add(4)

  y.Add(2)
  y.Add(3)
  y.Add(4)

  z := x.SymmetricDifference(y)
  t.Log("Set x:", x)
  t.Log("Set y:", y)
  t.Log("Set z:", z)
}

// TestSetSubset tests the subset function
func TestSetSubset(t *testing.T) {
  x := New()
  y := New()
  x.Add(1)
  x.Add(4)

  y.Add(2)
  y.Add(3)
  y.Add(4)

  if x.IsSubset(y) {
    t.Error("Expected x to not be a subset of y")
  } else {
    t.Log("Passed Subset test with x:", x, " and y:", y)
  }
}
