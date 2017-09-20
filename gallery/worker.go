package gallery

import (
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path"

	"github.com/nfnt/resize"
	"github.com/sirupsen/logrus"
)

var (
	// WorkerQueue channel controller
	WorkerQueue chan chan Image
	// WorkQueue channel
	WorkQueue chan Image
	// WorkerStop quit channel
	WorkerStop chan bool

	workers []Worker
)

// WorkRequest to be exported to json
type WorkRequest struct {
	Src   string
	Thumb string
}

// NewWorker creates, and returns a new Worker object. Its only argument
// is a channel that the worker can add itself to whenever it is done its
// work.
func NewWorker(id int, workerQueue chan chan Image) Worker {
	// Create, and return the worker.
	worker := Worker{
		ID:          id,
		Work:        make(chan Image),
		WorkerQueue: workerQueue,
		QuitChan:    make(chan bool)}

	return worker
}

// Worker type
type Worker struct {
	ID          int
	Work        chan Image
	WorkerQueue chan chan Image
	QuitChan    chan bool
}

// StartDispatcher creates nworkers and distributes incoming work to them
func StartDispatcher(nworkers int) {
	// First, initialize the channel we are going to but the workers' work channels into.
	WorkerQueue = make(chan chan Image, nworkers)

	//
	WorkQueue = make(chan Image, nworkers*nworkers)

	// Now, create all of our workers.
	workers = make([]Worker, nworkers)
	for i := 0; i < nworkers; i++ {
		logrus.Info("Starting worker ", i+1)
		workers[i] = NewWorker(i+1, WorkerQueue)
		workers[i].Start()
	}

	go func() {
		for {
			select {
			case work := <-WorkQueue:
				logrus.Debug("Received work requeust")
				go func() {
					worker := <-WorkerQueue

					logrus.Debug("Dispatching work request")
					worker <- work
				}()
			}
		}
	}()
}

// StopDispatcher blocks until all workers have finished
// and queues are closed
func StopDispatcher() {
	logrus.Debug("Received quit")
	for _, worker := range workers {
		logrus.Debug("Stopping worker ", worker.ID)
		worker.Stop()
	}
	if WorkQueue != nil {
		close(WorkQueue)
		close(WorkerQueue)
	}
}

// Start the worker by starting a goroutine, that is an infinite "for-select" loop.
func (w *Worker) Start() {
	go func() {
		var source image.Image
		var thumb image.Image
		for {
			// Add ourselves into the worker queue.
			if w.WorkerQueue != nil {
				w.WorkerQueue <- w.Work
			}

			select {
			case work := <-w.Work:
				// Receive a work request.
				logrus.Debugf("Worker %d: Received work request, Generating thumbnail: %s", w.ID, work.Thumb)

				file, err := os.Open(ImageDir + work.Src)
				if err != nil {
					logrus.Error(err)
					break
				}

				switch path.Ext(work.Src) {
				case ".jpg", ".jpeg":
					if err = sizeCheck(jpeg.DecodeConfig, file); err != nil {
						break
					}
					source, err = jpeg.Decode(file)
				case ".png":
					if err = sizeCheck(png.DecodeConfig, file); err != nil {
						break
					}
					source, err = png.Decode(file)
				case ".gif":
					if err = sizeCheck(gif.DecodeConfig, file); err != nil {
						break
					}
					source, err = gif.Decode(file)
				default:
					err = errors.New("invalid format " + path.Ext(work.Src))
				}

				file.Close()

				if err != nil {
					logrus.Warnf("Worker %d: %v", w.ID, err)
					break
				} else {
					x := 0
					y := 96
					if bool(source.Bounds().Max.X > source.Bounds().Max.Y) {
						x = 96
						y = 0
					}
					thumb = resize.Resize(uint(x), uint(y), source, resize.NearestNeighbor)
					source = nil
				}

				out, err := os.Create(ImageDir + "thumbs/" + work.Thumb)
				if err != nil {
					logrus.Errorf("Worker %d: %v", w.ID, err)
					break
				}

				// Encode the thumbnail
				if err := png.Encode(out, thumb); err != nil {
					logrus.Warnf("Worker %d: %v", w.ID, err)
					break
				}

				// Update gallery entry
				work.Valid = true
				Images.Set(work.ID, work)

				// Cleanup
				thumb = nil
				out.Close()
			case <-w.QuitChan:
				// We have been asked to stop.
				logrus.Infof("Worker %d: stopped", w.ID)
				return
			}
		}
	}()
}

// Stop tells the worker to stop listening for work requests.
// Note that the worker will only stop *after* it has finished its work.
func (w *Worker) Stop() {
	w.QuitChan <- true
	close(w.QuitChan)
}

// Check image dimensions and rewind reader to the start
func sizeCheck(f func(r io.Reader) (image.Config, error), file *os.File) error {
	cfg, err := f(file)
	if err != nil {
		return err
	}
	if cfg.Width > 6000 || cfg.Height > 6000 {
		return fmt.Errorf("image %s is too large", file.Name())
	}
	_, err = file.Seek(0, 0)
	return err
}
