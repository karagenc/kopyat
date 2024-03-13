package ifile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

type WatchJob struct {
	log     *zap.Logger
	stopped chan struct{}

	scanPath string
	ifile    string
}

func NewWatchJob(log *zap.Logger, ScanPath, Ifile string) *WatchJob {
	return &WatchJob{
		log:      log,
		stopped:  make(chan struct{}),
		scanPath: ScanPath,
		ifile:    Ifile,
	}
}

func (j *WatchJob) ScanPath() string { return j.scanPath }

func (j *WatchJob) Ifile() string { return j.ifile }

func (j *WatchJob) Start() error {
	go func() {
		err := j.walk()
		if err != nil {
			j.log.Error(err.Error())
			return
		}

		watcher, c, err := watch(j.scanPath)
		if err != nil {
			j.log.Error(err.Error())
			return
		}

		for {
			select {
			case <-c:
				err := j.walk()
				if err != nil {
					j.log.Error(err.Error())
					return
				}
			case <-j.stopped:
				watcher.Close()
				return
			}
		}
	}()
	return nil
}

func (j *WatchJob) walk() error {
	// TODO: Arguments
	i, err := New(j.ifile, Include, true, true)
	if err != nil {
		return err
	}
	defer i.Close()
	return i.Walk(j.scanPath)
}

func (j *WatchJob) Stop() error {
	return nil
}

func watch(path string) (*fsnotify.Watcher, <-chan string, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, nil, err
	}
	err = watcher.Add(path)
	if err != nil {
		return nil, nil, err
	}

	c := make(chan string, 1)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
					base := filepath.Base(event.Name)
					if base == gitignore || base == csignore {

					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Fprintf(os.Stderr, "Event error: %v\n", err)
			}
		}
	}()
	return watcher, c, nil
}
