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

		//set copy options
		options.PreserveTimes = true
		options.PermissionControl = cp.AddPermission(0755)

		//copy the directory to the staging area
		dst := filepath.Join(stagingLoc, fi.Name())
		fmt.Printf("\nCopying package from %s to %s\n", aipLoc, dst)
		if err := cp.Copy(aipLoc, dst, options); err != nil {
			panic(err)
		}

		//run the update process
		if tmpLocation == "" {
			tmpLocation = "/tmp"
		}

		outputFile, err := os.Create("ampp-results.txt")
		if err != nil {
			panic(err)
		}
		defer outputFile.Close()
		writer := bufio.NewWriter(outputFile)

		fmt.Printf("\nUpdating package at %s\n", dst)
		if err := processAIP(dst, tmpLocation); err != nil {
			writer.WriteString(dst + " " + err.Error())
		} else {
			writer.WriteString(dst + " " + "SUCCESS")
		}

	}
}
