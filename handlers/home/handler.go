package home

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	"github.com/thetooth/thetooth.name/gallery"
)

// Handler type
type Handler struct {
	tmpl *template.Template
}

// Page type
type Page struct {
	Title      string
	Debug      string
	Gallery    []gallery.Image
	Pagination Pagination
}

// NewHandler for home page
func NewHandler() (*Handler, error) {
	t, err := template.New("test").ParseFiles("template.html")
	if err != nil {
		return nil, errors.Wrap(err, "Could not load template.html")
	}
	return &Handler{tmpl: t}, nil
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

	// Render page
	if err := h.tmpl.ExecuteTemplate(w, "template.html", p); err != nil {
		logrus.Error(err)
	}
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
