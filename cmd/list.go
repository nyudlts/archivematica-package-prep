package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	cp "github.com/otiai10/copy"
	"github.com/spf13/cobra"
)

var (
	aipFileLoc string
	stagingLoc string
	options    cp.Options
)

func init() {
	listCmd.Flags().StringVar(&aipFileLoc, "aip-file", "", "")
	listCmd.Flags().StringVar(&stagingLoc, "staging-location", "", "")
	listCmd.Flags().StringVar(&tmpLocation, "tmp-location", "", "")
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use: "list",
	Run: func(cmd *cobra.Command, args []string) {
		processList()
	},
}

func processList() {
	fmt.Println(aipFileLoc)
	aipFile, err := os.Open(aipFileLoc)
	if err != nil {
		panic(err)
	}
	defer aipFile.Close()
	scanner := bufio.NewScanner(aipFile)
	for scanner.Scan() {
		aipLoc := scanner.Text()
		fi, err := os.Stat(aipLoc)
		if err != nil {
			panic(err)
		}
		fmt.Println(fi.Name())

		//set copy options
		options.PreserveTimes = true
		options.PermissionControl = cp.AddPermission(0644)

		//copy the directory to the staging area
		dst := filepath.Join(stagingLoc, fi.Name())
		if err := cp.Copy(aipLoc, dst, options); err != nil {
			panic(err)
		}

		//run the update process
		if tmpLocation == "" {
			tmpLocation = "/tmp"
		}

		if err := processAIP(dst, tmpLocation); err != nil {
			fmt.Println(err.Error())
		}

	}
}
