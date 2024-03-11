package ifile

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	pathspec "github.com/shibumi/go-pathspec"
)

const (
	gitignore = ".gitignore"
	csignore  = ".csignore"
)

type (
	Ifile struct {
		filePath string
		file     *os.File
		fileMu   sync.Mutex
	}

	entry struct {
		path  string
		isDir bool
	}

	entries []*entry

	ignorefile struct {
		p   *pathspec.PathSpec
		dir string
	}
)

func New(filePath string) (i *Ifile, err error) {
	i = &Ifile{
		filePath: filePath,
	}
	i.file, err = os.OpenFile(i.filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	return
}

func (i *Ifile) Walk(path string) error {
	ignorefiles := make([]*ignorefile, 0, 100)
	entries := make(entries, 0, 10000)
	i.addIgnoreIfExists(&ignorefiles, path)

	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if e, ok := err.(*fs.PathError); ok {
				if e, ok := e.Err.(syscall.Errno); ok && e.Is(fs.ErrPermission) {
					return nil
				}
			}
			return err
		}

		t := d.Type()
		for i := len(ignorefiles) - 1; i >= 0; i-- {
			f := ignorefiles[i]

			if strings.HasPrefix(path, f.dir) {
				trimmed := path[len(f.dir):]
				if t.IsDir() && !strings.HasSuffix(trimmed, "/") {
					trimmed += "/"
				}
				if f.p.Match(trimmed) {
					if t.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}
		}

		// if t.IsDir() {
		// 	i.addIgnoreIfExists(&ignorefiles, path)
		// }

		entries = append(entries, &entry{
			path:  path,
			isDir: t.IsDir(),
		})
		return nil
	})
	if err != nil {
		return err
	}

	i.fileMu.Lock()
	defer i.fileMu.Unlock()

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

		e := entry.String()
		e = strings.ReplaceAll(e, "[", "\\[")
		e = strings.ReplaceAll(e, "]", "\\]")
		_, err = i.file.WriteString(e)
		if err != nil {
			return err
		}
	}

	_, err = i.file.WriteString("\n")
	return err
}

func (i *Ifile) addIgnoreIfExists(ignorefiles *[]*ignorefile, dir string) error {
	gitignorePath := filepath.Join(dir, gitignore)
	csignorePath := filepath.Join(dir, csignore)

	if f, err := os.Stat(gitignorePath); err == nil && f.Mode().Type().IsRegular() {
		p, err := pathspec.FromFile(gitignorePath)
		if err != nil {
			return err
		}
		*ignorefiles = append(*ignorefiles, &ignorefile{
			p:   p,
			dir: dir,
		})
	}

	if f, err := os.Stat(csignorePath); err == nil && f.Mode().Type().IsRegular() {
		p, err := pathspec.FromFile(csignorePath)
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

func (i *Ifile) Close() error { return i.file.Close() }

func (e *entry) String() string { return e.path + "\n" }

func splitPath(path string) []string {
	splitted := strings.Split(path, string(os.PathSeparator))
	if splitted[0] == "" {
		splitted[0] = "/"
	}
	return splitted
}
