package gallery

import (
	"os"
	"path"
	"sort"
	"sync"

	"github.com/sirupsen/logrus"
)

// ImageDir is where you keep your images.
// Needs write permissions on `thumbs` subdirectory
var ImageDir string

// Image type
type Image struct {
	ID    int
	Src   string
	Thumb string
	Valid bool
}

type byModTime []os.FileInfo

func (f byModTime) Len() int           { return len(f) }
func (f byModTime) Less(i, j int) bool { return f[i].ModTime().Unix() > f[j].ModTime().Unix() }
func (f byModTime) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

// Images global, stores a cache of gallery entries and
// if they have valid thumbnails or not
var Images ImageCache

// ImageCache threadsafe cache
type ImageCache struct {
	sync.RWMutex
	imgs []Image
}

// List provides a copy of the cache
func (c *ImageCache) List() []Image {
	c.RLock()
	defer c.RUnlock()

	tmp := make([]Image, len(c.imgs))
	copy(tmp, c.imgs)

	return tmp
}

// Set the value of a gallery entry
func (c *ImageCache) Set(id int, img Image) {
	c.Lock()
	defer c.Unlock()
	c.imgs[id] = img
}

// Update will rescan the directory for any files that don't
// have a thumbnail and attempt to generate one
func (c *ImageCache) Update() error {
	// Stop currently running workers and start a new one
	StopDispatcher()
	StartDispatcher(2)

	c.Lock()
	defer c.Unlock()

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
			case ".jpg", ".jpeg", ".png", ".gif":
				filteredFileList = append(filteredFileList, v)
				break
			}
		}
	}

	// Generate thumbnails
	c.imgs = make([]Image, len(filteredFileList))
	for i, f := range filteredFileList {
		c.imgs[i] = createThumbnail(i, f)
	}

	logrus.Info("Updated gallery")

	return nil
}

func createThumbnail(id int, f os.FileInfo) Image {
	resizeName := f.Name()[:(len(f.Name())-len(path.Ext(f.Name())))] + ".png"

	img := Image{
		ID:    id,
		Src:   f.Name(),
		Thumb: resizeName,
		Valid: true,
	}

	_, err := os.Stat(ImageDir + "thumbs/" + resizeName)
	if os.IsNotExist(err) {
		img.Valid = false
		WorkQueue <- img
	}

	return img
}
