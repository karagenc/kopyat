package scripting

import (
	"context"
	"os"
	"os/exec"
)

type Exec struct {
	ctx context.Context
	w   []string
}

func newExec(ctx context.Context, w ...string) *Exec {
	return &Exec{
		ctx: ctx,
		w:   w,
	}
}

func (e *Exec) Location() string { return e.w[0] }

func (e *Exec) Run() error {
	cmd := exec.CommandContext(e.ctx, e.w[0], e.w[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
