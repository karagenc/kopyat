package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/karagenc/kopyat/internal/utils"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use: "doctor",
	Run: func(cmd *cobra.Command, args []string) {
		errorFound := false
		utils.Bold.Println("Doctor:")
		fmt.Printf("    Using config: %s\n", v.ConfigFileUsed())

		restic, err := exec.LookPath("restic")
		if err != nil {
			utils.Warn.Printf("    Warning: restic not found: %v\n", err)
			errorFound = true
		} else {
			fmt.Printf("    restic found at: %s\n", restic)
		}

		lockDir := filepath.Dir(lockFile)
		createLockDir := func() {
			_, err := os.Stat(lockFile)
			if os.IsNotExist(err) {
				f, err := os.Create(lockFile)
				if err != nil {
					utils.Error.Print("    Error creating lockfile: ")
					fmt.Printf("%v\n", err)
					errorFound = true
					if os.IsPermission(err) {
						utils.Warn.Println("    Looks like you don't have permission to create the lockfile. Try running doctor with sudo.")
					}
				} else {
					fmt.Printf("    Lockfile %s is created.\n", lockFile)
				}
				if f != nil {
					f.Close()
				}
			} else if err != nil {
				utils.Error.Print("    Error stat'ing lockfile: ")
				fmt.Printf("%v\n", err)
			} else {
				fmt.Printf("    Lockfile is at: %s\n", lockFile)
			}
		}

		if _, err := os.Stat(lockDir); os.IsNotExist(err) {
			fmt.Printf("    Lockfile directory %s (for service) doesn't exist. Creating it.\n", lockDir)
			err := os.MkdirAll(lockDir, 0755)
			if err != nil {
				utils.Error.Print("        Error: ")
				fmt.Printf("%v\n", err)
				errorFound = true
			} else {
				createLockDir()
			}
		} else {
			createLockDir()
		}

		hc, err := newHTTPClient()
		if err != nil {
			utils.Error.Printf("    Error: %v\n", err)
			errorFound = true
		} else {
			fmt.Printf("    Pinging to API: %s\n", hc)
			err = ping()
			if err != nil {
				fmt.Printf("        API is %s: Error: %v\n", utils.Red.Sprint("down"), err)
				errorFound = true
			} else {
				fmt.Printf(`        API is %s: "Pong" received`+"\n", utils.HiGreen.Sprint("up"))
			}
		}

		if !errorFound {
			utils.Success.Println("All good.")
		} else {
			color.Red("Error(s) occured.")
			exit(exitErrAny)
		}
	},
}
