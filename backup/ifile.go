package backup

import (
	"context"
	"path/filepath"

	"github.com/tomruk/kopyaship/ifile"
	"golang.org/x/sync/errgroup"
)

func (p *paths) ifilePath() string {
	return filepath.Join(p.cacheDir, p.backup.Name+".list")
}

func (p *paths) generateIfile() error {
	g, _ := errgroup.WithContext(context.Background())

	i, err := ifile.New(p.ifilePath(), ifile.Include, false, p.log)
	if err != nil {
		return err
	}
	defer i.Close()

	for _, path := range p.Paths() {
		path := path
		g.Go(func() error {
			return i.Walk(path)
		})
	}
	return g.Wait()
}
