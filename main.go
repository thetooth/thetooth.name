package main

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
	"net/http"
	"runtime"

	"github.com/namsral/flag"

	fsnotify "gopkg.in/fsnotify.v1"

	"github.com/pkg/errors"
	"github.com/pkg/profile"
	"github.com/sirupsen/logrus"
	"github.com/thetooth/thetooth.name/gallery"
	"github.com/thetooth/thetooth.name/handlers/home"
)

func init() {
	flag.StringVar(&gallery.ImageDir, "image_dir", "images/", "Where you keep images")
	flag.Parse()
}

func main() {
	defer profile.Start(profile.ProfilePath("."), profile.MemProfile).Stop()

	runtime.GOMAXPROCS(1)

	logrus.SetLevel(logrus.DebugLevel)
	logrus.Info("thetooth.name starting up...")

	// Start workers
	gallery.StartDispatcher(2)

	// Start file system watcher for images
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logrus.Fatal(err)
	}

	// Start making thumbnails right away
	gallery.Update()

	// Process events
	go func() {
		for {
			select {
			case <-watcher.Events:
				logrus.Info("FS")
				if err := gallery.Update(); err != nil {
					logrus.Error(err)
				}
				logrus.Info("Updated gallery")
			case err := <-watcher.Errors:
				logrus.Error(err)
			}
		}
	}()

	err = watcher.Add(gallery.ImageDir)
	if err != nil {
		logrus.Fatal(errors.Wrapf(err, "fsnotify \"%s\"", gallery.ImageDir))
	}

	// Renderer
	h := &home.Handler{}

	// Create mux
	mux := http.NewServeMux()
	mux.Handle("/", h)
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(gallery.ImageDir))))
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	mux.Handle("/tmp/", http.StripPrefix("/tmp/", http.FileServer(http.Dir("tmp/"))))

	logrus.Info("Listening")
	http.ListenAndServe("0.0.0.0:9000", mux)
}
