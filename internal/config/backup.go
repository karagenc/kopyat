package config

type (
	Backups struct {
		Run []*BackupRun `mapstructure:"run"`
	}

	BackupRun struct {
		Name   string  `mapstructure:"name"`
		Restic *Restic `mapstructure:"restic"`

		UseIfile bool `mapstructure:"use_ifile"`

		Hooks     Hooks     `mapstructure:"hooks"`
		Reminders Reminders `mapstructure:"reminders"`

		Base  string   `mapstructure:"base"`
		Paths []string `mapstructure:"paths"`
	}
)
