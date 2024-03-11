package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tomruk/kopyaship/backup"
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
			f            = cmd.Flags()
			noRemind, _  = f.GetBool("no-remind")
			noHook, _    = f.GetBool("no-hook")
			postHookFail = false
		)

		backups, err := backup.FromConfig(config, cacheDir, true)
		if err != nil {
			exit(err)
		}

		if !noRemind {
			remindAll(config.Backups.Reminders.Pre)
		}
		if !noHook {
			hooks := config.Backups.Hooks.Pre
			for _, backup := range config.Backups.Run {
				_, ok := backups[backup.Name]
				if ok {
					hooks = append(hooks, backup.Hooks.Pre...)
				}
			}
			for i, hook := range hooks {
				fmt.Printf("\nRunning hook %d of %d: %s\n\n", i, len(hooks), hook)
				err := runHook(hook)
				if err != nil {
					exit(fmt.Errorf("pre hook failed: %v: exiting.", err))
				}
			}

		}

		for _, backup := range backups {
			err = backup.Do()
			if err != nil {
				exit(err)
			}
		}

		if !noRemind {
			remindAll(config.Backups.Reminders.Post)
		}
		if !noHook {
			hooks := config.Backups.Hooks.Post
			for _, backup := range config.Backups.Run {
				_, ok := backups[backup.Name]
				if ok {
					hooks = append(hooks, backup.Hooks.Post...)
				}
			}
			for i, hook := range hooks {
				fmt.Printf("\nRunning hook %d of %d: %s\n\n", i, len(hooks), hook)
				err := runHook(hook)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Post hook failed: %v\n", err)
					postHookFail = true
				}
			}
		}

		if postHookFail {
			fmt.Println("Backup successful with at least 1 post hook failure.\nConsider running failed scripts by hand.")
		} else {
			fmt.Println("Backup successful")
		}
	},
}

func remindAll(reminders []string) {
	if len(reminders) == 0 {
		return
	}
	fmt.Printf("Reminders — Hit 'Enter' after completing each task.\n\n")
	for _, reminder := range reminders {
		fmt.Printf("    %s", reminder)
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
	fmt.Printf("\n")
}

func runHook(script string) error { return nil }
