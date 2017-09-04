package home

/*
	thetooth.name - Gallery and home page written in Go

	Copyright (c) 2005-2015 Ameoto Systems Inc. All rights reserved.

	Redistribution and use in source and binary forms, with or without
	modification, are permitted provided that the following conditions are met:

	1. Redistributions of source code must retain the above copyright notice, this
	list of conditions and the following disclaimer.
	2. Redistributions in binary form must reproduce the above copyright notice,
	this list of conditions and the following disclaimer in the documentation
	and/or other materials provided with the distribution.

	THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
	ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
	WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
	DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
	ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
	(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
	LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
	ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
	(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
	SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

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
