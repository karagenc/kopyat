package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	gochoice "github.com/TwiN/go-choice"
	"github.com/karagenc/finddirs-go"
	_config "github.com/karagenc/kopyat/internal/config"
	"github.com/karagenc/kopyat/internal/utils"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	_ "github.com/karagenc/kopyat/internal/statik"
)

var (
	configDir string
	stateDir  string
	cacheDir  string

	config *_config.Config
	v      *viper.Viper

	debugLog *zap.Logger

	rootCmd = &cobra.Command{Use: "kopyat"}
)

func main() { rootCmd.Execute() }

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(pingCmd)
	rootCmd.AddCommand(watchJobCmd)
	watchJobCmd.AddCommand(watchJobListCmd)
	watchJobCmd.AddCommand(watchJobStopCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(runScript)
	rootCmd.AddCommand(serviceCmd)
	serviceCmd.AddCommand(serviceReloadCmd)
	rootCmd.AddCommand(ifileCmd)
	ifileCmd.AddCommand(ifileGenerateCmd)
	ifileGenerateCmd.AddCommand(ifileGenerateSyncthingCmd)

	rootCmd.PersistentFlags().StringP("config", "c", "", "Config file")
	rootCmd.PersistentFlags().Bool("enable-log", false, "Enable debug logging to stdout")

	cobra.OnInitialize(sync.OnceFunc(func() {
		// If running as service, defer initialization to svc.Start, and
		// don't handle signals manually, as they will be handled by `kardianos/service`.
		if willRunAsService() {
			return
		}

		sigChan := make(chan os.Signal, 2)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			sig := <-sigChan
			switch sig {
			case syscall.SIGINT:
				fallthrough
			case syscall.SIGTERM:
				exit(exitTerm)
			}
		}()

		err := initEverything()
		if err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				exists, extractErr := extractConfigInteractive()
				if !exists {
					utils.Red.Print("Error: ")
					fmt.Printf("%v: in addition to that, example config is not found within executable. consider fetching it from https://github.com/karagenc/kopyat\n", err)
					exit(exitErrAny)
				} else if extractErr != nil {
					errPrintln(extractErr)
					exit(exitErrAny)
				}
				exit(exitSuccess)
			}

			errPrintln(err)
			exit(exitErrAny)
		}
		err = config.CheckNonService()
		if err != nil {
			errPrintln(err)
			exit(exitErrAny)
		}
	}))
}

var initEverything = sync.OnceValue(func() error {
	userAppDirs, err := finddirs.RetrieveAppDirs(false, &utils.FindDirsConfig)
	if err != nil {
		return err
	}
	systemAppDirs, err := finddirs.RetrieveAppDirs(true, &utils.FindDirsConfig)
	if err != nil {
		return err
	}
	systemWide, err := initConfig(userAppDirs.ConfigDir, systemAppDirs.ConfigDir)
	if err != nil {
		return err
	}
	err = initStateDir(systemWide, userAppDirs.StateDir, systemAppDirs.StateDir)
	if err != nil {
		return err
	}
	err = initCacheDir(systemWide, userAppDirs.CacheDir, systemAppDirs.CacheDir)
	if err != nil {
		return err
	}
	err = initLogging()
	if err != nil {
		return err
	}
	return config.PlaceEnvironmentVariables()
})

func initConfig(userConfigDir, systemConfigDir string) (systemWide bool, err error) {
	configFileArg, _ := rootCmd.PersistentFlags().GetString("config")
	config, v, systemWide, err = _config.Read(configFileArg, userConfigDir, systemConfigDir)
	if err != nil {
		return false, err
	}
	configDir = filepath.Dir(v.ConfigFileUsed())
	err = os.Chdir(configDir)
	return
}

func extractConfigInteractive() (exists bool, err error) {
	statikFS, err := fs.New()
	if err != nil {
		return false, err
	}
	exampleFile, err := statikFS.Open("/kopyat_example.yml")
	if err != nil {
		return false, err
	}
	defer exampleFile.Close()
	example, err := io.ReadAll(exampleFile)
	if err != nil {
		return true, err
	}

	userAppDirs, err := finddirs.RetrieveAppDirs(false, &utils.FindDirsConfig)
	if err != nil {
		return true, err
	}
	systemAppDirs, err := finddirs.RetrieveAppDirs(true, &utils.FindDirsConfig)
	if err != nil {
		return true, err
	}

	dirs := _config.DirsLocal()
	dirs = append(dirs, []string{
		userAppDirs.ConfigDir,
		systemAppDirs.ConfigDir,
	}...)
	dir, _, err := gochoice.Pick(
		"Where to store the config file, kopyat.yml?\nPick:",
		dirs,
	)
	if err != nil {
		return true, err
	}

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return true, err
	}
	err = os.WriteFile(filepath.Join(dir, "kopyat.yml"), example, 0644)
	if err != nil {
		return true, err
	}
	utils.Success.Println("Config written. Make sure you've fully read and edited it before running kopyat")
	return true, nil
}

func initStateDir(systemWide bool, userStateDir, systemStateDir string) (err error) {
	stateDir = os.Getenv("KOPYAT_STATE_DIR")
	if stateDir == "" {
		if systemWide {
			stateDir = systemStateDir
		} else {
			stateDir = userStateDir
		}
		err = os.Setenv("KOPYAT_STATE_DIR", stateDir)
		if err != nil {
			return err
		}
	}
	return os.MkdirAll(stateDir, 0755)
}

func initCacheDir(systemWide bool, userCacheDir, systemCacheDir string) (err error) {
	cacheDir = os.Getenv("KOPYAT_CACHE")
	if cacheDir == "" {
		if systemWide {
			cacheDir = systemCacheDir
		} else {
			cacheDir = userCacheDir
		}
		err = os.Setenv("KOPYAT_CACHE", cacheDir)
		if err != nil {
			return err
		}
	}
	return os.MkdirAll(cacheDir, 0755)
}

type exitCode int

const (
	exitSuccess exitCode = iota
	exitErrAny
	exitTerm
	exitServiceFail
)

func errPrintln(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", utils.Red.Sprint("Error:"), err)
	}
}

var (
	exitHandlers   []func()
	exitHandlersMu sync.Mutex
)

func addExitHandler(f func()) {
	exitHandlersMu.Lock()
	exitHandlers = append(exitHandlers, sync.OnceFunc(f))
	exitHandlersMu.Unlock()
}

func onExit() {
	exitHandlersMu.Lock()
	defer exitHandlersMu.Unlock()
	for _, f := range exitHandlers {
		f()
	}
}

func exit(code exitCode) {
	onExit()
	os.Exit(int(code))
}
