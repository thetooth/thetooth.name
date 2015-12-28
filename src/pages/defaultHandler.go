package pages

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
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strconv"
	"text/template"
	"worker"
)

// Handler type
type Handler struct{}

// GalleryT type
type GalleryT struct {
	Image    string
	ImageSrc string
	Valid    bool
}

// GalleryInfoT type
type GalleryInfoT struct {
	Size     int
	Index    int
	IndexMax int
}

// Page type
type Page struct {
	Title       string
	Debug       string
	Gallery     []GalleryT
	GalleryInfo GalleryInfoT
}

type byModTime []os.FileInfo

func (f byModTime) Len() int           { return len(f) }
func (f byModTime) Less(i, j int) bool { return f[i].ModTime().Unix() > f[j].ModTime().Unix() }
func (f byModTime) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

// ListPages generates pagination
func (g GalleryInfoT) ListPages() string {
	buff := ""
	for i := 0; i < g.IndexMax; i++ {
		buff = fmt.Sprintf("%s<a href=\"?offset=%d\">%d</a>", buff, i, i+1)
	}
	return buff
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

var itemsPerPage = 140

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

	var p = new(Page)
	p.Title = "thetooth.name"

	d, err := os.Open(worker.ImageDir)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	fi, err := d.Readdir(-1)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	sort.Sort(byModTime(fi))

	var filteredFileList []os.FileInfo
	for _, v := range fi {
		if !v.IsDir() {
			switch path.Ext(v.Name()) {
			case ".jpg", ".jpeg", ".png", ".gif":
				filteredFileList = append(filteredFileList, v)
				p.GalleryInfo.Size++
				break
			}
		}
	}

	rawIndex, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	maxIndex := int(math.Ceil((float64(len(filteredFileList)) / float64(itemsPerPage))))
	absIndex := min(rawIndex, maxIndex-1) * itemsPerPage

	pageList := filteredFileList[absIndex:(absIndex + min(itemsPerPage, len(filteredFileList[absIndex:])))]

	p.GalleryInfo.Index = rawIndex + 1
	p.GalleryInfo.IndexMax = maxIndex
	p.Gallery = make([]GalleryT, len(pageList))

	for i, v := range pageList {
		createThumbnail(p, i, v)
	}

	tmpl, _ := ioutil.ReadFile("template.html")
	t, err := template.New("test").Parse(string(tmpl))
	if err != nil {
		log.Printf("[%s] ERROR", err)
	}
	t.Execute(w, p)
}

func createThumbnail(p *Page, i int, v os.FileInfo) {
	resizeName := v.Name()[0:(len(v.Name())-len(path.Ext(v.Name())))] + ".png"

	_, err := os.Stat(worker.ImageDir + "thumbs/" + resizeName)
	if os.IsNotExist(err) {
		p.Gallery[i] = GalleryT{"noimage.png", url.QueryEscape(v.Name()), true}
		work := worker.WorkRequest{
			Src:   v.Name(),
			Thumb: resizeName,
		}
		worker.WorkQueue <- work
	} else {
		p.Gallery[i] = GalleryT{url.QueryEscape(resizeName), url.QueryEscape(v.Name()), true}
	}
}
