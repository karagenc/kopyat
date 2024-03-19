package main

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tomruk/kopyaship/scripting"
)

var runCmd = &cobra.Command{
	Use:  "run",
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		addExitHandler(cancel)
		command := strings.Join(args, " ")

		script, err := scripting.NewScript(ctx, command)
		if err != nil {
			exit(err, nil)
		}
		err = script.Run()
		if err != nil {
			exit(err, nil)
		}
		return nil
	},
}
