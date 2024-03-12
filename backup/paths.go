package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type paths struct {
	backup *Backup
	// Not to be confused with Base:
	// This is kopyaship's base directory.
	cacheDir string

	base  string
	paths []string
}

func (p *paths) Paths() []string { return p.paths }

func (p *paths) check() error {
	if p.base != "" {
		if !filepath.IsAbs(p.base) {
			return fmt.Errorf("backup base path `%s` is not absolute. to avoid confusion, backup base path must be absolute.", p.base)
		}
		p.base = strings.TrimSuffix(p.base, "/")
		p.base = os.ExpandEnv(p.base)
		if _, err := os.Stat(p.base); os.IsNotExist(err) {
			return err
		}
	}

	for i, path := range p.paths {
		if path == "" {
			return fmt.Errorf("one of the backup paths is empty")
		}
		if !filepath.IsAbs(path) {
			return fmt.Errorf("backup path `%s` is not absolute. to avoid confusion, either backup base path or the paths must be absolute.", path)
		}
		path = strings.TrimSuffix(path, "/")
		path = os.ExpandEnv(path)
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
