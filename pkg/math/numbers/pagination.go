package numbers

// Page describes one page of a paged list: which page, how many pages there
// are, where the page starts, and its neighbors. Pages count from 1.
type Page struct {
	Number, PerPage, Total, Pages int
	// Offset is the index of the page's first item; End is one past its last.
	Offset, End int
	// Prev and Next are the neighboring page numbers, 0 when there is none.
	Prev, Next int
}

// Paginate returns the page number of a list of total items, perPage a page
// (at least 1). A number below 1 gives page 1 and one past the end gives the
// last page; an empty list has one empty page.
//
// Time: O(1).
func Paginate(number, perPage, total int) Page {
	if perPage < 1 {
		perPage = 1
	}
	if total < 0 {
		total = 0
	}
	pages := max(1, (total+perPage-1)/perPage)
	number = min(max(number, 1), pages)
	p := Page{Number: number, PerPage: perPage, Total: total, Pages: pages}
	p.Offset = (number - 1) * perPage
	p.End = min(p.Offset+perPage, total)
	if number > 1 {
		p.Prev = number - 1
	}
	if number < pages {
		p.Next = number + 1
	}
	return p
}
