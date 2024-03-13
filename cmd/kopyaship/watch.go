package main

import (
	"encoding/json"
	"fmt"
	"io"

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

			for _, info := range infos {
				fmt.Printf("Watch Info:\n%s\n\n", info.String())
			}
		},
	}
)
