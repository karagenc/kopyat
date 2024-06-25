package main

import (
	"context"
	"strings"

	"github.com/tomruk/kopyaship/internal/scripting"
	"github.com/tomruk/kopyaship/internal/scripting/ctx"
	"golang.org/x/sync/errgroup"
)

func runHook(errGroup *errgroup.Group, command string, c ctx.Context) error {
	goRoutine := false
	if strings.HasPrefix(command, "go ") {
		command = command[3:]
		goRoutine = true
	}

	ctx, cancel := context.WithCancel(context.Background())
	addExitHandler(cancel)
	script, err := scripting.NewScript(ctx, command)
	if err != nil {
		return err
	}

	run := func() error { return script.Run(c) }
	if goRoutine {
		errGroup.Go(run)
		return nil
	} else {
		return run()
	}
}
