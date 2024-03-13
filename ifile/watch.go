package ifile

import (
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

type (
	WatchJob struct {
		log     *zap.Logger
		status  atomic.Int32
		stopped chan struct{}

		scanPath string
		ifile    string
	}

	Status int32
)

const (
	StatusWillRun Status = iota
	StatusRunning
	StatusFailed
	StatusStopped
)

const failAfter = 20 // 20 seconds

func NewWatchJob(log *zap.Logger, ScanPath, Ifile string) *WatchJob {
	j := &WatchJob{
		log:      log,
		status:   atomic.Int32{},
		stopped:  make(chan struct{}),
		scanPath: ScanPath,
		ifile:    Ifile,
	}
	j.status.Store(int32(StatusWillRun))
	return j
}

func (j *WatchJob) ScanPath() string { return j.scanPath }

func (j *WatchJob) Ifile() string { return j.ifile }

func (j *WatchJob) Status() Status { return Status(j.status.Load()) }

func (j *WatchJob) Start() error {
	go func() {
		err := j.walk()
		if err != nil {
			j.logError(err)
			j.fail()
			return
		}

		firstAttempt := time.Now()

	outer:
		for {
			watcher, c, err := watch(j.scanPath)
			if err != nil {
				j.logError(err)
				j.sleepBeforeRetry(1)
				if time.Since(firstAttempt).Seconds() >= failAfter {
					j.fail()
				}
				continue
			}

			j.status.Store(int32(StatusRunning))

			for {
				select {
				case <-c:
					err := j.walk()
					if err != nil {
						j.logError(err)
						j.sleepBeforeRetry(1)
						if time.Since(firstAttempt).Seconds() >= failAfter {
							j.fail()
						}
						continue outer
					}
				case <-j.stopped:
					watcher.Close()
					return
				}
			}
		}
	}()
	return nil
}

func (j *WatchJob) logError(err error) {
	if err != nil {
		err = fmt.Errorf("watch: %v", err)
		j.log.Error(err.Error())
	}
}

func (j *WatchJob) sleepBeforeRetry(seconds time.Duration) {
	time.Sleep(seconds * time.Second)
	j.log.Sugar().Info("retry in %d second(s)", seconds)
}

func (j *WatchJob) fail() { j.status.Store(int32(StatusFailed)) }

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
	close(j.stopped)
	j.status.Store(int32(StatusStopped))
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
