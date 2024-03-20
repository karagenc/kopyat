package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/tomruk/kopyaship/ifile"
	"github.com/tomruk/kopyaship/scripting"
	"golang.org/x/sync/errgroup"
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
		runHook := func(g *errgroup.Group, command string) error {
			goroutine := false
			if strings.HasPrefix(command, "go ") {
				command = command[3:]
				goroutine = true
			}

			ctx, cancel := context.WithCancel(context.Background())
			v.addExitHandler(cancel)
			script, err := scripting.NewScript(ctx, command)
			if err != nil {
				return err
			}

			if goroutine {
				g.Go(script.Run)
				return nil
			} else {
				return script.Run()
			}
		}

		runPreHooks := func() error {
			preHooks := v.config.IfileGeneration.Hooks.Pre
			preHooks = append(preHooks, run.Hooks.Pre...)

			g := &errgroup.Group{}
			for _, hook := range preHooks {
				err := runHook(g, hook)
				if err != nil {
					return err
				}
			}
			return g.Wait()
		}

		runPostHooks := func() error {
			postHooks := v.config.IfileGeneration.Hooks.Post
			postHooks = append(postHooks, run.Hooks.Post...)

			g := &errgroup.Group{}
			for _, hook := range postHooks {
				err := runHook(g, hook)
				if err != nil {
					return err
				}
			}
			return g.Wait()
		}

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

		j := ifile.NewWatchJob(run.Ifile, mode, runPreHooks, runPostHooks, v.log)

		jobs = append(jobs, j)
		v.jobsMu.Lock()
		v.watchJobs = append(v.watchJobs, j)
		v.jobsMu.Unlock()
	}
	return
}
