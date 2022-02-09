package container

import "testing"

var (
	c = New()
)

func TestContainer(t *testing.T) {
	// initialize a new container
	t.Log("New() -> ", c)
}

// TestContainerAdd - adds to the list
func TestContainerAdd(t *testing.T) {
	c.Add(7)
	t.Log("Add(7) -> ", c)
	c.Add(8)
	t.Log("Add(8) -> ", c)
}
