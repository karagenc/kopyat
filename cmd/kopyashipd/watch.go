package main

import "github.com/tomruk/kopyaship/ifile"

func (v *svice) initJobsFromConfig() (jobs []*ifile.WatchJob, err error) {
	for _, run := range v.config.IfileGeneration.Run {
		j := ifile.NewWatchJob(v.log, run.ScanPath, run.Ifile)
		jobs = append(jobs, j)
		v.jobsMu.Lock()
		v.jobs = append(v.jobs, j)
		v.jobsMu.Unlock()
	}
	return
}
