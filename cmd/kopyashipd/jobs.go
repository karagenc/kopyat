package main

import "github.com/tomruk/kopyaship/ifile"

type job interface {
	Start() error
	Stop() error
}

type ifileGenerationJob struct {
	v       *svice
	stopped chan struct{}

	ScanPath string
	Ifile    string
}

func (j *ifileGenerationJob) Start() error {
	go func() {
		// TODO: Arguments
		i, err := ifile.New(j.Ifile, ifile.Include, true, true)
		if err != nil {
			j.v.log.Error(err.Error())
		}
		defer i.Close()
		err = i.Walk(j.ScanPath)
		if err != nil {
			j.v.log.Error(err.Error())
		}

		ifile.Watch()
	}()
	return nil
}

func (j *ifileGenerationJob) Stop() error {
	return nil
}

func (v *svice) initJobsFromConfig() (jobs []job, err error) {
	for _, run := range v.config.IfileGeneration.Run {
		j := &ifileGenerationJob{
			ScanPath: run.ScanPath,
			Ifile:    run.Ifile,
			stopped:  make(chan struct{}),
		}
		jobs = append(jobs, j)
		v.jobsMu.Lock()
		v.jobs = append(v.jobs, j)
		v.jobsMu.Unlock()
	}
	return
}
