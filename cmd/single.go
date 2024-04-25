package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	cp "github.com/otiai10/copy"
	"github.com/spf13/cobra"
)

var (
	aipLocation   string
	tmpLocation   string
	stageLocation string
)

func init() {
	singleCmd.Flags().StringVar(&aipLocation, "aip-location", "", "")
	singleCmd.Flags().StringVar(&tmpLocation, "tmp-location", "", "")
	singleCmd.Flags().StringVar(&stageLocation, "staging-location", "", "")
	rootCmd.AddCommand(singleCmd)
}

var singleCmd = &cobra.Command{
	Use: "single",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		logFile, err = os.Create(filepath.Base(aipLocation) + ".log")
		if err != nil {
			panic(err)
		}
		defer logFile.Close()
		log.SetOutput(logFile)

		fmt.Println("Running Archivematica Package Prep version", version)
		log.Println("- INFO - Running Archivematica Package Prep version", version)

		if tmpLocation == "" {
			tmpLocation = "/tmp"
		}

		if stageLocation != "" {
			newLocation := filepath.Join(stageLocation, filepath.Base(aipLocation))
			log.Printf("- INFO - Copying %s to %s", aipLocation, newLocation)
			fmt.Printf("Copying %s to %s\n", aipLocation, newLocation)

			if err := cp.Copy(aipLocation, newLocation); err != nil {
				log.Fatalf("- FATAL - %s", err.Error())
			}
			aipLocation = newLocation
		} else {
			log.Printf("- INFO - Processing AIP in place")
			fmt.Printf("Processing AIP in place")
		}

		log.Printf("- INFO - Processing %s", aipLocation)
		if err := processAIP(aipLocation, tmpLocation); err != nil {
			log.Fatalf("- FATAL - %s", err.Error())
		}
	},
}
