package provider

type Provider interface {
	Init() error
	TargetLocation() string
	Backup(path string) error
	BackupWithIfile(ifile string) error
	PasswordIsSet() bool
}
