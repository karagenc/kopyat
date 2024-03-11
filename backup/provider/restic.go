package provider

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/mattn/go-shellwords"
)

type Restic struct {
	repoPath  string
	extraArgs string
	sudo      bool
}

func NewRestic(repoPath string, extraArgs string) *Restic {
	return &Restic{
		repoPath:  repoPath,
		extraArgs: extraArgs,
		sudo:      false,
	}
}

func (r *Restic) TargetLocation() string { return r.repoPath }

func (r *Restic) Init() error {
	return r.run(fmt.Sprintf("restic -r %s init", r.repoPath))
}

func (r *Restic) Backup(path string) error {
	command := fmt.Sprintf("restic -r %s backup", r.repoPath)
	if r.extraArgs != "" {
		command += " " + r.extraArgs
	}
	command += " " + path
	return r.run(command)
}

func (r *Restic) BackupWithIfile(ifile string) error {
	command := fmt.Sprintf("restic -r %s backup", r.repoPath)
	if r.extraArgs != "" {
		command += " " + r.extraArgs
	}
	return r.run(fmt.Sprintf("%s --files-from %s", command, ifile))
}

func (r *Restic) run(command string) error {
	parser := shellwords.NewParser()
	parser.ParseBacktick = true
	parser.ParseEnv = true

	if r.sudo {
		command = "sudo " + command
	}

	w, err := parser.Parse(command)
	if err != nil {
		return err
	}
	if len(w) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := exec.Command(w[0], w[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
