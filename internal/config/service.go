package config

type (
	Service struct {
		Log string `mapstructure:"log"`
		API API    `mapstructure:"api"`
	}

	API struct {
		Enabled   bool      `mapstructure:"enabled"`
		Listen    string    `mapstructure:"listen"`
		Cert      string    `mapstructure:"cert"`
		Key       string    `mapstructure:"key"`
		BasicAuth BasicAuth `mapstructure:"basic_auth"`
	}

	BasicAuth struct {
		Enabled  bool   `mapstructure:"enabled"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	}
)
