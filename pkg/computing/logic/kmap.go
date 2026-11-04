package logic

import (
	"fmt"
	"slices"
)

// Cell values of a Karnaugh map.
const (
	Zero     = 0
	One      = 1
	DontCare = 2
)

// KMap is a Karnaugh map: the truth table laid out so that neighboring cells,
// wrapping around the edges, differ in one variable (rows and columns in Gray
// code order). Cells holds Zero, One or DontCare row by row; Minterm[i] is the
// minterm in cell i; Groups are the cells each term of the minimal cover
// spans, with Terms the matching products.
type KMap struct {
	Vars                 []string
	Rows, Cols           int
	RowLabels, ColLabels []string
	Cells, Minterm       []int
	Groups               [][]int
	Terms                []string
	SOP                  string
}

// gray lists the k-bit Gray code: 0, 1 for one bit; 00, 01, 11, 10 for two.
func gray(k int) []int {
	if k == 1 {
		return []int{0, 1}
	}
	return []int{0, 1, 3, 2}
}

func bitsLabel(v, k int) string {
	return fmt.Sprintf("%0*b", k, v)
}

// Karnaugh builds the map of the function of vars (2 to 4 of them) that is
// true on minterms and does not matter on dontCares, and draws the groups of
// its minimal sum of products (found with QuineMcCluskey, so the groups are
// the largest rectangles of 1, 2, 4 or 8 cells, wrapping around the edges,
// that the minimal cover needs).
//
// Time: that of QuineMcCluskey for at most 4 variables, so O(1) in practice.
// Space: O(16).
func Karnaugh(vars []string, minterms, dontCares []int) (KMap, error) {
	n := len(vars)
	if n < 2 || n > 4 {
		return KMap{}, fmt.Errorf("%w: a Karnaugh map takes 2 to 4 variables", ErrInput)
	}
	res, err := QuineMcCluskey(n, minterms, dontCares)
	if err != nil {
		return KMap{}, err
	}
	rowBits := n / 2
	colBits := n - rowBits
	rg, cg := gray(rowBits), gray(colBits)
	k := KMap{Vars: vars, Rows: len(rg), Cols: len(cg), SOP: res.SOP(vars)}
	for _, r := range rg {
		k.RowLabels = append(k.RowLabels, bitsLabel(r, rowBits))
	}
	for _, c := range cg {
		k.ColLabels = append(k.ColLabels, bitsLabel(c, colBits))
	}
	for _, r := range rg {
		for _, c := range cg {
			m := r<<colBits | c
			v := Zero
			switch {
			case slices.Contains(minterms, m):
				v = One
			case slices.Contains(dontCares, m):
				v = DontCare
			}
			k.Cells = append(k.Cells, v)
			k.Minterm = append(k.Minterm, m)
		}
	}
	for _, p := range res.Cover {
		var cells []int
		for i, m := range k.Minterm {
			if p.Covers(m) {
				cells = append(cells, i)
			}
		}
		k.Groups = append(k.Groups, cells)
		k.Terms = append(k.Terms, p.Term(vars))
	}
	return k, nil
}
