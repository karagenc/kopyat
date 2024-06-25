package backup

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/term"

	"github.com/tomruk/kopyaship/internal/backup/provider"
	"github.com/tomruk/kopyaship/internal/config"
	"github.com/tomruk/kopyaship/internal/utils"
)

type (
	Backups map[string]*Backup

	Backup struct {
		asService bool
		log       *zap.Logger
		Config    *config.BackupRun

		Name     string
		Provider provider.Provider
		UseIfile bool

		Paths *paths
	}
)

func FromConfig(
	ctx context.Context,
	configBackups *config.Backups,
	cacheDir string,
	log *zap.Logger,
	asService bool,
	include ...string,
) (backups Backups, err error) {
	backups = make(Backups)

	if len(include) > 0 {
		for _, include := range include {
			found := false
			for _, run := range configBackups.Run {
				if run.Name == include {
					found = true
				}
			}
			if !found {
				return nil, fmt.Errorf("no backup with name: %s", include)
			}
		}
	}

	for _, run := range configBackups.Run {
		if strings.TrimSpace(run.Name) == "" {
			return nil, fmt.Errorf("no name given to the backup config")
		}
		if len(include) > 0 {
			found := false
			for _, include := range include {
				if run.Name == include {
					found = true
				}
			}
			if !found {
				continue
			}
		}

		backup, skip, err := fromConfig(ctx, run, cacheDir, log, asService)
		if skip {
			utils.Warn.Print("Skipping backup: ")
			fmt.Println(run.Name)
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("backup config `%s`: %v", run.Name, err)
		}
		backups[run.Name] = backup
	}
	return
}

func fromConfig(
	ctx context.Context,
	config *config.BackupRun,
	cacheDir string,
	log *zap.Logger,
	asService bool,
) (backup *Backup, skip bool, err error) {
	backup = &Backup{
		asService: asService,
		log:       log,
		Config:    config,
		Name:      config.Name,
		Provider:  provider.NewRestic(ctx, config.Restic.Repo, config.Restic.ExtraArgs, config.Restic.Password, config.Restic.Sudo, log),
		UseIfile:  config.UseIfile,
	}

	backup.Paths = &paths{
		log:      log,
		cacheDir: cacheDir,
		backup:   backup,
		base:     config.Base,
		paths:    config.Paths,
	}
	err = backup.Paths.check()
	if err != nil {
		return nil, false, err
	}
	return
}

func (b *Backup) Do() error {
	if !b.UseIfile {
		paths := b.Paths.Paths()

		if len(paths) > 1 {
			_, isRestic := b.Provider.(*provider.Restic)
			if !b.asService && isRestic && !b.Provider.PasswordIsSet() {
				fmt.Printf("Enter password for the repository %s: ", b.Provider.TargetPath())
				password, err := term.ReadPassword(int(os.Stdin.Fd()))
				fmt.Println()
				if err != nil {
					return err
				}

				err = os.Setenv("RESTIC_PASSWORD", string(password))
				if err != nil {
					return err
				}
				defer os.Unsetenv("RESTIC_PASSWORD")
			}
		}

		for _, path := range paths {
			b.log.Sugar().Infof("Backup: %s", path)
			if !b.asService {
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
	return nil
}
