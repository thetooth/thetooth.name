package main

import (
	"net/http"
	"runtime"
	"time"

	"github.com/namsral/flag"

	fsnotify "gopkg.in/fsnotify.v1"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/thetooth/thetooth.name/gallery"
	"github.com/thetooth/thetooth.name/handlers/home"
)

func init() {
	flag.StringVar(&gallery.ImageDir, "image_dir", "images/", "Where you keep images")
	flag.Parse()
}

func main() {
	//defer profile.Start(profile.ProfilePath("."), profile.MemProfile).Stop()

	runtime.GOMAXPROCS(1)

	logrus.SetLevel(logrus.DebugLevel)
	logrus.Info("thetooth.name starting up...")

	// Start file system watcher for images
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logrus.Fatal(err)
	}

	// Start making thumbnails right away
	gallery.Images.Update()

	// Process events
	go func() {
		for {
			select {
			case <-watcher.Events:
				if err := gallery.Images.Update(); err != nil {
					logrus.Error(err)
				}
			case err := <-watcher.Errors:
				logrus.Warn(err)
			}
		}
	}()

	err = watcher.Add(gallery.ImageDir)
	if err != nil {
		logrus.Fatal(errors.Wrapf(err, "fsnotify \"%s\"", gallery.ImageDir))
	}

	// Renderer
	h, err := home.NewHandler()
	if err != nil {
		logrus.Fatal(err)
	}

	// Create mux
	mux := http.NewServeMux()
	mux.Handle("/", h)
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(gallery.ImageDir))))
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	mux.Handle("/tmp/", http.StripPrefix("/tmp/", http.FileServer(http.Dir("tmp/"))))

	logrus.Info("Listening")
	server := http.Server{
		Addr:    "0.0.0.0:9000",
		Handler: mux,

		ReadHeaderTimeout: time.Second,
		ReadTimeout:       time.Second * 5,
		WriteTimeout:      time.Second * 10,
		IdleTimeout:       time.Second * 120,
	}
	server.ListenAndServe()
}
