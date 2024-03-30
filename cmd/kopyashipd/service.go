package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/flock"
	"github.com/kardianos/service"
	"github.com/labstack/echo/v4"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tomruk/finddirs-go"
	"github.com/tomruk/kopyaship/config"
	_config "github.com/tomruk/kopyaship/config"
	"github.com/tomruk/kopyaship/ifile"
	"github.com/tomruk/kopyaship/utils"
	"go.uber.org/zap"
)

type svice struct {
	service service.Service

	watchJobs []*ifile.WatchJob
	jobsMu    sync.Mutex

	e *echo.Echo
	s *http.Server

	stateDir string
	cacheDir string
	config   *_config.Config
	v        *viper.Viper
	lock     *flock.Flock
	log      *zap.Logger

	once    sync.Once
	errChan <-chan error

	exitHandlers   []func()
	exitHandlersMu sync.Mutex
}

func (v *svice) Start(s service.Service) (err error) {
	v.once.Do(func() {
		errChan := make(chan error, 1)
		v.errChan = errChan
		v.service = s

		pflag.StringP("config", "c", "", "Configuration file")
		pflag.Parse()

		var (
			userAppDirs   *finddirs.AppDirs
			systemAppDirs *finddirs.AppDirs
			systemWide    bool
		)

		userAppDirs, err = finddirs.RetrieveAppDirs(false, &utils.FindDirsConfig)
		if err != nil {
			return
		}
		systemAppDirs, err = finddirs.RetrieveAppDirs(true, &utils.FindDirsConfig)
		if err != nil {
			return
		}

		systemWide, err = v.initConfig(userAppDirs.ConfigDir, systemAppDirs.ConfigDir)
		if err != nil {
			return
		}
		err = v.initStateDir(systemWide, userAppDirs.StateDir, systemAppDirs.StateDir)
		if err != nil {
			return
		}
		err = v.initCacheDir(systemWide, userAppDirs.CacheDir, systemAppDirs.CacheDir)
		if err != nil {
			return
		}
		err = v.config.PlaceEnvironmentVariables()
		if err != nil {
			return
		}
		err = v.config.CheckDaemon()
		if err != nil {
			return
		}
		err = v.initLock()
		if err != nil {
			return
		}
		v.log, err = v.newLogger(false)
		if err != nil {
			return
		}

		if v.config.Daemon.API.Enabled {
			var listen func() error
			v.e, v.s, listen, err = v.newAPIServer()
			if err != nil {
				return
			}

			go func() {
				err := listen()
				if err != nil {
					errChan <- fmt.Errorf("api: %v", err)
					err := s.Stop()
					if err != nil {
						v.log.Error(err.Error())
					}
				}
			}()
		}

		var jobs []*ifile.WatchJob
		jobs, err = v.initWatchJobsFromConfig()
		if err != nil {
			return
		}
		for _, j := range jobs {
			go func(j *ifile.WatchJob) {
				err := j.Run()
				if err != nil {
					errChan <- fmt.Errorf("api: %v", err)
					err := s.Stop()
					if err != nil {
						v.log.Error(err.Error())
					}
				}
			}(j)
		}
	})
	return
}

func (v *svice) initConfig(userConfigDir, systemConfigDir string) (systemWide bool, err error) {
	configFileArg, _ := pflag.CommandLine.GetString("config")
	v.config, v.v, systemWide, err = config.Read(configFileArg, userConfigDir, systemConfigDir)
	if err != nil {
		return
	}
	err = os.Chdir(filepath.Dir(v.v.ConfigFileUsed()))
	return
}

func (v *svice) initStateDir(systemWide bool, userStateDir, systemStateDir string) (err error) {
	v.stateDir = os.Getenv("KOPYASHIP_STATE_DIR")
	if v.stateDir == "" {
		if systemWide {
			v.stateDir = systemStateDir
		} else {
			v.stateDir = userStateDir
		}
		err = os.Setenv("KOPYASHIP_STATE_DIR", v.stateDir)
		if err != nil {
			return err
		}
	}

	return os.MkdirAll(v.stateDir, 0755)
}

func (v *svice) initCacheDir(systemWide bool, userCacheDir, systemCacheDir string) (err error) {
	v.cacheDir = os.Getenv("KOPYASHIP_CACHE")
	if v.cacheDir == "" {
		if systemWide {
			v.cacheDir = systemCacheDir
		} else {
			v.cacheDir = userCacheDir
		}
		err = os.Setenv("KOPYASHIP_CACHE", v.cacheDir)
		if err != nil {
			return err
		}
	}

	return os.MkdirAll(v.cacheDir, 0755)
}

func (v *svice) initLock() error {
	user, err := user.Current()
	if err != nil {
		return err
	}
	username := user.Username
	username = strings.ReplaceAll(username, "\\", "_")
	lockFile := filepath.Join(v.stateDir, "kopyashipd_"+username+".lock")
	v.lock = flock.New(lockFile)

	go func() {
		time.Sleep(2 * time.Second)
		if !v.lock.Locked() {
			fmt.Printf("The lockfile %s is being used by another instance. Waiting.", lockFile)
		}
	}()
	err = v.lock.Lock()
	if err != nil {
		return err
	}
	return nil
}

func (v *svice) Stop(s service.Service) (err error) {
	select {
	case err = <-v.errChan:
	default:
	}
	if v.lock != nil {
		v.lock.Unlock()
	}
	if v.config != nil && v.e != nil {
		if v.config.Daemon.API.Listen == "ipc" {
			socketPath := filepath.Join(v.stateDir, "api.socket")
			defer os.Remove(socketPath)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		v.e.Shutdown(ctx)
		cancel()
	}

	v.jobsMu.Lock()
	defer v.jobsMu.Unlock()
	for _, job := range v.watchJobs {
		jobErr := job.Shutdown()
		if err == nil && jobErr != nil {
			err = jobErr
		}
	}

	v.exitHandlersMu.Lock()
	defer v.exitHandlersMu.Unlock()
	for _, f := range v.exitHandlers {
		f()
	}
	return
}

func (v *svice) addExitHandler(f func()) {
	v.exitHandlersMu.Lock()
	v.exitHandlers = append(v.exitHandlers, f)
	v.exitHandlersMu.Unlock()
}
