// Command api-pagination computes the page links of a paged REST API
// response (offset, limit, previous and next), as every list endpoint does.
package main

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/math/numbers"
)

func main() {
	const total, perPage = 47, 10
	for _, n := range []int{1, 3, 5} {
		p := numbers.Paginate(n, perPage, total)
		fmt.Printf("GET /items?page=%d -> items %d-%d of %d, pages %d, prev %d, next %d\n",
			p.Number, p.Offset+1, p.End, p.Total, p.Pages, p.Prev, p.Next)
	}
}
