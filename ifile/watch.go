package ifile

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/tomruk/kopyaship/utils"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type (
	WatchJob struct {
		log     utils.Logger
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
		ScanPath string   `json:"scanPath"`
		Ifile    string   `json:"ifile"`
		Errors   []string `json:"errors"`
		Mode     string   `json:"mode"`
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

func NewWatchJob(log utils.Logger, scanPath, ifile string, mode Mode) *WatchJob {
	j := &WatchJob{
		log:      log,
		status:   atomic.Int32{},
		stopped:  make(chan struct{}),
		errs:     make(chan error, 5),
		scanPath: scanPath,
		ifile:    ifile,
		mode:     mode,
	}
	j.walk = func() error {
		i, err := New(j.ifile, j.mode, true, j.log)
		if err != nil {
			return err
		}
		defer i.Close()
		return i.Walk(j.scanPath)
	}
	j.status.Store(int32(WatchJobStatusWillRun))
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
		ScanPath: j.scanPath,
		Ifile:    j.ifile,
		Errors:   errs,
		Mode:     titleCaser.String(j.mode.String()),
	}
}

func (j *WatchJob) Run() error {
	err := j.walk()
	if err != nil {
		j.logError(err)
		j.fail()
		return err
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
				return err
			}
			continue
		}

		j.status.Store(int32(WatchJobStatusRunning))

		for {
			select {
			case path := <-c:
				j.log.Debugf("event received. path: %s\n", path)
				err := j.walk()
				if err != nil {
					j.logError(err)
					j.sleepBeforeRetry(1)
					if time.Since(firstAttempt).Seconds() >= failAfter {
						j.fail()
						return err
					}
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
	j.log.Infoln("retry in %d second(s)", seconds)
}

func (j *WatchJob) fail() { j.status.Store(int32(WatchJobStatusFailed)) }

func (j *WatchJob) Shutdown() error {
	close(j.stopped)
	j.status.Store(int32(WatchJobStatusStopped))
	return nil
}

func watch(path string) (*fsnotify.Watcher, <-chan string, error) {
	watcher, err := fsnotify.NewWatcher()
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
				if event.Has(fsnotify.Create) {
					if st, err := os.Stat(event.Name); err == nil {
						if st.IsDir() {
							watcher.Add(event.Name)
						}
					}
					c <- event.Name
				} else if event.Has(fsnotify.Write) {
					base := filepath.Base(event.Name)
					if base == gitignore || base == csignore {
						c <- event.Name
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

	err = filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			err := watcher.Add(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return watcher, c, nil
}
