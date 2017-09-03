package gallery

import (
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path"
	"runtime/debug"

	"golang.org/x/image/vp8"

	"github.com/nfnt/resize"
	"github.com/sirupsen/logrus"
)

var (
	// WorkerQueue channel controller
	WorkerQueue chan chan WorkRequest
	// WorkQueue channel
	WorkQueue = make(chan WorkRequest, 4)
)

// WorkRequest to be exported to json
type WorkRequest struct {
	Src   string
	Thumb string
}

// NewWorker creates, and returns a new Worker object. Its only argument
// is a channel that the worker can add itself to whenever it is done its
// work.
func NewWorker(id int, workerQueue chan chan WorkRequest) Worker {
	// Create, and return the worker.
	worker := Worker{
		ID:          id,
		Work:        make(chan WorkRequest),
		WorkerQueue: workerQueue,
		QuitChan:    make(chan bool)}

	return worker
}

// Worker type
type Worker struct {
	ID          int
	Work        chan WorkRequest
	WorkerQueue chan chan WorkRequest
	QuitChan    chan bool
}

// StartDispatcher creates nworkers and distributes incoming work to them
func StartDispatcher(nworkers int) {
	// First, initialize the channel we are going to but the workers' work channels into.
	WorkerQueue = make(chan chan WorkRequest, nworkers)

	// Now, create all of our workers.
	for i := 0; i < nworkers; i++ {
		logrus.Info("Starting worker ", i+1)
		worker := NewWorker(i+1, WorkerQueue)
		worker.Start()
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

// Start the worker by starting a goroutine, that is an infinite "for-select" loop.
func (w Worker) Start() {
	go func() {
		for {
			// Add ourselves into the worker queue.
			w.WorkerQueue <- w.Work

			select {
			case work := <-w.Work:
				// Receive a work request.
				logrus.Debugf("Worker %d: Received work request, Generating thumbnail: %s", w.ID, work.Thumb)

				file, err := os.Open(ImageDir + work.Src)
				if err != nil {
					logrus.Error(err)
					break
				}

				var source image.Image
				var thumb image.Image
				switch path.Ext(work.Src) {
				case ".jpg", ".jpeg":
					source, err = jpeg.Decode(file)
					break
				case ".png":
					source, err = png.Decode(file)
					break
				case ".gif":
					source, err = gif.Decode(file)
					break
				case ".webm":
					dec := vp8.NewDecoder()
					dec.Init(file, 1024768)
					source, err = dec.DecodeFrame()
				default:
					err = errors.New("Invalid format " + path.Ext(work.Src))
				}

				file.Close()

				if err != nil {
					logrus.Error(err)
					thumb = image.NewRGBA(image.Rect(0, 0, 96, 96))
				} else {
					x := 0
					y := 96
					if bool(source.Bounds().Max.X > source.Bounds().Max.Y) {
						x = 96
						y = 0
					}
					thumb = resize.Resize(uint(x), uint(y), source, resize.Lanczos3)
				}

				out, err := os.Create(ImageDir + "thumbs/" + work.Thumb)
				if err != nil {
					logrus.Error(err)
					break
				}

				png.Encode(out, thumb)
				out.Close()

				debug.FreeOSMemory()
				break

			case <-w.QuitChan:
				// We have been asked to stop.
				logrus.Infof("Worker %d stopping\n", w.ID)
				return
			}
		}
	}()
}

// Stop tells the worker to stop listening for work requests.
// Note that the worker will only stop *after* it has finished its work.
func (w Worker) Stop() {
	go func() {
		w.QuitChan <- true
	}()
}
