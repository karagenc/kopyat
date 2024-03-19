package scripting

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/mattn/go-shellwords"
)

type Script interface {
	Run() error
	Location() string
}

func NewScript(ctx context.Context, command string) (script Script, err error) {
	parser := shellwords.NewParser()
	parser.ParseBacktick = true
	parser.ParseEnv = true

	var sw []string
	sw, err = parser.Parse(command)
	if err != nil {
		return
	}
	if len(sw) == 0 {
		return nil, fmt.Errorf("scripting: empty command")
	}

	bin := sw[0]
	ext := filepath.Ext(bin)
	switch ext {
	case ".go":
		script, err = newYaegiScript(ctx, sw...)
		// If error occurs, don't break and directly return
		// because yaegi already includes the script location.
		if err != nil {
			return nil, fmt.Errorf("scripting: %v", err)
		}
	case ".sh":
		script = newShellScript(ctx, "bash", sw...)
	case ".zsh":
		script = newShellScript(ctx, "zsh", sw...)
	default:
		var stat fs.FileInfo
		stat, err = os.Stat(bin)
		if err != nil {
			break
		}
		if stat.IsDir() {
			err = fmt.Errorf("reading scripts from directories is not supported at the moment")
			break
		} else {
			script = newExec(ctx, sw...)
		}
	}

	if err != nil {
		err = fmt.Errorf("scripting: %s: %v", command, err)
	}
	return
}
