package backup

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/docker/go-units"
	"github.com/tomruk/kopyaship/backup/provider"
	"github.com/tomruk/kopyaship/config"
	"github.com/tomruk/kopyaship/utils"
)

type (
	Backups map[string]*Backup

	Backup struct {
		isDaemon bool

		Name     string
		Provider provider.Provider

		GenerateIfile bool
		IfOSIs        string
		WarnSize      int64

		Paths *paths
	}
)

func FromConfig(ctx context.Context, config *config.Config, cacheDir string, log utils.Logger, isDaemon bool) (backups Backups, err error) {
	backups = make(Backups)

	for _, backupConfig := range config.Backups.Run {
		if strings.TrimSpace(backupConfig.Name) == "" {
			return nil, fmt.Errorf("no name given to the backup configuration")
		}
		backup, skip, err := fromConfig(ctx, backupConfig, cacheDir, log, isDaemon)
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

func fromConfig(ctx context.Context, backupConfig *config.Backup, cacheDir string, log utils.Logger, isDaemon bool) (backup *Backup, skip bool, err error) {
	backup = &Backup{
		isDaemon:      isDaemon,
		Name:          backupConfig.Name,
		Provider:      provider.NewRestic(ctx, backupConfig.Restic.Repo, backupConfig.Restic.ExtraArgs, backupConfig.Restic.Password, backupConfig.Restic.Sudo, log),
		GenerateIfile: backupConfig.IfileGeneration,
		IfOSIs:        backupConfig.Filter.IfOSIs,
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

	backup.Paths = &paths{
		log:      log,
		cacheDir: cacheDir,
		backup:   backup,
		base:     backupConfig.Base,
		paths:    backupConfig.Paths,
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
		return true
	case osNameShort == "darwin" && ifOSIs == "macos" || ifOSIs == "mac" || ifOSIs == "osx":
		return true
	}
	return false
}

func (backup *Backup) Do() error {
	if !backup.GenerateIfile {
		_, isRestic := backup.Provider.(*provider.Restic)
		if !backup.isDaemon && isRestic && !backup.Provider.PasswordIsSet() {
			fmt.Printf("Enter password for the repository: %s: ", backup.Provider.TargetLocation())
			password, err := terminal.ReadPassword(syscall.Stdin)
			fmt.Println()
			if err != nil {
				return err
			}
			os.Setenv("RESTIC_PASSWORD", string(password))
			defer os.Unsetenv("RESTIC_PASSWORD")
		}

		paths := backup.Paths.Paths()
		for _, path := range paths {
			fmt.Printf("Backing up: %s\n", path)
			err := backup.Provider.Backup(path)
			if err != nil {
				return err
			}
		}
	} else {
		err := backup.Paths.generateIfile()
		//defer os.Remove(backup.Paths.ifilePath())
		if err != nil {
			return err
		}

		err = backup.Provider.BackupWithIfile(backup.Paths.ifilePath())
		if err != nil {
			return err
		}
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
