package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tomruk/kopyaship/utils"
)

type paths struct {
	log      utils.Logger
	cacheDir string

	backup *Backup
	base   string
	paths  []string
}

func (p *paths) Paths() []string { return p.paths }

func (p *paths) check() error {
	if p.base != "" {
		p.base = strings.TrimSuffix(p.base, "/")
		if _, err := os.Stat(p.base); os.IsNotExist(err) {
			return err
		}
	}

	for i, path := range p.paths {
		if path == "" {
			return fmt.Errorf("one of the backup paths is empty")
		}
		path = strings.TrimSuffix(path, "/")
		path = filepath.Join(p.base, path)

		if _, err := os.Stat(path); os.IsNotExist(err) {
			return err
		}
		p.paths[i] = path
	}

	return checkPathCollision(p.paths)
}

func checkPathCollision(paths []string) error {
	for i, longer := range paths {
		longerSplitted := strings.Split(longer, "/")

	outer:
		for j, shorter := range paths {
			if i == j {
				continue
			}
			if longer == shorter {
				return fmt.Errorf("duplicate path: %s", longer)
			}
			shorterSplitted := strings.Split(shorter, "/")
			if len(longerSplitted) < len(shorterSplitted) {
				continue
			}

			for i := range shorterSplitted {
				if shorterSplitted[i] != longerSplitted[i] {
					continue outer
				}
			}
			return fmt.Errorf("path collision: %s collides with %s", shorter, longer)
		}
	}
	return nil
}
