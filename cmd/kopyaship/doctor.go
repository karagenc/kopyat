package main

import (
	"fmt"
	"os/exec"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/tomruk/kopyaship/utils"
)

var doctorCmd = &cobra.Command{
	Use: "doctor",
	Run: func(cmd *cobra.Command, args []string) {
		errorFound := false
		utils.Bold.Println("Doctor:")

		fmt.Printf("    Using configuration file: %s\n", v.ConfigFileUsed())

		restic, err := exec.LookPath("restic")
		if err != nil {
			utils.Warn.Printf("    Warning: restic not found: %v\n", err)
			errorFound = true
		} else {
			fmt.Printf("    restic found at: %s\n", restic)
		}

		hc, err := newHTTPClient()
		if err != nil {
			utils.Error.Printf("    Error: %v\n", err)
			errorFound = true
		} else {
			fmt.Printf("    Pinging to API: %s\n", hc)
			err = ping()
			if err != nil {
				fmt.Printf("    API is %s: Error: %v\n", utils.Red.Sprint("down"), err)
				errorFound = true
			} else {
				fmt.Printf(`    API is %s: "Pong" received`+"\n", utils.HiGreen.Sprint("up"))
			}
		}

		if !errorFound {
			color.HiGreen("All good.")
		} else {
			color.Red("Error(s) occured.")
			code := 1
			exit(nil, &code)
		}
	},
}
