package main

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use: "doctor",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Doctor:")
		fmt.Printf("    Using configuration file: %s\n", v.ConfigFileUsed())
		restic, err := exec.LookPath("restic")
		if err != nil {
			fmt.Printf("    Warning: restic executable not found: %v\n", err)
		} else {
			fmt.Printf("    restic found at: %s\n", restic)
		}
	},
}
