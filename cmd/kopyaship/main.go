package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"github.com/kirsle/configdir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_config "github.com/tomruk/kopyaship/config"
	"github.com/tomruk/kopyaship/utils"
)

var (
	cacheDir string
	config   *_config.Config
	v        *viper.Viper
	log      = utils.NewCLILogger(false)

	rootCmd = &cobra.Command{Use: "kopyaship"}
)

func main() { rootCmd.Execute() }

func init() {
	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			sig := <-sigChan
			switch sig {
			case syscall.SIGINT:
				for {
					r := bufio.NewReader(os.Stdin)
					fmt.Print("Are you sure you want to exit? (y/N): ")
					input, _ := r.ReadString('\n')
					input = strings.TrimSpace(input)

					if strings.EqualFold(input, "y") {
						code := 2
						exit(nil, &code)
					} else if strings.EqualFold(input, "n") {
						break
					} else {
						fmt.Println("Invalid input. y or n expected.")
					}
				}
			case syscall.SIGTERM:
				code := 2
				exit(nil, &code)
			}
		}
	}()

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(pingCmd)
	rootCmd.AddCommand(watchCmd)

	rootCmd.PersistentFlags().StringP("config", "c", "", "Configuration file")

	watchCmd.AddCommand(watchListCmd)

	cobra.OnInitialize(func() {
		systemWide := initConfig()
		initCache(systemWide)
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
			if runtime.GOOS != "windows" {
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

var (
	exitFuncs   []func()
	exitFuncsMu sync.Mutex
)

func addExitHandler(f func()) {
	exitFuncsMu.Lock()
	exitFuncs = append(exitFuncs, f)
	exitFuncsMu.Unlock()
}

func exit(err error, code *int) {
	exitFuncsMu.Lock()
	exitFuncs := exitFuncs
	exitFuncsMu.Unlock()
	for _, f := range exitFuncs {
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
