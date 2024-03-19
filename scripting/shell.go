package scripting

import (
	"context"
	"os"
	"os/exec"
)

type ShellScript struct {
	ctx   context.Context
	shell string
	sw    []string
}

func newShellScript(ctx context.Context, shell string, sw ...string) *ShellScript {
	return &ShellScript{
		ctx:   ctx,
		shell: shell,
		sw:    sw,
	}
}

func (s *ShellScript) Location() string { return s.sw[0] }

func (s *ShellScript) Run() error {
	cmd := exec.CommandContext(s.ctx, s.shell, s.sw...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
