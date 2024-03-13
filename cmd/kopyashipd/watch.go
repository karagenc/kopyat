package main

import "github.com/tomruk/kopyaship/ifile"

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
