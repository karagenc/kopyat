package provider

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/mattn/go-shellwords"
	"github.com/tomruk/kopyaship/utils"
)

type Restic struct {
	repoPath  string
	extraArgs string
	sudo      bool
	password  string

	log utils.Logger
}

func NewRestic(repoPath, extraArgs, password string, sudo bool, log utils.Logger) *Restic {
	return &Restic{
		repoPath:  repoPath,
		extraArgs: extraArgs,
		sudo:      sudo,
		password:  password,
		log:       log,
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

func (r *Restic) PasswordIsSet() bool {
	return r.password != "" || os.Getenv("RESTIC_PASSWORD") != ""
}

func (r *Restic) run(command string) error {
	parser := shellwords.NewParser()
	parser.ParseBacktick = true
	parser.ParseEnv = true

	if r.sudo {
		command = "sudo " + command
	}
	r.log.Infof("Running: %s\n", command)
	if r.password != "" {
		os.Setenv("RESTIC_PASSWORD", r.password)
		defer os.Unsetenv("RESTIC_PASSWORD")
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
