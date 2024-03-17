package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
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
					fmt.Print("Are you sure (y/N): ")
					input, _ := r.ReadString('\n')
					input = strings.TrimSpace(input)

					if strings.EqualFold(input, "y") {
						utils.Exit(2)
					} else if strings.EqualFold(input, "n") {
						break
					} else {
						fmt.Println("Invalid input. y or n expected.")
					}
				}
			case syscall.SIGTERM:
				utils.Exit(3)
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
			exit(err)
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
		exit(err)
	}
	err = os.Chdir(filepath.Dir(v.ConfigFileUsed()))
	if err != nil {
		exit(err)
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
			exit(fmt.Errorf("could not create the cache directory: %v", err))
		}
	}
}

func exit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		utils.Exit(1)
	}
	utils.Exit(0)
}
