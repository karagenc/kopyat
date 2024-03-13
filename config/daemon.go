package config

type (
	Daemon struct {
		Log          string       `mapstructure:"log"`
		Notification Notification `mapstructure:"notification"`
		API          API          `mapstructure:"api"`
	}

	Notification struct {
		Enabled bool `mapstructure:"enabled"`
	}

	API struct {
		Enabled bool   `mapstructure:"enabled"`
		Listen  string `mapstructure:"listen"`
		Cert    string `mapstructure:"cert"`
		Key     string `mapstructure:"key"`
	}
)
