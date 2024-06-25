package scripting

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/mattn/go-shellwords"
	"github.com/tomruk/kopyaship/internal/scripting/ctx"
)

type Script interface {
	// c can be nil
	Run(c ctx.Context) error
	Path() string
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
	} else if len(sw) >= 2 && sw[0] == "sudo" {
		bin := sw[1]
		ext := filepath.Ext(bin)
		switch ext {
		case ".go":
			currExe, err := os.Executable()
			if err != nil {
				return nil, err
			}
			newSW := append([]string{"sudo", "-E", currExe, "run"}, sw[1:]...)
			script = newExec(ctx, newSW...)
		case ".sh":
			script = newShellScript(ctx, "bash", true, sw[1:]...)
		case ".zsh":
			script = newShellScript(ctx, "zsh", true, sw[1:]...)
		default:
			var stat fs.FileInfo
			stat, err = os.Stat(bin)
			if err == nil && stat.IsDir() {
				err = fmt.Errorf("reading scripts from directories is not supported at the moment")
				break
			} else {
				err = nil
			}
			script = newExec(ctx, sw...)
		}
	} else {
		bin := sw[0]
		ext := filepath.Ext(bin)
		switch ext {
		case ".go":
			script, err = newYaegiScript(ctx, sw...)
			// If error occurs, don't break and directly return
			// because yaegi already includes the script path.
			if err != nil {
				return nil, fmt.Errorf("scripting: %v", err)
			}
		case ".sh":
			script = newShellScript(ctx, "bash", false, sw...)
		case ".zsh":
			script = newShellScript(ctx, "zsh", false, sw...)
		default:
			var stat fs.FileInfo
			stat, err = os.Stat(bin)
			if err == nil && stat.IsDir() {
				err = fmt.Errorf("reading scripts from directories is not supported at the moment")
				break
			} else {
				err = nil
			}
			script = newExec(ctx, sw...)
		}
	}

	if err != nil {
		err = fmt.Errorf("scripting: %s: %v", command, err)
	}
	return
}
