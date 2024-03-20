package ifile

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type (
	WatchJob struct {
		log     *zap.Logger
		logS    *zap.SugaredLogger
		status  atomic.Int32
		stopped chan struct{}
		errs    chan error

		scanPath string
		ifile    string
		mode     Mode

		walk func() error
	}

	WatchJobStatus int32

	WatchJobInfo struct {
		Ifile  string   `json:"ifile"`
		Errors []string `json:"errors"`
		Mode   string   `json:"mode"`
	}
)

const (
	WatchJobStatusWillRun WatchJobStatus = iota
	WatchJobStatusRunning
	WatchJobStatusFailed
	WatchJobStatusStopped
)

const failAfter = 20 // 20 seconds

func (s WatchJobStatus) String() string {
	switch s {
	case WatchJobStatusWillRun:
		return "will run"
	case WatchJobStatusRunning:
		return "running"
	case WatchJobStatusFailed:
		return "failed"
	case WatchJobStatusStopped:
		return "stopped"
	default:
		return "<invalid status>"
	}
}

func NewWatchJob(ifile string, mode Mode, runPreHooks, runPostHooks func() error, log *zap.Logger) *WatchJob {
	if runPreHooks == nil {
		runPreHooks = func() error { return nil }
	}
	if runPostHooks == nil {
		runPostHooks = func() error { return nil }
	}

	j := &WatchJob{
		log:      log,
		logS:     log.Sugar(),
		status:   atomic.Int32{},
		stopped:  make(chan struct{}),
		errs:     make(chan error, 5),
		scanPath: filepath.Dir(ifile),
		ifile:    ifile,
		mode:     mode,
	}
	j.status.Store(int32(WatchJobStatusWillRun))

	j.walk = func() error {
		err := runPreHooks()
		if err != nil {
			j.logS.Errorf("One of the prehooks has failed: %v", err)
		}
		i, err := New(j.ifile, j.mode, true, j.log)
		if err != nil {
			return err
		}
		defer i.Close()
		walkErr := i.Walk(j.scanPath)
		err = runPostHooks()
		if err != nil {
			j.logS.Errorf("One of the posthooks has failed: %v", err)
		}
		return walkErr
	}
	return j
}

func (j *WatchJob) ScanPath() string { return j.scanPath }

func (j *WatchJob) Ifile() string { return j.ifile }

func (j *WatchJob) Status() WatchJobStatus { return WatchJobStatus(j.status.Load()) }

var titleCaser = cases.Title(language.AmericanEnglish)

func (j *WatchJob) Info() *WatchJobInfo {
	errs := make([]string, 0, len(j.errs))
	for len(j.errs) > 0 {
		select {
		case err := <-j.errs:
			errs = append(errs, err.Error())
		default:
		}
	}

	return &WatchJobInfo{
		Ifile:  j.ifile,
		Errors: errs,
		Mode:   titleCaser.String(j.mode.String()),
	}
}

func (j *WatchJob) Run() error {
	err := j.walk()
	if err != nil {
		j.logError(err)
		j.fail()
		return err
	}

	last := time.Now()

outer:
	for {
		watcher, eventChan, err := watch(j.scanPath)
		if err != nil {
			j.logError(err)
			j.sleepBeforeRetry(1)
			// Time since first attempt or last successful walk
			if time.Since(last).Seconds() >= failAfter {
				j.fail()
				return err
			}
			continue
		}

		j.status.Store(int32(WatchJobStatusRunning))

		for {
			select {
			case path := <-eventChan:
				j.logS.Debugf("event received. path: %s", path)
				err := j.walk()
				if err != nil {
					j.logError(err)
					j.sleepBeforeRetry(1)
					// Time since first attempt or last successful walk
					if time.Since(last).Seconds() >= failAfter {
						j.fail()
						return err
					}
					watcher.Close()
					continue outer
				}
				last = time.Now()
			case err, ok := <-watcher.Errors:
				if ok {
					j.logError(err)
					j.sleepBeforeRetry(1)
					// Time since first attempt or last successful walk
					if time.Since(last).Seconds() >= failAfter {
						j.fail()
						return err
					}
					watcher.Close()
					continue outer
				}
			case <-j.stopped:
				watcher.Close()
				return nil
			}
		}
	}
}

func (j *WatchJob) logError(err error) {
	if err != nil {
		err = fmt.Errorf("watch: %v", err)
		j.log.Error(err.Error())
		select {
		case j.errs <- err:
		default:
		}
	}
}

func (j *WatchJob) sleepBeforeRetry(seconds time.Duration) {
	time.Sleep(seconds * time.Second)
	j.logS.Info("retry in %d second(s)", seconds)
}

func (j *WatchJob) fail() { j.status.Store(int32(WatchJobStatusFailed)) }

func (j *WatchJob) Shutdown() error {
	close(j.stopped)
	j.status.Store(int32(WatchJobStatusStopped))
	return nil
}

func watch(path string) (watcher *fsnotify.Watcher, eventChan chan string, err error) {
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return
	}

	eventChan = make(chan string, 1)
	go func() {
		for {
			event, ok := <-watcher.Events
			if !ok {
				return
			}
			if event.Has(fsnotify.Create) {
				if st, err := os.Stat(event.Name); err == nil {
					if st.IsDir() {
						watcher.Add(event.Name)
					}
				}
				eventChan <- event.Name
			} else if event.Has(fsnotify.Write) {
				base := filepath.Base(event.Name)
				if base == gitignore || base == csignore {
					eventChan <- event.Name
				}
			}
		}
	}()

	err = filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			err := watcher.Add(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return
}
