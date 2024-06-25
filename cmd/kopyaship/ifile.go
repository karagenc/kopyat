package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tomruk/kopyaship/internal/ifile"
)

var (
	ifileCmd         = &cobra.Command{Use: "ifile"}
	ifileGenerateCmd = &cobra.Command{Use: "generate"}

	ifileGenerateSyncthingCmd = &cobra.Command{
		Use:  "syncthing",
		Long: "Manually generate syncthing ifile at the specified directory",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dir := args[0]
			stignore := filepath.Join(dir, ".stignore")

			stat, err := os.Stat(dir)
			if err != nil {
				errPrintln(err)
				exit(exitErrAny)
			}
			if !stat.IsDir() {
				errPrintln(fmt.Errorf("not a directory: %s", dir))
				exit(exitErrAny)
			}

			fmt.Printf("Creating/opening %s\n", stignore)
			ifile, err := ifile.New(stignore, ifile.ModeSyncthing, true, debugLog)
			if err != nil {
				errPrintln(err)
				exit(exitErrAny)
			}
			defer ifile.Close() // For panic
			addExitHandler(func() { ifile.Close() })

			fmt.Printf("Walking %s\n", dir)
			err = ifile.Walk(dir)
			if err != nil {
				errPrintln(err)
				exit(exitErrAny)
			}
		},
	}
)
