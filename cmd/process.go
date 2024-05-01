package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nyudlts/go-aspace"
	go_bagit "github.com/nyudlts/go-bagit"
)

var (
	woMatcher = regexp.MustCompile("aspace_wo.tsv$")
	tiMatcher = regexp.MustCompile("transfer-info.txt")
	pause     = 1 * time.Millisecond
)

func processAIP(bagLocation string, tmpLocation string) error {

	flag.Parse()

	//ensure that the bag exists and is a directory
	fmt.Println("Performing preliminary checks on AIP: ")
	fmt.Print("  * Checking that bag location exists: ")
	fi, err := os.Stat(bagLocation)
	if err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		return err
	}
	fmt.Print("OK\n")

	//check that
	time.Sleep(pause)
	fmt.Print("  * Checking that bag location is a directory: ")
	if !fi.IsDir() {
		err := fmt.Errorf("%s is not a directory", bagLocation)
		fmt.Printf("KO\t%s\n", err)
		return err
	}
	fmt.Print("OK\n")

	//validate the bag
	time.Sleep(pause)
	fmt.Printf("  * Validating bag at %s: ", bagLocation)
	if err := go_bagit.ValidateBag(bagLocation, false, false); err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		return err
	}
	fmt.Printf("OK\n")

	//find the work order
	fmt.Printf("  * Locating work order: ")
	woPath, err := go_bagit.FindFileInBag(bagLocation, woMatcher)
	if err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		return err
	}
	fmt.Printf("OK\n")

	fmt.Printf("  * validating work order: ")
	//validate the work order
	woBytes, err := os.Open(woPath)
	if err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		return err
	}

	wo := aspace.WorkOrder{}
	if err := wo.Load(woBytes); err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		return err
	}
	if err := woBytes.Close(); err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		return err
	}
	fmt.Printf("OK\n")
	log.Println("- INFO - work order is valid")

	//get the transfer-info.txt
	time.Sleep(pause)
	fmt.Printf("  * Locating transfer-info.txt: ")
	transferInfoPath, err := go_bagit.FindFileInBag(bagLocation, tiMatcher)
	if err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		return err
	}
	fmt.Printf("OK\n")

	//create a tag set from transfer-info.txt
	time.Sleep(pause)
	fmt.Printf("  * Creating new tag set from %s: ", filepath.Base(transferInfoPath))
	transferInfoPath = strings.ReplaceAll(transferInfoPath, bagLocation, "")
	transferInfo, err := go_bagit.NewTagSet(transferInfoPath, bagLocation)
	if err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		return err
	}
	fmt.Printf("OK\n")

	//validate the transfer-info file
	fmt.Printf("  * Validating transfer-info.txt: ")
	if err := validateTransferInfo(transferInfo); err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		return err
	}
	fmt.Printf("OK\n")

	//start the update phase
	time.Sleep(pause)
	fmt.Println("Updating bag at: ", bagLocation)

	//move the work order to bag root and add to tag manifest
	time.Sleep(pause)
	fmt.Printf("  * Moving work order to bag's root ")
	if err := go_bagit.AddFileToBag(bagLocation, woPath); err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		return err
	}
	fmt.Printf("OK\n")

	//Update the hostname
	time.Sleep(pause)
	fmt.Printf("  * Adding hostname to tag set: ")
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		return err
	}
	transferInfo.Tags["nyu-dl-hostname"] = hostname
	fmt.Printf("OK\n")

	//add pathname to the tag-set
	time.Sleep(pause)
	fmt.Printf("  * Adding bag's path to tag set: ")
	path, err := filepath.Abs(bagLocation)
	if err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		return err
	}

	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		return err
	}
	transferInfo.Tags["nyu-dl-pathname"] = path
	fmt.Printf("OK\n")

	//backup bag-info
	time.Sleep(pause)
	fmt.Print("  * Backing up bag-info.txt")
	bagInfoLocation := filepath.Join(bagLocation, "bag-info.txt")
	backupLocation := filepath.Join(tmpLocation, "bag-info.txt")
	backup, err := os.Create(backupLocation)
	if err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		return err
	}
	defer backup.Close()

	//open bag-info.txt
	source, err := os.Open(bagInfoLocation)
	if err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		return err
	}
	defer source.Close()

	_, err = io.Copy(backup, source)
	if err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		return err
	}
	fmt.Printf(" OK\n")

	//getting tagset from bag-info
	fmt.Printf("  * Creating new tag set from %s: ", bagInfoLocation)
	bagInfo, err := go_bagit.NewTagSet("bag-info.txt", bagLocation)
	if err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		return err
	}
	fmt.Printf("OK\n")

	//merge tagsets
	time.Sleep(pause)
	fmt.Printf("  * Merging Tag Sets: ")
	bagInfo.AddTags(transferInfo.Tags)
	fmt.Printf("OK\n")

	time.Sleep(pause)
	//write the new baginfo file
	fmt.Printf("  * Getting data as byte array: ")
	bagInfoBytes := bagInfo.GetTagSetAsByteSlice()
	fmt.Printf("OK\n")

	fmt.Printf("  * Opening bag-info.txt: ")
	bagInfoFile, err := os.Open(bagInfoLocation)
	if err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		return err
	}
	defer bagInfoFile.Close()
	fmt.Printf("OK\n")

	fmt.Printf("  * Writing bag-info.txt: ")
	writer := bufio.NewWriter(bagInfoFile)
	if _, err := writer.Write(bagInfoBytes); err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		return err
	}
	writer.Flush()
	fmt.Printf("OK\n")

	fmt.Printf("  * Rewriting bag-info.txt: ")
	if err := os.WriteFile(bagInfoLocation, bagInfoBytes, 0755); err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		return err
	}
	fmt.Printf("OK\n")

	//create new manifest object for tagmanifest-sha256.txt
	time.Sleep(pause)
	fmt.Printf("  * Creating new manifest for tagmanifest-sha256.txt: ")
	tagManifest, err := go_bagit.NewManifest(bagLocation, "tagmanifest-sha256.txt")
	if err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		return err
	}
	fmt.Printf("OK\n")

	//update the checksum for bag-info.txt
	time.Sleep(pause)
	fmt.Printf("  * Updating checksum for bag-info.txt in tagmanifest-sha256.txt: ")
	if err := tagManifest.UpdateManifest("bag-info.txt"); err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		return err
	}
	fmt.Printf("OK\n")

	time.Sleep(pause)
	fmt.Printf("  * Rewriting tagmanifest-sha256.txt: ")
	if err := tagManifest.Serialize(); err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		return err
	}
	fmt.Printf("OK\n")

	//validate the updated bag
	time.Sleep(pause)
	fmt.Printf("Validating the updated bag: ")
	if err := go_bagit.ValidateBag(bagLocation, false, false); err != nil {
		fmt.Printf("KO\t%s\n", err.Error())
		return err
	}
	fmt.Printf("OK\n")

	fmt.Printf("Package preparation complete for %s\n", filepath.Base(bagLocation))
	return nil
}

func validateTransferInfo(transferInfo go_bagit.TagSet) error {
	for tag, value := range transferInfo.Tags {
		switch tag {
		case "nyu-dl-transfer-type":
			{
				if !(value == "AIP" || value == "DIP") {
					return fmt.Errorf("nyu-dl-transfer-type must be equal to AIP or DIP was %s", value)
				}
			}
		case "nyu-dl-rstar-collection-id":
			{
				if _, err := uuid.Parse(value); err != nil {
					return err
				}
			}
		case "External-Identifier":
			{
				if _, err := uuid.Parse(value); err != nil {
					return err
				}
			}
		case "nyu-dl-project-name:":
			{
				if !partnerAndCode.MatchString(value) {
					return fmt.Errorf("%s is an invalid partner/collection code", value)
				}
			}
		case "Internal-Sender-Identifier":
			{
				if !partnerAndCode.MatchString(value) {
					return fmt.Errorf("%s is an invalid partner/collection code", value)
				}
			}
		default:
		}
	}
	return nil

}
