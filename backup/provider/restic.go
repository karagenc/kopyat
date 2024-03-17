package provider

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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
		repoPath:  filepath.ToSlash(repoPath),
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
	path = filepath.ToSlash(path)
	command := fmt.Sprintf("restic -r %s backup", r.repoPath)
	if r.extraArgs != "" {
		command += " " + r.extraArgs
	}
	command += " " + path
	return r.run(command)
}

func (r *Restic) BackupWithIfile(ifile string) error {
	ifile = filepath.ToSlash(ifile)
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

	ctx, cancel := context.WithCancel(context.Background())
	utils.AddExitHandler(func() { cancel() })

	cmd := exec.CommandContext(ctx, w[0], w[1:]...)
	// https://stackoverflow.com/questions/33165530/prevent-ctrlc-from-interrupting-exec-command-in-golang/33171307
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
