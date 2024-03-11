package config

type (
	Backups struct {
		Hooks     Hooks     `mapstructure:"hooks"`
		Reminders Reminders `mapstructure:"hooks"`
		Run       []*Backup `mapstructure:"run"`
	}

	Backup struct {
		Name   string  `mapstructure:"name"`
		Restic *Restic `mapstructure:"restic"`

		Filter          BackupFilter `mapstructure:"filter"`
		Warn            BackupWarn   `mapstructure:"warn"`
		IfileGeneration bool         `mapstructure:"ifile_generation"`

		Hooks     Hooks     `mapstructure:"hooks"`
		Reminders Reminders `mapstructure:"reminders"`

		Base  string   `mapstructure:"base"`
		Paths []string `mapstructure:"paths"`
	}

	BackupFilter struct {
		IfOSIs string `mapstructure:"if_os_is"`
	}

	BackupWarn struct {
		Size string `mapstructure:"size"`
	}
)
