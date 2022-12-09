package pagination

import (
	"math"
)

// Pagination represents a pagination.
type Pagination struct {
	page     int
	perPage  int
	total    int
	pages    int
	previous int
	next     int
	offset   int
}

// New creates a new pagination.
// @param page is the current page.
// @param perPage is the number of items per page.
// @param total - is the total number of items.
// @return *Pagination
func New(page, perPage, total, offset int) *Pagination {
	// Create a new pagination.
	p := &Pagination{}

	// Set the perPage.
	p.perPage = perPage

	// Set the total.
	p.total = total

	// Set the pages.
	p.pages = func() int {
		// If the total is less than or equal to perPage, return 1.
		if total <= perPage {
			return 1
		}

		// divide the total by the perPage and get the ceiling.
		if total%perPage > 0 {
			// Return the total / perPage + 1.
			return int(math.Trunc(float64(total)/float64(perPage))) + 1
		}

		return total / perPage
	}()

	// Set the page.
	p.page = func() int {
		// If the page is less than 1, return 1.
		if page < 1 {
			return 1
		}

		// If the page is greater than the pages, return the pages.
		if page > p.pages {
			return p.pages
		}

		return page
	}()

	// Set the previous.
	p.previous = func() int {
		// If the page is less than 1, return the total pages.
		if p.page < 1 {
			return p.pages
		}
		// Return the page - 1
		return p.page - 1
	}()

	// Set the next.
	p.next = func() int {
		// If the page is greater than the pages, return the first page.
		if p.page > p.pages {
			return 1
		}

		// Return the page + 1
		return p.page + 1
	}()

	// Set the offset.
	p.offset = func() int {
		if offset == 0 {
			// offset is where to start from in a new page.
			// If the page is less than 1
			if p.page < 2 {
				return 0
			}

			// Return the current (page * perPage) - 1
			return (p.page * p.perPage) - 1
		}
		return offset
	}()

	return p
}

// Page returns the page.
func (p *Pagination) Page() int {
	return p.page
}

// PerPage returns the perPage.
func (p *Pagination) PerPage() int {
	return p.perPage
}

// Total returns the total.
func (p *Pagination) Total() int {
	return p.total
}

// Pages returns the total number of pages.
func (p *Pagination) Pages() int {
	return p.pages
}

// Previous returns the previous.
func (p *Pagination) Previous() int {
	return p.previous
}

// Next returns the next.
func (p *Pagination) Next() int {
	return p.next
}

// Offset returns the offset.
func (p *Pagination) Offset() int {
	return p.offset
}

// HasPrevious returns true if the pagination has a previous page, false otherwise.
func (p *Pagination) HasPrevious() bool {
	return p.page > 1
}

// HasNext returns true if the pagination has a next page, false otherwise.
func (p *Pagination) HasNext() bool {
	return p.page < p.pages
}

// set the page
func (p *Pagination) SetPage(page int) {
	p.page = page
}

// set the perPage
func (p *Pagination) SetPerPage(perPage int) {
	p.perPage = perPage
}

// set the total
func (p *Pagination) SetTotal(total int) {
	p.total = total
}

// set the pages
func (p *Pagination) SetPages(pages int) {
	p.pages = pages
}

// set the previous
func (p *Pagination) SetPrevious(previous int) {
	p.previous = previous
}

// set the next
func (p *Pagination) SetNext(next int) {
	p.next = next
}

// set the offset
func (p *Pagination) SetOffset(offset int) {
	p.offset = offset
}
