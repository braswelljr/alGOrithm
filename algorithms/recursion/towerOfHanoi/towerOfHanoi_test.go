package towerOfHanoi

import (
	"testing"
)

// TowerOfHanoiTest ...
func TestTowerOfHanoi(t *testing.T) {
	t.Run("TestTowerOfHanoi", func(t *testing.T) {
		t.Log("TowerOfHanoi")
		n := 3
		TowerOfHanoi(n)
	})
}
