package pagination

import (
	"testing"
)

func TestPagination(t *testing.T) {
	// TODO: Write test
	page := New(1, 7, 100, 0)

	if page.Page() != 1 {
		t.Errorf("Page() should return 1")
		return
	}

	if page.PerPage() != 7 {
		t.Errorf("PerPage() should return 7")
		return
	}

	if page.Total() != 100 {
		t.Errorf("Total() should return 100")
		return
	}

	if page.Pages() != 15 {
		t.Errorf("Pages() should return 15")
		return
	} else {
		t.Log("Passed Pages() test", page.Pages())
	}
}
