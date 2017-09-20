package home

import (
	"fmt"
	"html/template"
)

// Pagination type
type Pagination struct {
	Index int
	End   int
	Size  int
}

// ListPages generates pagination
func (p Pagination) ListPages() template.HTML {
	buff := ""
	for i := 1; i < p.End; i++ {
		buff = fmt.Sprintf("%s<a href=\"?offset=%d\">%d</a>", buff, i, i)
	}
	return template.HTML(buff)
}
