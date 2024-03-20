package main

import (
	"fmt"
	"os/exec"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use: "doctor",
	Run: func(cmd *cobra.Command, args []string) {
		errorFound := false
		bold.Println("Doctor:")

		fmt.Printf("    Using configuration file: %s\n", v.ConfigFileUsed())

		restic, err := exec.LookPath("restic")
		if err != nil {
			warn.Print("    Warning: ")
			fmt.Printf("restic not found: %v\n", err)
			errorFound = true
		} else {
			fmt.Printf("    restic found at: %s\n", restic)
		}

		if !errorFound {
			color.HiGreen("All good.")
		} else {
			color.Red("Errors occured.")
			code := 1
			exit(nil, &code)
		}
	},
}
