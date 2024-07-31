package main

import (
	"context"
	"fmt"

	"github.com/karagenc/kopyat/internal/backup"
	"github.com/karagenc/kopyat/internal/backup/provider"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use: "init",
	Run: func(cmd *cobra.Command, args []string) {
		include := args

		ctx, cancel := context.WithCancel(context.Background())
		addExitHandler(cancel)
		backups, err := backup.FromConfig(ctx, &config.Backups, cacheDir, debugLog, false, include...)
		if err != nil {
			errPrintln(err)
			exit(exitErrAny)
		}

		// If certain backups were choosen:
		if len(args) >= 1 {
			backups = make(backup.Backups)
			for _, name := range args {
				b, ok := backups[name]
				if !ok {
					errPrintln(fmt.Errorf("backup with name %s could not be found", name))
					exit(exitErrAny)
				}
				backups[name] = b
			}
		}

		for _, backup := range backups {
			restic, ok := backup.Provider.(*provider.Restic)
			if !ok {
				errPrintln(fmt.Errorf("backup with name %s is not a restic backup", backup.Name))
				exit(exitErrAny)
			}
			err = restic.Init()
			if err != nil {
				errPrintln(err)
				exit(exitErrAny)
			}
		}
	},
}
