package design

import "fmt"

// PaginationConfig is a list being paged while new items arrive.
type PaginationConfig struct {
	Rows       int // items in the list, newest first
	PageSize   int // items a page
	Pages      int // pages the user reads
	NewPerPage int // items added at the top between two page loads
}

// Pagination pages through a newest-first list (a feed, search results)
// while new items keep arriving at the top, both ways. Offset pagination
// (LIMIT 20 OFFSET 40) is simple and can jump to any page, but the database
// reads and throws away all the offset rows each time, and every insert
// shifts the list, so the next page repeats items the user already saw.
// Cursor (keyset) pagination (WHERE id < last_seen LIMIT 20) reads only the
// page and never repeats or skips, but cannot jump to page 50. It reports
// the duplicates seen and the rows read for each.
//
// Time: O(Pages * (offset + PageSize)). Space: O(Rows).
func Pagination(c PaginationConfig) (Report, error) {
	if err := check(c.Rows >= 1 && c.Rows <= 1_000_000 && c.PageSize >= 1 && c.Pages >= 1 && c.Pages <= 1000 && c.NewPerPage >= 0, "1 to 1,000,000 rows, positive page size, 1 to 1,000 pages"); err != nil {
		return Report{}, err
	}
	var r Report
	t := r.table("offset: rows read", "offset: repeats", "cursor: rows read", "cursor: repeats")
	newest := c.Rows // ids 1..newest, shown newest first
	seenOffset := map[int]bool{}
	cursor := newest + 1
	var offRead, offDup, curRead int
	for p := 0; p < c.Pages; p++ {
		if p > 0 {
			newest += c.NewPerPage
		}
		offset := p * c.PageSize
		readO, dupO := 0, 0
		for i := 0; i < c.PageSize; i++ {
			id := newest - offset - i
			if id < 1 {
				break
			}
			readO++
			if seenOffset[id] {
				dupO++
			}
			seenOffset[id] = true
		}
		readO += min(offset, newest)
		readC := 0
		for i := 0; i < c.PageSize && cursor-1 >= 1; i++ {
			cursor--
			readC++
		}
		offRead += readO
		offDup += dupO
		curRead += readC
		t.add(fmt.Sprintf("page %d", p+1), readO, dupO, readC, 0)
	}
	r.say("%d pages of %d while %d new items arrived before each page", c.Pages, c.PageSize, c.NewPerPage)
	r.say("offset: %d rows read, %d items shown twice; cursor: %d rows read, none twice", offRead, offDup, curRead)
	r.put("offset rows read", float64(offRead))
	r.put("offset repeats", float64(offDup))
	r.put("cursor rows read", float64(curRead))
	return r, nil
}
