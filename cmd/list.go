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
	fmt.Println("Running Archivematica Package Prep version", version)
	log.Println("- INFO - Running Archivematica Package Prep version", version)

	fmt.Println("\nInitial setup")
	var err error
	logFileLocation := filepath.Join(stagingLoc, logFileName)
	fmt.Printf("  * Creating log file %s: ", logFileLocation)

	logFile, err = os.Create(logFileLocation)
	if err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		log.Fatalf("- FATAL - %s", err.Error())
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	fmt.Println("OK")

	fmt.Printf("  * Opening aip-file %s: ", aipFileLoc)
	log.Println("- INFO -  Opening aip-file", aipFileLoc)

	aipFile, err := os.Open(aipFileLoc)
	if err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		log.Fatalf("- FATAL - %s", err.Error())
	}
	fmt.Println("OK")
	scanner := bufio.NewScanner(aipFile)

	summaryFileLocation := filepath.Join(stagingLoc, "ampp-results.txt")
	fmt.Printf("  * Creating output results file %s: ", summaryFileLocation)
	summaryFile, err := os.Create(summaryFileLocation)
	if err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		log.Fatalf("- FATAL - %s", err.Error())
	}
	defer summaryFile.Close()
	fmt.Println("OK")
	writer := bufio.NewWriter(summaryFile)

	//set copy options
	options.PreserveTimes = true
	options.PermissionControl = cp.AddPermission(0755)

	if tmpLocation == "" {
		tmpLocation = "/tmp"
	}

	for scanner.Scan() {
		aipLoc := scanner.Text()
		fmt.Printf("\nProcessing AIP %s\n", filepath.Base(aipLoc))
		log.Printf("- INFO - processing AIP %s", filepath.Base(aipLoc))

		//check the aip exists
		fmt.Printf("Checking that AIP exists at %s: ", aipLoc)
		fi, err := os.Stat(aipLoc)
		if err != nil {
			fmt.Printf("KO\t%s\n", err.Error())
			log.Fatalf("- FATAL - %s", err.Error())
		}
		fmt.Println("OK")

		//copy the directory to the staging area
		dst := filepath.Join(stagingLoc, fi.Name())
		fmt.Printf("Copying AIP %s to %s: ", filepath.Base(aipLoc), dst)
		log.Printf("- INFO - copying AIP %s to %s\n", filepath.Base(aipLoc), dst)
		if err := cp.Copy(aipLoc, dst, options); err != nil {
			fmt.Printf("KO%s\n", err.Error())
			log.Fatalf("- FATAL - %s", err.Error())
		}
		fmt.Println("OK")

		//run the update process
		if err := processAIP(dst, tmpLocation); err != nil {
			writer.WriteString(fmt.Sprintf("%s\t%s\n", dst, err.Error()))
			writer.Flush()
			log.Fatalf("- FATAL - %s", err.Error())
		} else {
			writer.WriteString(fmt.Sprintf("%s\tSUCCESS\n", dst))
			writer.Flush()
		}
		writer.Flush()
	}

}
