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
	if strings.HasPrefix(configFile, "/etc") || (runtime.GOOS == "windows" && strings.HasPrefix(configFile, os.Getenv("PROGRAMDATA"))) {
		systemWide = true
	}
	return
}

func (c *Config) Check() error {
	for _, backup := range c.Backups.Run {
		if backup.Restic == nil {
			return fmt.Errorf("configuration: field `restic` cannot be empty")
		}
	}
	return nil
}
