package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

type (
	Config struct {
		Daemon  Daemon            `mapstructure:"daemon"`
		Env     map[string]string `mapstructure:"env"`
		Scripts Scripts           `mapstructure:"scripts"`
		IfileGeneration
		Backups Backups `mapstructure:"backups"`
	}

	IfileGeneration struct {
		Hooks Hooks                 `mapstructure:"hooks"`
		Run   []*IfileGenerationRun `mapstructure:"run"`
	}

	IfileGenerationRun struct {
		ScanPath  string `mapstructure:"scan_path"`
		Recursive bool   `mapstructure:"recursive"`
		Ifile     string `mapstructure:"ifile"`
		Hooks     Hooks  `mapstructure:"hooks"`
	}

	Daemon struct {
		Log          string `mapstructure:"log"`
		Notification struct {
			Enabled bool `mapstructure:"enabled"`
		} `mapstructure:"notification"`
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
	err = config.Check()
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
	replace := func(r *string) {
		for _, env := range os.Environ() {
			splitted := strings.Split(env, "=")
			key := splitted[0]
			value := splitted[1]
			*r = strings.ReplaceAll(*r, "$"+key, value)
		}
	}

	replace(&c.Daemon.Log)
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
	for _, backup := range c.Backups.Run {
		if backup.Restic == nil {
			return fmt.Errorf("configuration: field `restic` cannot be empty")
		}
	}
	return nil
}
