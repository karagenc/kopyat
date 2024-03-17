package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

var pingCmd = &cobra.Command{
	Use: "ping",
	Run: func(cmd *cobra.Command, args []string) {
		err := ping()
		if err != nil {
			exit(err, nil)
		}
		fmt.Println("Pong!")
	},
}

func ping() error {
	hc, err := newHTTPClient()
	if err != nil {
		return err
	}
	r, err := hc.Get("/ping")
	if err != nil {
		return err
	}
	defer r.Body.Close()
	content, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if string(content) != "Pong" {
		return fmt.Errorf("invalid response received: %s", string(content))
	}
	return nil
}
