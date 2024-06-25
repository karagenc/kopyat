package config

type Restic struct {
	Repo      string `mapstructure:"repo"`
	Sudo      bool   `mapstructure:"sudo"`
	ExtraArgs string `mapstructure:"extra_args"`
	Password  string `mapstructure:"password"`
}
