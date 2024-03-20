package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tomruk/kopyaship/backup"
	"github.com/tomruk/kopyaship/backup/provider"
)

var initCmd = &cobra.Command{
	Use: "init",
	Run: func(cmd *cobra.Command, args []string) {
		include := args

		ctx, cancel := context.WithCancel(context.Background())
		addExitHandler(cancel)
		backups, err := backup.FromConfig(ctx, &config.Backups, cacheDir, log, false, include...)
		if err != nil {
			exit(err, nil)
		}

		// If certain backups were choosen:
		if len(args) >= 1 {
			backups = make(backup.Backups)
			for _, name := range args {
				b, ok := backups[name]
				if !ok {
					exit(fmt.Errorf("backup with name %s could not be found", name), nil)
				}
				backups[name] = b
			}
		}

		for _, backup := range backups {
			restic, ok := backup.Provider.(*provider.Restic)
			if !ok {
				exit(fmt.Errorf("backup with name %s is not a restic backup", backup.Name), nil)
			}
			err = restic.Init()
			if err != nil {
				exit(err, nil)
			}
		}
	},
}
