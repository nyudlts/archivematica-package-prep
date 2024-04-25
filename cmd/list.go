package cmd

import (
	"bufio"
	"fmt"
	"log"
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
	var err error
	logFile, err = os.Create("ampp-list.log")
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	fmt.Println("Running Archivematica Package Prep version", version)
	log.Println("- INFO - Running Archivematica Package Prep version", version)

	fmt.Println("Parsing aip-file:", aipFileLoc)
	log.Println("- INFO - Parsing aip-file:", aipFileLoc)

	aipFile, err := os.Open(aipFileLoc)
	if err != nil {
		log.Fatal("- FATAL - ", err.Error())
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
		log.Printf("- INFO - copying package from %s to %s\n", aipLoc, dst)
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
		log.Printf("- INFO - Updating package at %s\n", dst)
		if err := processAIP(dst, tmpLocation); err != nil {
			writer.WriteString(dst + " " + err.Error())
		} else {
			writer.WriteString(dst + " " + "SUCCESS")
		}
	}
}
