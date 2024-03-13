package main

import (
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
		j := ifile.NewWatchJob(v.log, run.ScanPath, run.Ifile)
		jobs = append(jobs, j)
		v.jobsMu.Lock()
		v.watchJobs = append(v.watchJobs, j)
		v.jobsMu.Unlock()
	}
	return
}
