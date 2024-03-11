package config

type Restic struct {
	Repo      string `mapstructure:"repo"`
	ExtraArgs string `mapstructure:"extra_args"`
}
