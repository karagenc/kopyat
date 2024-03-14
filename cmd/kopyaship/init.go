package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tomruk/kopyaship/backup"
	"github.com/tomruk/kopyaship/backup/provider"
)

var initCmd = &cobra.Command{
	Use: "init",
	Run: func(cmd *cobra.Command, args []string) {
		backups, err := backup.FromConfig(config, cacheDir, log, false)
		if err != nil {
			exit(err)
		}

		// If certain backups were choosen:
		if len(args) >= 1 {
			backups = make(backup.Backups)
			for _, name := range args {
				b, ok := backups[name]
				if !ok {
					exit(fmt.Errorf("backup with name %s could not be found", name))
				}
				backups[name] = b
			}
		}

		for _, backup := range backups {
			restic, ok := backup.Provider.(*provider.Restic)
			if !ok {
				exit(fmt.Errorf("backup with name %s is not a restic backup", backup.Name))
			}
			err = restic.Init()
			if err != nil {
				exit(err)
			}
		}
	},
}
