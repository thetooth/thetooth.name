package home

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/thetooth/thetooth.name/gallery"
)

// Handler type
type Handler struct{}

// Pagination type
type Pagination struct {
	Index int
	End   int
	Size  int
}

// Page type
type Page struct {
	Title      string
	Debug      string
	Gallery    []gallery.Image
	Pagination Pagination
}

// ListPages generates pagination
func (p Pagination) ListPages() string {
	buff := ""
	for i := 1; i < p.End; i++ {
		buff = fmt.Sprintf("%s<a href=\"?offset=%d\">%d</a>", buff, i, i)
	}
	return buff
}

// Satisfy http.Handler interface
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// display page
		h.Get(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

// Get Handler
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {

	p := Page{Title: "thetooth.name"}

	itemsPerPage := 140
	images := gallery.Images.Load().([]gallery.Image)

	// Get offset from query
	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil || offset < 1 {
		offset = 1
	}

	// FFL length
	length := len(images)

	// Slice index A
	index := (offset - 1) * itemsPerPage
	if index > length {
		http.Error(w, http.StatusText(http.StatusTeapot), http.StatusTeapot)
		return
	}

	// Slice index B
	end := min(index+itemsPerPage, length)

	p.Gallery = images[index:end]
	p.Pagination.Index = offset
	p.Pagination.End = (length / itemsPerPage) + 1
	p.Pagination.Size = length

	tmpl, _ := ioutil.ReadFile("template.html")
	t, err := template.New("test").Parse(string(tmpl))
	if err != nil {
		logrus.Error(err)
	}
	t.Execute(w, p)
}

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}
