package scripting

import (
	"context"
	"os"
	"os/exec"
)

type ShellScript struct {
	ctx   context.Context
	shell string
	sudo  bool
	sw    []string
}

func newShellScript(ctx context.Context, shell string, sudo bool, sw ...string) *ShellScript {
	return &ShellScript{
		ctx:   ctx,
		shell: shell,
		sudo:  sudo,
		sw:    sw,
	}
}

func (s *ShellScript) Location() string { return s.sw[0] }

func (s *ShellScript) Run() error {
	var cmd *exec.Cmd
	if s.sudo {
		sw := append([]string{s.shell}, s.sw...)
		cmd = exec.CommandContext(s.ctx, "sudo", sw...)
	} else {
		cmd = exec.CommandContext(s.ctx, s.shell, s.sw...)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
