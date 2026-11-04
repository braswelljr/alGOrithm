package circuits

import (
	"slices"
	"testing"
)

// TestOutputs checks that a circuit names exactly its Output gates, in order,
// and its inputs apart from them.
func TestOutputs(t *testing.T) {
	c := Circuit{Gates: []Gate{
		{Name: "A", Kind: Input},
		{Name: "B", Kind: Input},
		{Name: "n", Kind: Nand, In: []string{"A", "B"}},
		{Name: "Y", Kind: Output, In: []string{"n"}},
		{Name: "Z", Kind: Output, In: []string{"A"}},
	}}
	if got := c.Outputs(); !slices.Equal(got, []string{"Y", "Z"}) {
		t.Errorf("Outputs() = %v, want [Y Z]", got)
	}
	if got := c.Inputs(); !slices.Equal(got, []string{"A", "B"}) {
		t.Errorf("Inputs() = %v, want [A B]", got)
	}
	if got := (Circuit{}).Outputs(); len(got) != 0 {
		t.Errorf("an empty circuit has outputs %v", got)
	}
}
