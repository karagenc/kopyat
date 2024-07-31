package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/karagenc/kopyat/internal/backup"
	_ctx "github.com/karagenc/kopyat/internal/scripting/ctx"
	"github.com/karagenc/kopyat/internal/utils"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func init() {
	f := backupCmd.Flags()
	f.Bool("no-remind", false, "Disable reminders")
	f.Bool("no-hook", false, "Disable hook scripts")
}

var backupCmd = &cobra.Command{
	Use: "backup",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			f           = cmd.Flags()
			noRemind, _ = f.GetBool("no-remind")
			noHook, _   = f.GetBool("no-hook")
			include     = args
		)

		remindAll := func(reminders []string) {
			if !noRemind {
				if len(reminders) == 0 {
					return
				}
				utils.BgWhite.Print("Reminders — Hit 'Enter' after completing each task.")
				fmt.Print("\n")
				for _, reminder := range reminders {
					fmt.Printf("    %s", reminder)
					bufio.NewReader(os.Stdin).ReadBytes('\n')
				}
				fmt.Println()
			}
		}

		runHooks := func(hooks []string, c _ctx.Context) error {
			if !noHook {
				errGroup := errgroup.Group{}
				for i, hook := range hooks {
					fmt.Println()
					utils.Bold.Printf("Running hook %d of %d: %s", i+1, len(hooks), hook)
					fmt.Print("\n\n")
					err := runHook(&errGroup, hook, c)
					if err != nil {
						return err
					}
				}
				return errGroup.Wait()
			}
			return nil
		}

		ctx, cancel := context.WithCancel(context.Background())
		addExitHandler(cancel)
		backups, err := backup.FromConfig(ctx, &config.Backups, cacheDir, debugLog, false, include...)
		if err != nil {
			errPrintln(err)
			exit(exitErrAny)
		}

		for _, backup := range backups {
			if !noRemind {
				remindAll(backup.Config.Reminders.Pre)
			}
			skip := false
			if !noHook {
				err = runHooks(
					backup.Config.Hooks.Pre,
					_ctx.NewBackupContext(
						true,
						backup.Name,
						backup.Provider.TargetPath(),
						backup.Config.Base,
						backup.Config.Paths,
						func() {
							skip = true
						},
						backup.UseIfile,
					))
				if err != nil {
					errPrintln(fmt.Errorf("failed to run pre hook: %v: exiting", err))
					exit(exitErrAny)
				}
			}
			if skip {
				utils.BgWhite.Printf("Skipping backup: %s\n", backup.Name)
				continue
			}

			err = backup.Do()
			if err != nil {
				errPrintln(err)
				exit(exitErrAny)
			}

			if !noHook {
				err = runHooks(
					backup.Config.Hooks.Post,
					_ctx.NewBackupContext(
						false,
						backup.Name,
						backup.Provider.TargetPath(),
						backup.Config.Base,
						backup.Config.Paths,
						func() {}, // Noop for post hooks
						backup.UseIfile,
					))
				if err != nil {
					errPrintln(fmt.Errorf("failed to run post hook: %v: exiting", err))
					exit(exitErrAny)
				}
			}
			if !noRemind {
				remindAll(backup.Config.Reminders.Post)
			}
		}

		utils.Success.Println("\nBackup successful")
	},
}
