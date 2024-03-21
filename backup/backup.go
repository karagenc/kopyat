package backup

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/term"

	"github.com/docker/go-units"
	"github.com/tomruk/kopyaship/backup/provider"
	"github.com/tomruk/kopyaship/config"
	"github.com/tomruk/kopyaship/utils"
)

type (
	Backups map[string]*Backup

	Backup struct {
		isDaemon bool
		log      *zap.Logger

		Name     string
		Provider provider.Provider

		GenerateIfile bool
		IfOSIs        string
		WarnSize      int64

		Paths *paths
	}
)

func FromConfig(ctx context.Context, configBackups *config.Backups, cacheDir string, log *zap.Logger, isDaemon bool, include ...string) (backups Backups, err error) {
	backups = make(Backups)

	if len(include) > 0 {
		for _, include := range include {
			found := false
			for _, backupConfig := range configBackups.Run {
				if backupConfig.Name == include {
					found = true
				}
			}
			if !found {
				return nil, fmt.Errorf("no backup with name: %s", include)
			}
		}
	}

	for _, backupConfig := range configBackups.Run {
		if strings.TrimSpace(backupConfig.Name) == "" {
			return nil, fmt.Errorf("no name given to the backup configuration")
		}
		if len(include) > 0 {
			found := false
			for _, include := range include {
				if backupConfig.Name == include {
					found = true
				}
			}
			if !found {
				continue
			}
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

func fromConfig(ctx context.Context, backupConfig *config.Backup, cacheDir string, log *zap.Logger, isDaemon bool) (backup *Backup, skip bool, err error) {
	backup = &Backup{
		isDaemon:      isDaemon,
		log:           log,
		Name:          backupConfig.Name,
		Provider:      provider.NewRestic(ctx, backupConfig.Restic.Repo, backupConfig.Restic.ExtraArgs, backupConfig.Restic.Password, backupConfig.Restic.Sudo, log),
		GenerateIfile: backupConfig.IfileGeneration,
		IfOSIs:        backupConfig.Filter.IfOSIs,
	}

	if backup.skipOS() {
		log.Sugar().Infof("Skipping backup %s, due to unmatched OS: %s", backup.Name, backup.IfOSIs)
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

func (b *Backup) Do() error {
	if !b.GenerateIfile {
		paths := b.Paths.Paths()

		if len(paths) > 1 {
			_, isRestic := b.Provider.(*provider.Restic)
			if !b.isDaemon && isRestic && !b.Provider.PasswordIsSet() {
				fmt.Printf("Enter password for the repository %s: ", b.Provider.TargetLocation())
				password, err := term.ReadPassword(int(os.Stdin.Fd()))
				fmt.Println()
				if err != nil {
					return err
				}
				os.Setenv("RESTIC_PASSWORD", string(password))
				defer os.Unsetenv("RESTIC_PASSWORD")
			}
		}

		for _, path := range paths {
			b.log.Sugar().Infof("Backup: %s", path)
			if !b.isDaemon {
				fmt.Println()
				utils.BgBlue.Printf("Backup: %s", path)
				fmt.Println()
			}
			err := b.Provider.Backup(path)
			if err != nil {
				return err
			}
		}
	} else {
		err := b.Paths.generateIfile()
		defer os.Remove(b.Paths.ifilePath())
		if err != nil {
			return err
		}

		err = b.Provider.BackupWithIfile(b.Paths.ifilePath())
		if err != nil {
			return err
		}
	}

	if b.WarnSize != 0 {
		size, err := utils.DirSize(b.Provider.TargetLocation())
		if err != nil {
			return err
		}
		if size > b.WarnSize {
			humanSize := units.BytesSize(float64(b.WarnSize))
			b.log.Sugar().Warnf("Size of the backup %s has become larger than %s", b.Name, humanSize)
		}
	}
	return nil
}
