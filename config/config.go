package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

type (
	Config struct {
		Daemon          Daemon            `mapstructure:"daemon"`
		Env             map[string]string `mapstructure:"env"`
		Scripts         Scripts           `mapstructure:"scripts"`
		IfileGeneration IfileGeneration   `mapstructure:"ifile_generation"`
		Backups         Backups           `mapstructure:"backups"`
	}

	IfileGeneration struct {
		Hooks Hooks                 `mapstructure:"hooks"`
		Run   []*IfileGenerationRun `mapstructure:"run"`
	}

	IfileGenerationRun struct {
		ScanPath string `mapstructure:"scan_path"`
		Ifile    string `mapstructure:"ifile"`
		Mode     string `mapstructure:"mode"`
		Hooks    Hooks  `mapstructure:"hooks"`
	}

	Scripts struct {
		Location string `mapstructure:"location"`
	}

	Hooks struct {
		Pre  []string `mapstructure:"pre"`
		Post []string `mapstructure:"post"`
	}

	Reminders struct {
		Pre  []string `mapstructure:"pre"`
		Post []string `mapstructure:"post"`
	}
)

func Read(configFile string) (config *Config, v *viper.Viper, systemWide bool, err error) {
	v = viper.New()
	if configFile != "" {
		v.SetConfigFile(configFile)
	} else if configFile = os.Getenv("KOPYASHIP_CONFIG"); configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		v.SetConfigName("kopyaship")
		v.SetConfigType("yml")
		v.AddConfigPath(".")

		if runtime.GOOS != "windows" {
			if os.Getenv("$XDG_CONFIG_HOME") != "" {
				v.AddConfigPath("$XDG_CONFIG_HOME/kopyaship")
			} else {
				v.AddConfigPath("$HOME/.config/kopyaship")
			}
			v.AddConfigPath("$HOME/kopyaship")
			v.AddConfigPath("$HOME/.kopyaship")
			v.AddConfigPath("/etc")
		} else {
			v.AddConfigPath("$USERPROFILE/kopyaship")
			v.AddConfigPath("$USERPROFILE/.kopyaship")
			v.AddConfigPath("$PROGRAMDATA/kopyaship")
		}
	}

	err = v.ReadInConfig()
	if err != nil {
		return
	}
	config = new(Config)
	err = v.Unmarshal(config)
	if err != nil {
		return
	}
	configFile, err = filepath.Abs(v.ConfigFileUsed())
	if err != nil {
		return
	}
	os.Setenv("KOPYASHIP_CONFIG", configFile)
	if strings.HasPrefix(configFile, "/etc") || (runtime.GOOS == "windows" && strings.HasPrefix(configFile, os.Getenv("PROGRAMDATA"))) {
		systemWide = true
	}
	return
}

func (c *Config) PlaceEnvironmentVariables() {
	replace := func(r *string) { *r = os.ExpandEnv(*r) }

	for key, value := range c.Env {
		replace(&value)
		os.Setenv(key, value)
	}

	replace(&c.Daemon.Log)
	replace(&c.Daemon.API.Listen)
	replace(&c.Daemon.API.Cert)
	replace(&c.Daemon.API.Key)
	replace(&c.Scripts.Location)

	for i := range c.IfileGeneration.Hooks.Pre {
		replace(&c.IfileGeneration.Hooks.Pre[i])
	}
	for i := range c.IfileGeneration.Hooks.Post {
		replace(&c.IfileGeneration.Hooks.Post[i])
	}
	for i := range c.IfileGeneration.Run {
		replace(&c.IfileGeneration.Run[i].Ifile)
		replace(&c.IfileGeneration.Run[i].ScanPath)
		for j := range c.IfileGeneration.Run[i].Hooks.Pre {
			replace(&c.IfileGeneration.Run[i].Hooks.Pre[j])
		}
		for j := range c.IfileGeneration.Run[i].Hooks.Post {
			replace(&c.IfileGeneration.Run[i].Hooks.Post[j])
		}
	}

	for i := range c.Backups.Hooks.Pre {
		replace(&c.Backups.Hooks.Pre[i])
	}
	for i := range c.Backups.Hooks.Post {
		replace(&c.Backups.Hooks.Post[i])
	}
	for i := range c.Backups.Run {
		replace(&c.Backups.Run[i].Restic.Repo)
		replace(&c.Backups.Run[i].Restic.ExtraArgs)
		for j := range c.Backups.Run[i].Hooks.Pre {
			replace(&c.Backups.Run[i].Hooks.Pre[j])
		}
		for j := range c.Backups.Run[i].Hooks.Post {
			replace(&c.Backups.Run[i].Hooks.Post[j])
		}
		replace(&c.Backups.Run[i].Base)
		for j := range c.Backups.Run[i].Paths {
			replace(&c.Backups.Run[i].Paths[j])
		}
	}
}

func (c *Config) Check() error {
	if c.Daemon.API.Enabled {
		if c.Daemon.API.Listen != "ipc" {
			u, err := url.Parse(c.Daemon.API.Listen)
			if err != nil {
				return err
			} else if u.Path != "/" && u.Path != "" {
				return fmt.Errorf("custom path in URL is not supported. remove '%s' from config", u.Path)
			}
		}
	}

	for _, run := range c.Backups.Run {
		if run.Restic == nil {
			return fmt.Errorf("configuration: field `restic` cannot be empty")
		}
		if run.Base != "" && !filepath.IsAbs(run.Base) {
			return fmt.Errorf("backup base path `%s` is not absolute. to avoid confusion, backup base path must be absolute.", run.Base)
		}
		for _, path := range run.Paths {
			if path == "" {
				return fmt.Errorf("empty backup path. remove it or set it to a file/directory in configuration file.")
			}
			path = filepath.Join(run.Base, path)
			if !filepath.IsAbs(path) {
				return fmt.Errorf("backup path `%s` is not absolute. to prevent confusion, ensure clarity by either setting the base path or setting paths to absolute paths.", path)
			}
		}
	}
	return nil
}

func (c *Config) CheckDaemon() error {
	if c.Daemon.API.Enabled {
		if c.Daemon.API.Listen != "ipc" {
			u, err := url.Parse(c.Daemon.API.Listen)
			if err != nil {
				return err
			} else if u.Path != "/" && u.Path != "" {
				return fmt.Errorf("custom path in URL is not supported. remove '%s' from config", u.Path)
			}
		}
	}

	for _, run := range c.IfileGeneration.Run {
		if run.Ifile == "" {
			return fmt.Errorf("empty ifile path. remove it or set it to a file in configuration file.")
		}
		if !filepath.IsAbs(run.Ifile) {
			return fmt.Errorf("ifile path `%s` is not absolute. to avoid confusion, it must be absolute.", run.Ifile)
		}
		if run.ScanPath == "" {
			return fmt.Errorf("empty scan_path. remove it or set it to a directory in configuration file.")
		}
		if !filepath.IsAbs(run.ScanPath) {
			return fmt.Errorf("scan_path `%s` is not absolute. to avoid confusion, it must be absolute.", run.ScanPath)
		}
	}
	return nil
}
