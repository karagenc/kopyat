package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/gofrs/flock"
	"github.com/karagenc/kopyat/internal/ifile"
	"github.com/karagenc/kopyat/internal/utils"
	"github.com/kardianos/service"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var serviceCmd = &cobra.Command{
	Use: "service",
	Run: func(cmd *cobra.Command, args []string) {
		config := &service.Config{
			Name:        "kopyat",
			DisplayName: "Kopyat service",
		}

		svc := &svc{}
		s, err := service.New(svc, config)
		if err != nil {
			errPrintln(err)
			exit(exitServiceFail)
		}
		log, err := s.Logger(nil)
		if err != nil {
			errPrintln(err)
			exit(exitServiceFail)
		}
		err = s.Run()
		if err != nil {
			log.Error(err)
			errPrintln(err)
			exit(exitServiceFail)
		}
	},
}

var serviceReloadCmd = &cobra.Command{
	Use: "reload",
	Run: func(cmd *cobra.Command, args []string) {
		hc, err := newHTTPClient()
		if err != nil {
			errPrintln(err)
			exit(exitErrAny)
		}
		resp, err := hc.Get("/service/reload")
		if err != nil {
			errPrintln(err)
			exit(exitErrAny)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			errMsg, err := io.ReadAll(resp.Body)
			if err != nil {
				errPrintln(err)
				exit(exitErrAny)
				return
			}
			errPrintln(fmt.Errorf("status: %d, error message: %s", resp.StatusCode, errMsg))
			exit(exitErrAny)
			return
		}
		utils.Success.Println("Successful")
	},
}

type svc struct {
	service   service.Service
	startOnce sync.Once
	stopOnce  sync.Once
	errs      []error
	errsMu    sync.Mutex
	lock      *flock.Flock
	log       *zap.Logger

	watchJobs []*ifile.WatchJob
	jobsMu    sync.Mutex

	e *echo.Echo
	s *http.Server
}

func (s *svc) Start(sv service.Service) (err error) {
	s.startOnce.Do(func() {
		s.service = sv

		err = initEverything()
		if err != nil {
			return
		}
		err = config.CheckService()
		if err != nil {
			return
		}
		err = s.initLock()
		if err != nil {
			return
		}
		s.log, err = s.newLogger(false)
		if err != nil {
			return
		}

		if config.Service.API.Enabled {
			var listen func() error
			s.e, s.s, listen, err = s.newAPIServer()
			if err != nil {
				return
			}

			go func() {
				err := listen()
				if err != nil {
					s.appendErr(fmt.Errorf("api: %v", err))
					err := sv.Stop()
					if err != nil {
						s.log.Error(err.Error())
					}
				}
			}()
		}

		var jobs []*ifile.WatchJob
		jobs, err = s.initWatchJobs()
		if err != nil {
			return
		}
		for _, j := range jobs {
			go func(j *ifile.WatchJob) {
				err := j.Run()
				if err != nil {
					s.appendErr(fmt.Errorf("watch job: %v", err))
					err := sv.Stop()
					if err != nil {
						s.log.Error(err.Error())
					}
				}
			}(j)
		}

	})
	return
}

var lockFile = func() string {
	var lockDir string
	switch runtime.GOOS {
	case "windows":
		lockDir = os.Getenv("PROGRAMDATA")
		if lockDir == "" {
			lockDir = "C:/ProgramData/kopyat"
		}
	case "plan9":
		lockDir = "/lib/kopyat"
	default: // unix
		lockDir = "/var/lock"
	}
	return filepath.Join(lockDir, "kopyat_service.lock")
}()

func (s *svc) initLock() error {
	// Ensure the lockfile directory exists.
	err := os.MkdirAll(filepath.Dir(lockFile), 0755)
	if err != nil {
		return err
	}
	s.lock = flock.New(lockFile)
	go func() {
		time.Sleep(2 * time.Second)
		if !s.lock.Locked() {
			fmt.Printf("The lockfile %s is being used by another instance. Waiting.", lockFile)
		}
	}()
	err = s.lock.Lock()
	if err != nil {
		return err
	}
	return nil
}

func (s *svc) appendErr(err error) {
	if err != nil {
		s.errsMu.Lock()
		s.errs = append(s.errs, err)
		s.errsMu.Unlock()
	}
}

func (s *svc) reload(c echo.Context) error {
	return nil
}

func (s *svc) Stop(sv service.Service) (err error) {
	s.stopOnce.Do(func() {
		if s.lock != nil {
			s.lock.Unlock()
		}
		if config != nil && s.e != nil {
			if config.Service.API.Listen == "ipc" {
				socketPath := filepath.Join(stateDir, apiSocketFileName)
				os.Remove(socketPath)
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			s.e.Shutdown(ctx)
			cancel()
		}

		s.jobsMu.Lock()
		defer s.jobsMu.Unlock()
		for _, job := range s.watchJobs {
			jobErr := job.Shutdown()
			if err == nil && jobErr != nil {
				err = jobErr
			}
		}

		onExit()

		s.errsMu.Lock()
		defer s.errsMu.Unlock()
		for _, e := range s.errs {
			err = errors.Join(err, e)
		}
	})
	return
}

// Check whether serviceCmd is going to run.
// (And no sub cmd is going to run.)
func willRunAsService() bool {
	var (
		hasServiceCmd    bool
		hasServiceSubCmd bool
	)
	for _, arg := range os.Args[1:] {
		if arg[0] == '-' {
			continue
		} else if arg == "service" {
			hasServiceCmd = true
		} else if hasServiceCmd {
			hasServiceSubCmd = true
		}
	}
	return hasServiceCmd && !hasServiceSubCmd
}
