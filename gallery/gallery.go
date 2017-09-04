package gallery

import (
	"net/url"
	"os"
	"path"
	"sort"
	"sync/atomic"
)

// Images atomic global, stores a type of []Image
var Images atomic.Value

// ImageDir const
var ImageDir = "images/"

// Image type
type Image struct {
	Thumb string
	Src   string
	Valid bool
}

type byModTime []os.FileInfo

func (f byModTime) Len() int           { return len(f) }
func (f byModTime) Less(i, j int) bool { return f[i].ModTime().Unix() > f[j].ModTime().Unix() }
func (f byModTime) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

// Update will rescan the directory for any files that don't
// have a thumb and attempt to generate one
func Update() error {
	d, err := os.Open(ImageDir)
	if err != nil {
		return err
	}
	fi, err := d.Readdir(-1)
	if err != nil {
		return err
	}

	sort.Sort(byModTime(fi))

	var filteredFileList []os.FileInfo
	for _, v := range fi {
		if !v.IsDir() {
			switch path.Ext(v.Name()) {
			case ".jpg", ".jpeg", ".png", ".gif", ".webm":
				filteredFileList = append(filteredFileList, v)
				break
			}
		}
	}

	// Generate thumbnails
	images := make([]Image, len(filteredFileList))
	for i, f := range filteredFileList {
		images[i] = createThumbnail(f)
	}

	// Update the gallery
	Images.Store(images)

	return nil
}

func createThumbnail(f os.FileInfo) Image {
	resizeName := f.Name()[:(len(f.Name())-len(path.Ext(f.Name())))] + ".png"

	_, err := os.Stat(ImageDir + "thumbs/" + resizeName)
	if os.IsNotExist(err) {
		work := WorkRequest{
			Src:   f.Name(),
			Thumb: resizeName,
		}
		WorkQueue <- work
	}

	return Image{url.PathEscape(resizeName), url.PathEscape(f.Name()), true}
}
