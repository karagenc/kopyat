package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/kirsle/configdir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_config "github.com/tomruk/kopyaship/config"
)

var (
	cacheDir string
	config   *_config.Config
	v        *viper.Viper
	log      = newLogger()

	rootCmd = &cobra.Command{Use: "kopyaship"}
)

func main() { rootCmd.Execute() }

func init() {
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
		os.Exit(1)
	}
	os.Exit(0)
}
