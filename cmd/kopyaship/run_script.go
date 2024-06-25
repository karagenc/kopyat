package main

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tomruk/kopyaship/internal/scripting"
	_ctx "github.com/tomruk/kopyaship/internal/scripting/ctx"
)

var runScript = &cobra.Command{
	Use:  "run-script",
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		addExitHandler(cancel)
		command := strings.Join(args, " ")

		script, err := scripting.NewScript(ctx, command)
		if err != nil {
			errPrintln(err)
			exit(exitErrAny)
		}
		err = script.Run(_ctx.NewEmptyContext())
		if err != nil {
			errPrintln(err)
			exit(exitErrAny)
		}
		return nil
	},
}
