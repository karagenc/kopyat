package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/karagenc/kopyat/internal/ifile"
	"github.com/karagenc/kopyat/internal/scripting/ctx"
	"github.com/karagenc/kopyat/internal/utils"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var (
	watchJobCmd = &cobra.Command{Use: "watch-job"}

	watchJobListCmd = &cobra.Command{
		Use: "list",
		Run: func(cmd *cobra.Command, args []string) {
			hc, err := newHTTPClient()
			if err != nil {
				errPrintln(err)
				exit(exitErrAny)
			}
			resp, err := hc.Get("/watch-job")
			if err != nil {
				errPrintln(err)
				exit(exitErrAny)
			}
			defer resp.Body.Close()
			content, err := io.ReadAll(resp.Body)
			if err != nil {
				errPrintln(err)
				exit(exitErrAny)
			}

			var infos []*ifile.WatchJobInfo
			err = json.Unmarshal(content, &infos)
			if err != nil {
				errPrintln(err)
				exit(exitErrAny)
			}

			fmt.Println()
			w := table.NewWriter()
			w.AppendHeader(table.Row{
				"IFILE", "MODE", "ERRORS",
			})
			for _, info := range infos {
				e := ""
				for i, err := range info.Errors {
					e += utils.Red.Sprint(err)
					if i != len(info.Errors)-1 {
						e += "\n"
					}
				}
				w.AppendRow(table.Row{
					info.Ifile, info.Mode, e,
				})
			}
			fmt.Println(w.Render())
			fmt.Println()
		},
	}

	watchJobStopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stop a watch job",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			hc, err := newHTTPClient()
			if err != nil {
				errPrintln(err)
				exit(exitErrAny)
			}
			body, err := json.Marshal(args)
			if err != nil {
				errPrintln(err)
				exit(exitErrAny)
			}
			resp, err := hc.Post("/watch-job/stop", "application/json", bytes.NewBuffer(body))
			if err != nil {
				errPrintln(err)
				exit(exitErrAny)
			}
			defer resp.Body.Close()
			body, err = io.ReadAll(resp.Body)
			if err != nil {
				errPrintln(err)
				exit(exitErrAny)
			}
			var errs []string
			err = json.Unmarshal(body, &errs)
			if err != nil {
				errPrintln(err)
				exit(exitErrAny)
			}

			if len(errs) > 0 {
				utils.Error.Println("Errors occured:")
				for _, err := range errs {
					fmt.Printf("%s\n", err)
				}
				return
			}
			utils.Success.Println("Successful")
		},
	}
)

func (s *svc) getWatchJobs(c echo.Context) error {
	s.jobsMu.Lock()
	watchJobs := s.watchJobs
	infos := make([]*ifile.WatchJobInfo, 0, len(s.watchJobs))
	s.jobsMu.Unlock()

	for _, job := range watchJobs {
		info := job.Info()
		infos = append(infos, info)
	}
	return c.JSON(http.StatusOK, infos)
}

func (s *svc) stopWatchJobs(c echo.Context) error {
	var ifiles []string
	err := c.Bind(&ifiles)
	if err != nil {
		return err
	}

	s.jobsMu.Lock()
	errGroup := errgroup.Group{}
	for _, ifile := range ifiles {
		for i, job := range s.watchJobs {
			job := job
			if job.Ifile() == ifile {
				s.watchJobs = append(s.watchJobs[:i], s.watchJobs[i+1:]...)
				errGroup.Go(func() error { return job.Shutdown() })
			}
		}
	}
	s.jobsMu.Unlock()

	err = errGroup.Wait()
	if err != nil {
		return c.JSON(http.StatusOK, []string{err.Error()})
	}
	return c.JSON(http.StatusOK, []string{})
}

func (s *svc) initWatchJobs() (jobs []*ifile.WatchJob, err error) {
	for _, run := range config.IfileGeneration.Run {
		newHookRunner := func(hooks []string, c ctx.Context) func() error {
			return func() error {
				errGroup := &errgroup.Group{}
				for _, hook := range hooks {
					err := runHook(errGroup, hook, c)
					if err != nil {
						return err
					}
				}
				return errGroup.Wait()
			}
		}

		runPreHooks := newHookRunner(run.Hooks.Pre, ctx.NewIfileGenerationContext(true, run.Ifile, run.Mode))
		runPostHooks := newHookRunner(run.Hooks.Post, ctx.NewIfileGenerationContext(false, run.Ifile, run.Mode))

		var mode ifile.Mode
		switch run.Mode {
		case "syncthing":
			mode = ifile.ModeSyncthing
		default:
			if run.Mode == "" {
				return nil, fmt.Errorf("empty `mode` field. check config")
			}
			return nil, fmt.Errorf("invalid `mode` field: %s", run.Mode)
		}

		var job *ifile.WatchJob
		s.jobsMu.Lock()
		for i, j := range s.watchJobs {
			if run.Ifile == job.Ifile() {
				job = j
				s.watchJobs = append(s.watchJobs[:i], s.watchJobs[i+1:]...)
				break
			}
		}
		s.jobsMu.Unlock()
		if job != nil {
			shutdownErr := job.Shutdown()
			if shutdownErr != nil {
				s.log.Error(shutdownErr.Error())
			}
		}

		job = ifile.NewWatchJob(run.Ifile, filepath.Dir(run.Ifile), mode, runPreHooks, runPostHooks, s.log)

		jobs = append(jobs, job)
		s.jobsMu.Lock()
		s.watchJobs = append(s.watchJobs, job)
		s.jobsMu.Unlock()
	}
	return
}
