package home

import "fmt"

// Pagination type
type Pagination struct {
	Index int
	End   int
	Size  int
}

// ListPages generates pagination
func (p Pagination) ListPages() string {
	buff := ""
	for i := 1; i < p.End; i++ {
		buff = fmt.Sprintf("%s<a href=\"?offset=%d\">%d</a>", buff, i, i)
	}
	return buff
}
