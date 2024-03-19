package main

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/tomruk/kopyaship/scripting"
)

var runCmd = &cobra.Command{
	Use: "run",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		addExitHandler(cancel)
		script, err := scripting.NewScript(ctx, "./script-playground/foo.go 1 2 3")
		if err != nil {
			exit(err, nil)
		}
		err = script.Run()
		if err != nil {
			exit(err, nil)
		}
	},
}
