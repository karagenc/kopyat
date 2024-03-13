package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/tomruk/kopyaship/ifile"
)

var (
	watchCmd = &cobra.Command{Use: "watch"}

	watchListCmd = &cobra.Command{
		Use: "list",
		Run: func(cmd *cobra.Command, args []string) {
			hc, err := newHTTPClient()
			if err != nil {
				exit(err)
			}
			resp, err := hc.Get("/jobs/watch")
			if err != nil {
				exit(err)
			}
			defer resp.Body.Close()
			content, err := io.ReadAll(resp.Body)
			if err != nil {
				exit(err)
			}

			var infos []*ifile.WatchJobInfo
			err = json.Unmarshal(content, &infos)
			if err != nil {
				exit(err)
			}

			fmt.Println()
			w := table.NewWriter()
			w.AppendHeader(table.Row{
				"IFILE", "SCAN PATH", "MODE", "ERRORS",
			})
			for _, info := range infos {
				w.AppendRow(table.Row{
					info.Ifile, info.ScanPath, info.Mode, info.Errors,
				})
			}
			fmt.Println(w.Render())
			fmt.Println()
		},
	}
)
