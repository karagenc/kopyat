package ifile

import (
	"os"
	"strings"
	"sync"

	pathspec "github.com/shibumi/go-pathspec"
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

func (i *Ifile) Close() error { return i.file.Close() }

func (e *entry) String() string {
	s := e.path + "\n"
	s = strings.ReplaceAll(s, "[", "\\[")
	s = strings.ReplaceAll(s, "]", "\\]")
	return s
}
