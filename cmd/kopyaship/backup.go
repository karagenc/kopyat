package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tomruk/kopyaship/backup"
	"github.com/tomruk/kopyaship/scripting"
	"github.com/tomruk/kopyaship/utils"
	"golang.org/x/sync/errgroup"

	"github.com/fatih/color"
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
			include      = args
		)

		ctx, cancel := context.WithCancel(context.Background())
		addExitHandler(cancel)
		backups, err := backup.FromConfig(ctx, &config.Backups, stateDir, log, false, include...)
		if err != nil {
			exit(err, nil)
		}

		if !noRemind {
			reminders := config.Backups.Reminders.Pre
			for _, backup := range config.Backups.Run {
				_, ok := backups[backup.Name]
				if ok {
					reminders = append(reminders, backup.Reminders.Pre...)
				}
			}
			remindAll(reminders)
		}
		if !noHook {
			hooks := config.Backups.Hooks.Pre
			for _, backup := range config.Backups.Run {
				_, ok := backups[backup.Name]
				if ok {
					hooks = append(hooks, backup.Hooks.Pre...)
				}
			}
			g := &errgroup.Group{}
			for i, hook := range hooks {
				fmt.Println()
				utils.Bold.Printf("Running hook %d of %d: %s", i+1, len(hooks), hook)
				fmt.Print("\n\n")
				err := runHook(g, hook)
				if err != nil {
					exit(fmt.Errorf("pre hook failed: %v: exiting.", err), nil)
				}
			}
			err = g.Wait()
			if err != nil {
				exit(fmt.Errorf("pre hook failed: %v: exiting.", err), nil)
			}
		}

		for _, backup := range backups {
			err = backup.Do()
			if err != nil {
				exit(err, nil)
			}
		}

		if !noRemind {
			reminders := config.Backups.Reminders.Post
			for _, backup := range config.Backups.Run {
				_, ok := backups[backup.Name]
				if ok {
					reminders = append(reminders, backup.Reminders.Post...)
				}
			}
			if len(reminders) > 0 {
				fmt.Println()
				remindAll(reminders)
			}
		}
		if !noHook {
			hooks := config.Backups.Hooks.Post
			for _, backup := range config.Backups.Run {
				_, ok := backups[backup.Name]
				if ok {
					hooks = append(hooks, backup.Hooks.Post...)
				}
			}

			g := &errgroup.Group{}
			for i, hook := range hooks {
				utils.Bold.Printf("\nRunning hook %d of %d: %s", i+1, len(hooks), hook)
				fmt.Print("\n\n")
				err := runHook(g, hook)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s\n", utils.Red.Sprintf("Post hook failed: %v", err))
					postHookFail = true
				}
			}
			err = g.Wait()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", utils.Red.Sprintf("Post hook failed: %v", err))
				postHookFail = true
			}
		}

		if postHookFail {
			color.Red("\nBackup successful with at least 1 post hook failure.\nConsider running failed scripts by hand.")
		} else {
			color.HiGreen("\nBackup successful")
		}
	},
}

func remindAll(reminders []string) {
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

func runHook(g *errgroup.Group, command string) error {
	goroutine := false
	if strings.HasPrefix(command, "go ") {
		command = command[3:]
		goroutine = true
	}

	ctx, cancel := context.WithCancel(context.Background())
	addExitHandler(cancel)
	script, err := scripting.NewScript(ctx, command)
	if err != nil {
		return err
	}

	if goroutine {
		g.Go(script.Run)
		return nil
	} else {
		return script.Run()
	}
}
