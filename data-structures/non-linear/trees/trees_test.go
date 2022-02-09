package trees

import "testing"

// Test Tree
func TestTree(t *testing.T) {
  tree := New()

  tree.Insert(5)
  tree.Insert(5)
  tree.Insert(6)

  t.Log(tree.Root)
}
