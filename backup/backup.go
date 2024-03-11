package backup

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/docker/go-units"
	"github.com/tomruk/kopyaship/backup/provider"
	"github.com/tomruk/kopyaship/config"
	"github.com/tomruk/kopyaship/utils"
)

type (
	Backups map[string]*Backup

	Backup struct {
		Name     string
		Provider provider.Provider

		IfOSIs   string
		WarnSize int64

		Paths *Paths
	}
)

func FromConfig(config *config.Config, cacheDir string) (backups Backups, err error) {
	backups = make(Backups)

	for _, backupConfig := range config.Backups.Run {
		if strings.TrimSpace(backupConfig.Name) == "" {
			return nil, fmt.Errorf("no name given to the backup configuration")
		}
		backup, skip, err := fromConfig(backupConfig, cacheDir)
		if skip {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("backup configuration `%s`: %v", backupConfig.Name, err)
		}
		backups[backupConfig.Name] = backup
	}
	return
}

func fromConfig(backupConfig *config.Backup, cacheDir string) (backup *Backup, skip bool, err error) {
	backup = &Backup{
		Name:     backupConfig.Name,
		Provider: provider.NewRestic(backupConfig.Restic.Repo, backupConfig.Restic.ExtraArgs),
		IfOSIs:   backupConfig.Filter.IfOSIs,
	}

	if backup.skipOS() {
		fmt.Printf("Skipping backup %s, due to unmatched OS: %s\n", backup.Name, backup.IfOSIs)
		skip = true
		return
	}

	if backupConfig.Warn.Size != "" {
		backup.WarnSize, err = units.FromHumanSize(backupConfig.Warn.Size)
		if err != nil {
			return nil, false, err
		}
	}

	backup.Paths = &Paths{
		backup:   backup,
		cacheDir: cacheDir,
		Base:     backupConfig.Base,
		Paths:    backupConfig.Paths,
	}
	err = backup.Paths.check()
	if err != nil {
		return nil, false, err
	}
	return
}

func (b *Backup) skipOS() bool {
	var (
		osNameShort = runtime.GOOS
		ifOSIs      = strings.ToLower(b.IfOSIs)
	)
	switch {
	case ifOSIs != "" && osNameShort != ifOSIs:
		fallthrough
	case osNameShort == "darwin" && ifOSIs == "macos" || ifOSIs == "mac" || ifOSIs == "osx":
		return true
	}
	return false
}

func (backup *Backup) Do() error {
	err := backup.Paths.generateIfile()
	defer os.Remove(backup.Paths.ifilePath())
	if err != nil {
		return err
	}

	err = backup.Provider.BackupWithIfile(backup.Paths.ifilePath())
	if err != nil {
		return err
	}

	if backup.WarnSize != 0 {
		size, err := utils.DirSize(backup.Provider.TargetLocation())
		if err != nil {
			return err
		}
		if size > backup.WarnSize {
			humanSize := units.BytesSize(float64(backup.WarnSize))
			fmt.Printf("\nWARNING: Size of the backup %s has become larger than %s\n\n", backup.Name, humanSize)
		}
	}
	return nil
}
