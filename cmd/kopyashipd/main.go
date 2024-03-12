package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gofrs/flock"
	"github.com/kirsle/configdir"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	_config "github.com/tomruk/kopyaship/config"
)

var (
	cacheDir string
	config   *_config.Config
	v        *viper.Viper
	lock     *flock.Flock
)

func main() {
	pflag.StringP("config", "c", "", "Configuration file")
	pflag.Parse()

	systemWide := initConfig()
	initCache(systemWide)
	initLock()

	// To test lockfile
	// n := 5
	// fmt.Printf("Sleeping for %d seconds\n", n)
	// time.Sleep(time.Second * time.Duration(n))
}

func initConfig() (systemWide bool) {
	var (
		configFile, _ = pflag.CommandLine.GetString("config")
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
			cacheDir = configdir.LocalCache()
		}
	}
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		err = os.MkdirAll(cacheDir, 0755)
		if err != nil {
			exit(fmt.Errorf("could not create the cache directory: %v", err))
		}
	}
}

func initLock() {
	user, err := user.Current()
	if err != nil {
		exit(err)
	}
	lockFile := filepath.Join(cacheDir, "kopyashipd_"+user.Username+".lock")
	lock = flock.New(lockFile)

	go func() {
		time.Sleep(2 * time.Second)
		if !lock.Locked() {
			fmt.Printf("The lockfile %s is being used by another instance. Waiting.", lockFile)
		}
	}()
	err = lock.Lock()
	if err != nil {
		exit(err)
	}
}

func exit(err error) {
	if lock != nil {
		lock.Unlock()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
