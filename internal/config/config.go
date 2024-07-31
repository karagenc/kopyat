package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

type (
	Config struct {
		Backups         Backups           `mapstructure:"backups"`
		IfileGeneration IfileGeneration   `mapstructure:"ifile_generation"`
		Env             map[string]string `mapstructure:"env"`
		Service         Service           `mapstructure:"service"`
	}

	IfileGeneration struct {
		Run []*IfileGenerationRun `mapstructure:"run"`
	}

	IfileGenerationRun struct {
		Ifile string `mapstructure:"ifile"`
		Type  string `mapstructure:"type"`
		Hooks Hooks  `mapstructure:"hooks"`
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

func DirsLocal() []string {
	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	dirs := []string{
		filepath.Join(home, "kopyat"),
		filepath.Join(home, ".kopyat"),
	}

	// Avoid duplicate paths. If $XDG_CONFIG_HOME is the same as ~/.config,
	// userConfigDir will be a duplicate of configHome + "/kopyat"
	xdgConfigHome := filepath.Clean(os.Getenv("XDG_CONFIG_HOME"))
	configHome := filepath.Join(home, ".config")
	if xdgConfigHome != configHome {
		dirs = append(dirs, filepath.Join(configHome, "kopyat"))
	}
	return dirs
}

func Read(configFile, userConfigDir, systemConfigDir string) (
	config *Config,
	v *viper.Viper,
	systemWide bool,
	err error,
) {
	if runtime.GOOS == "windows" {
		home, err := homedir.Dir()
		if err != nil {
			return nil, nil, false, err
		}
		err = os.Setenv("HOME", home)
		if err != nil {
			return nil, nil, false, err
		}
	}

	v = viper.New()
	if configFile != "" {
		v.SetConfigFile(configFile)
	} else if configFileEnv := os.Getenv("KOPYAT_CONFIG"); configFileEnv != "" {
		v.SetConfigFile(configFileEnv)
	} else {
		v.SetConfigName("kopyat")
		v.SetConfigType("yml")

		v.AddConfigPath(".")
		// Prioritize local (user) dirs
		for _, dir := range DirsLocal() {
			v.AddConfigPath(dir)
		}
		v.AddConfigPath(userConfigDir)
		v.AddConfigPath(systemConfigDir)
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
	err = os.Setenv("KOPYAT_CONFIG", configFile)
	if err != nil {
		return
	}

	// If system config directory is used, set systemWide to true
	// to indicate we're using system-wide directories.
	if filepath.Dir(configFile) == systemConfigDir {
		systemWide = true
	}
	return
}

func (c *Config) PlaceEnvironmentVariables() error {
	replace := func(r *string) {
		*r = os.ExpandEnv(*r)
		*r = filepath.ToSlash(*r)
	}

	for key, value := range c.Env {
		key = strings.ToUpper(key)
		c.Env[key] = value
		replace(&value)
		err := os.Setenv(key, value)
		if err != nil {
			return err
		}
	}

	replace(&c.Service.Log)
	replace(&c.Service.API.Listen)
	replace(&c.Service.API.Cert)
	replace(&c.Service.API.Key)

	for i := range c.IfileGeneration.Run {
		if c.IfileGeneration.Run[i] == nil {
			continue
		}
		replace(&c.IfileGeneration.Run[i].Ifile)
		for j := range c.IfileGeneration.Run[i].Hooks.Pre {
			replace(&c.IfileGeneration.Run[i].Hooks.Pre[j])
		}
		for j := range c.IfileGeneration.Run[i].Hooks.Post {
			replace(&c.IfileGeneration.Run[i].Hooks.Post[j])
		}
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
	return nil
}

func (c *Config) CheckNonService() error {
	if c.Service.API.Enabled {
		if c.Service.API.Listen != "ipc" {
			u, err := url.Parse(c.Service.API.Listen)
			if err != nil {
				return err
			} else if u.Path != "/" && u.Path != "" {
				return fmt.Errorf("custom path in URL is not supported. remove '%s' from config", u.Path)
			}
		}
	}

	for _, run := range c.Backups.Run {
		if run.Restic == nil {
			return fmt.Errorf("config: field `restic` cannot be empty")
		}
		if run.Base != "" {
			if !filepath.IsAbs(run.Base) {
				return fmt.Errorf("backup base path `%s` is not absolute. to avoid confusion, backup base path must be absolute", run.Base)
			}
			run.Base = filepath.ToSlash(run.Base)
		}
		for i, path := range run.Paths {
			if path == "" {
				return fmt.Errorf("empty backup path. remove it or set it to a file/directory in config file")
			}
			path = filepath.Join(run.Base, path)
			if !filepath.IsAbs(path) {
				return fmt.Errorf("backup path `%s` is not absolute. to prevent confusion, ensure clarity by either setting the base path or setting paths to absolute paths", path)
			}
			path = filepath.ToSlash(path)
			run.Paths[i] = path
		}
	}
	return nil
}

func (c *Config) CheckService() error {
	if c.Service.API.Enabled {
		if c.Service.API.Listen != "ipc" {
			u, err := url.Parse(c.Service.API.Listen)
			if err != nil {
				return err
			} else if u.Path != "/" && u.Path != "" {
				return fmt.Errorf("custom path in URL is not supported. remove '%s' from config", u.Path)
			}
		}
		if c.Service.API.BasicAuth.Enabled {
			if c.Service.API.BasicAuth.Username == "" {
				return fmt.Errorf("empty API username. assign a username or disable API in config")
			}
			if c.Service.API.BasicAuth.Password == "" {
				return fmt.Errorf("empty API password. assign a password or disable API in config")
			}
		}
	}

	for _, run := range c.IfileGeneration.Run {
		if run.Ifile == "" {
			return fmt.Errorf("empty ifile path. remove it or set it to a file in config file")
		}
		if !filepath.IsAbs(run.Ifile) {
			return fmt.Errorf("ifile path `%s` is not absolute. to avoid confusion, it must be absolute", run.Ifile)
		}
	}
	return nil
}
