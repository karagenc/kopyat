package config

type (
	Backups struct {
		Hooks     BackupHooks     `mapstructure:"hooks"`
		Reminders BackupReminders `mapstructure:"hooks"`
		Run       []*Backup       `mapstructure:"run"`
	}

	BackupHooks struct {
		Pre  []string `mapstructure:"pre"`
		Post []string `mapstructure:"post"`
	}

	BackupReminders struct {
		Pre  []string `mapstructure:"pre"`
		Post []string `mapstructure:"post"`
	}

	Backup struct {
		Name   string  `mapstructure:"name"`
		Restic *Restic `mapstructure:"restic"`
		Ignore bool    `mapstructure:"ignore"`

		Filter BackupFilter `mapstructure:"filter"`
		Warn   BackupWarn   `mapstructure:"warn"`

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
