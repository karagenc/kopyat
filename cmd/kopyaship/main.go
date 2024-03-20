package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/kirsle/configdir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_config "github.com/tomruk/kopyaship/config"
	"github.com/tomruk/kopyaship/utils"
	"go.uber.org/zap"
)

var (
	cacheDir string
	config   *_config.Config
	v        *viper.Viper
	log      *zap.Logger

	rootCmd = &cobra.Command{Use: "kopyaship"}
)

func main() { rootCmd.Execute() }

func init() {
	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		switch sig {
		case syscall.SIGINT:
			fallthrough
		case syscall.SIGTERM:
			code := 2
			exit(nil, &code)
		}
	}()

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(pingCmd)
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(runCmd)
	watchCmd.AddCommand(watchListCmd)

	rootCmd.PersistentFlags().StringP("config", "c", "", "Configuration file")
	rootCmd.PersistentFlags().Bool("enable-log", false, "Enable logging to stdout")

	cobra.OnInitialize(func() {
		systemWide := initConfig()
		initCache(systemWide)
		initLogging()
		config.PlaceEnvironmentVariables()
		err := config.Check()
		if err != nil {
			exit(err, nil)
		}
	})
}

func initConfig() (systemWide bool) {
	var (
		configFile, _ = rootCmd.PersistentFlags().GetString("config")
		err           error
	)
	config, v, systemWide, err = _config.Read(configFile)
	if err != nil {
		exit(err, nil)
	}
	err = os.Chdir(filepath.Dir(v.ConfigFileUsed()))
	if err != nil {
		exit(err, nil)
	}
	return
}

func initCache(systemWide bool) {
	cacheDir = os.Getenv("KOPYASHIP_CACHE")
	if cacheDir == "" {
		if systemWide {
			if !utils.RunningOnWindows {
				cacheDir = "/var/cache/kopyaship"
			} else {
				cacheDir = filepath.Join(os.Getenv("PROGRAMDATA"), "kopyaship", "cache")
			}
		} else {
			cacheDir = filepath.Join(configdir.LocalCache(), "kopyaship")
		}
		os.Setenv("KOPYASHIP_CACHE", cacheDir)
	}
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		err = os.MkdirAll(cacheDir, 0755)
		if err != nil {
			exit(fmt.Errorf("could not create the cache directory: %v", err), nil)
		}
	}
}

func initLogging() {
	enable, _ := rootCmd.PersistentFlags().GetBool("enable-log")
	if enable {
		var err error
		log, err = newLogger()
		if err != nil {
			exit(fmt.Errorf("could not create a new logger: %v", err), nil)
		}
	} else {
		log = zap.NewNop()
	}
}

var (
	exitHandlers   []func()
	exitHandlersMu sync.Mutex
)

func addExitHandler(f func()) {
	exitHandlersMu.Lock()
	exitHandlers = append(exitHandlers, f)
	exitHandlersMu.Unlock()
}

func exit(err error, code *int) {
	exitHandlersMu.Lock()
	defer exitHandlersMu.Unlock()
	for _, f := range exitHandlers {
		f()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		if code != nil {
			os.Exit(*code)
		} else {
			os.Exit(1)
		}
	}
	if code != nil {
		os.Exit(*code)
	} else {
		os.Exit(0)
	}
}
