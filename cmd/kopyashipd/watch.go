package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tomruk/kopyaship/ifile"
)

func (v *svice) getWatchJobs(c echo.Context) error {
	v.jobsMu.Lock()
	infos := make([]*ifile.WatchJobInfo, 0, len(v.watchJobs))
	for _, job := range v.watchJobs {
		info := job.Info()
		infos = append(infos, info)
	}
	v.jobsMu.Unlock()
	return c.JSON(http.StatusOK, infos)
}

func (v *svice) initWatchJobsFromConfig() (jobs []*ifile.WatchJob, err error) {
	for _, run := range v.config.IfileGeneration.Run {
		var mode ifile.Mode
		switch run.For {
		case "syncthing":
			mode = ifile.ModeSyncthing
		default:
			if run.For == "" {
				return nil, fmt.Errorf("empty 'for' field. check configuration.")
			}
			return nil, fmt.Errorf("invalid 'for': %s", run.For)
		}
		j := ifile.NewWatchJob(v._log, run.ScanPath, run.Ifile, mode)
		jobs = append(jobs, j)
		v.jobsMu.Lock()
		v.watchJobs = append(v.watchJobs, j)
		v.jobsMu.Unlock()
	}
	return
}
