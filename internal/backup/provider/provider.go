package provider

type Provider interface {
	Init() error
	TargetPath() string
	Backup(path string) error
	BackupWithIfile(ifile string) error
	PasswordIsSet() bool
}
