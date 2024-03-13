package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

var pingCmd = &cobra.Command{
	Use: "ping",
	Run: func(cmd *cobra.Command, args []string) {
		hc, err := newHTTPClient()
		if err != nil {
			exit(err)
		}
		r, err := hc.Get("/ping")
		if err != nil {
			exit(err)
		}
		defer r.Body.Close()
		content, err := io.ReadAll(r.Body)
		if err != nil {
			exit(err)
		}
		if string(content) != "Pong" {
			exit(fmt.Errorf("invalid response received: %s", string(content)))
		}
		fmt.Printf("%s!\n", string(content))
	},
}
