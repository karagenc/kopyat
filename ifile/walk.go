package ifile

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	pathspec "github.com/shibumi/go-pathspec"
)

const (
	gitignore = ".gitignore"
	csignore  = ".csignore"
)

func (i *Ifile) Walk(root string) error {
	ignorefiles := make([]*ignorefile, 0, 100)
	entries := make(entries, 0, 10000)
	err := addIgnoreIfExists(&ignorefiles, root)
	if err != nil {
		return err
	}

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if e, ok := err.(*fs.PathError); ok {
				if e, ok := e.Err.(syscall.Errno); ok && e.Is(fs.ErrPermission) {
					return nil
				}
			}
			return err
		}

		t := d.Type()
		if t.IsDir() {
			err = addIgnoreIfExists(&ignorefiles, path)
			if err != nil {
				return err
			}
		}

		for j := len(ignorefiles) - 1; j >= 0; j-- {
			igf := ignorefiles[j]

			if strings.HasPrefix(path, igf.dir) {
				trimmed := path[len(igf.dir):]
				if t.IsDir() && !strings.HasSuffix(trimmed, "/") {
					trimmed += "/"
				}
				match := igf.p.Match(trimmed)
				if (match && i.mode == ModeRestic || !match && i.mode == ModeSyncthing) && path != root {
					if t.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}
		}

		entries = append(entries, &entry{
			path:  path,
			isDir: t.IsDir(),
		})
		return nil
	})
	if err != nil {
		return err
	}

	growLen := 0
	for _, entry := range entries {
		growLen += entry.Len()
	}

	i.bufMu.Lock()
	defer i.bufMu.Unlock()
	i.buf.Grow(growLen)

outer:
	for _, entry := range entries {
		// If the entry is a directory, check if it contains a children.
		// If it's empty (doesn't have a children), add it to the list.
		if entry.isDir {
			for _, _entry := range entries {
				if strings.HasPrefix(_entry.path, entry.path) && strings.ContainsRune(strings.TrimPrefix(_entry.path, entry.path), os.PathSeparator) {
					continue outer
				}
			}
		}
		s := entry.String()
		if _, ok := i.existing[s[:len(s)-1]]; ok {
			continue
		}

		_, err = i.buf.WriteString(s)
		if err != nil {
			return err
		}
	}
	return err
}

func addIgnoreIfExists(ignorefiles *[]*ignorefile, dir string) error {
	path := filepath.Join(dir, gitignore)
	if f, err := os.Stat(path); err == nil && f.Mode().Type().IsRegular() {
		p, err := pathspec.FromFile(path)
		if err != nil {
			return err
		}
		*ignorefiles = append(*ignorefiles, &ignorefile{
			p:   p,
			dir: dir,
		})
	}

	path = filepath.Join(dir, csignore)
	if f, err := os.Stat(path); err == nil && f.Mode().Type().IsRegular() {
		p, err := pathspec.FromFile(path)
		if err != nil {
			return err
		}
		*ignorefiles = append(*ignorefiles, &ignorefile{
			p:   p,
			dir: dir,
		})
	}
	return nil
}
