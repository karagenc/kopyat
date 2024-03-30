package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tomruk/finddirs-go"
	_config "github.com/tomruk/kopyaship/config"
	"github.com/tomruk/kopyaship/utils"
	"go.uber.org/zap"
)

var (
	stateDir string
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
	rootCmd.PersistentFlags().Bool("enable-log", false, "Enable logging to stdout. (For debugging purposes.)")

	cobra.OnInitialize(func() {
		userAppDirs, err := finddirs.RetrieveAppDirs(false, &utils.FindDirsConfig)
		if err != nil {
			exit(err, nil)
		}
		systemAppDirs, err := finddirs.RetrieveAppDirs(true, &utils.FindDirsConfig)
		if err != nil {
			exit(err, nil)
		}

		systemWide := initConfig(userAppDirs.ConfigDir, systemAppDirs.ConfigDir)
		err = initStateDir(systemWide, userAppDirs.StateDir, systemAppDirs.StateDir)
		if err != nil {
			exit(err, nil)
		}
		err = initCacheDir(systemWide, userAppDirs.CacheDir, systemAppDirs.CacheDir)
		if err != nil {
			exit(err, nil)
		}
		initLogging()
		err = config.PlaceEnvironmentVariables()
		if err != nil {
			exit(err, nil)
		}
		err = config.Check()
		if err != nil {
			exit(err, nil)
		}
	})
}

func initConfig(userConfigDir, systemConfigDir string) (systemWide bool) {
	var (
		configFileArg, _ = rootCmd.PersistentFlags().GetString("config")
		err              error
	)
	config, v, systemWide, err = _config.Read(configFileArg, userConfigDir, systemConfigDir)
	if err != nil {
		exit(err, nil)
	}
	err = os.Chdir(filepath.Dir(v.ConfigFileUsed()))
	if err != nil {
		exit(err, nil)
	}
	return
}

func initStateDir(systemWide bool, userStateDir, systemStateDir string) (err error) {
	stateDir = os.Getenv("KOPYASHIP_STATE_DIR")
	if stateDir == "" {
		if systemWide {
			stateDir = systemStateDir
		} else {
			stateDir = userStateDir
		}
		err = os.Setenv("KOPYASHIP_STATE_DIR", stateDir)
		if err != nil {
			return err
		}
	}

	return os.MkdirAll(stateDir, 0755)
}

func initCacheDir(systemWide bool, userCacheDir, systemCacheDir string) (err error) {
	cacheDir = os.Getenv("KOPYASHIP_CACHE")
	if cacheDir == "" {
		if systemWide {
			cacheDir = systemCacheDir
		} else {
			cacheDir = userCacheDir
		}
		err = os.Setenv("KOPYASHIP_CACHE", cacheDir)
		if err != nil {
			return err
		}
	}

	return os.MkdirAll(cacheDir, 0755)
}

func initLogging() {
	enable, _ := rootCmd.PersistentFlags().GetBool("enable-log")
	if enable {
		var err error
		log, err = utils.NewDebugLogger()
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
		fmt.Fprintf(os.Stderr, "%s\n", utils.Red.Sprintf("Error: %v", err))
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
