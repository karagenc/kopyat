package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Paths struct {
	backup *Backup
	// Not to be confused with Base:
	// This is kopyaship's base directory.
	cacheDir string

	Base  string
	Paths []string
}

func (p *Paths) check() error {
	if p.Base != "" {
		p.Base = strings.TrimSuffix(p.Base, "/")
		p.Base = os.ExpandEnv(p.Base)
		if !filepath.IsAbs(p.Base) {
			return fmt.Errorf("backup path is not absolute")
		}
		if _, err := os.Stat(p.Base); os.IsNotExist(err) {
			return err
		}
	}

	for i, path := range p.Paths {
		if path == "" {
			return fmt.Errorf("one of the paths is empty")
		}

		path = strings.TrimSuffix(path, "/")
		path = os.ExpandEnv(path)
		path = filepath.Join(p.Base, path)

		if _, err := os.Stat(path); os.IsNotExist(err) {
			return err
		}
		p.Paths[i] = path
	}

	return checkPathCollision(p.Paths)
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
