package main

import (
	"flag"
	"fmt"
	go_bagit "github.com/nyudlts/go-bagit"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const version string = "0.2.1a"

var (
	bag         string
	bagFiles    = []string{}
	tmpLocation = "/var/archivematica/ampp/tmp/"
	tmpBagDir   string
	uuidMatcher = regexp.MustCompile("\\b[0-9a-f]{8}\\b-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-\\b[0-9a-f]{12}\\b")
	woMatcher   = regexp.MustCompile("aspace_wo.tsv$")
	tiMatcher   = regexp.MustCompile("transfer-info.txt")
	pause       = 500 * time.Millisecond
)

func init() {
	flag.StringVar(&bag, "bag", "", "location of bag")
}

func main() {

	fmt.Println("Running Archivematica Package Prep version", version)
	flag.Parse()

	time.Sleep(pause)

	//ensure that the bag exists and is a directory
	fmt.Println("\nPerforming preliminary checks: ")
	fmt.Print("  * Checking that bag location exists: ")
	fi, err := os.Stat(bag)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Print("OK\n")

	//check that
	time.Sleep(pause)
	fmt.Print("  * Checking that bag location is a directory: ")
	if fi.IsDir() != true {
		log.Fatal(fmt.Errorf("Location provided is not a directory"))
	}
	fmt.Print("OK\n")

	//validate the bag
	time.Sleep(pause)
	fmt.Printf("  * Validating bag at %s: ", bag)
	if err := go_bagit.ValidateBag(bag, false, false); err != nil {
		log.Fatal(err.Error())
	}
	fmt.Printf("OK\n")

	/* taking this out for now
	//create the tmp directory
	time.Sleep(pause)
	tmpBagDir = filepath.Join(tmpLocation, bag)
	fmt.Printf("  * Creating temp dir at %s: ", tmpBagDir)
	err = os.Mkdir(tmpBagDir, 0777)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Print("OK\n")
	*/

	//start the update phase
	time.Sleep(pause)
	fmt.Println("\nUpdating bag at: ", bag)

	//move the work order to bag root and add to tag manifest
	time.Sleep(pause)
	fmt.Printf("  * Locating work order: ")
	woPath, err := go_bagit.FindFileInBag(bag, woMatcher)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Printf("OK\n")

	time.Sleep(pause)
	fmt.Printf("  * Moving work order to bag's root ")
	if err := go_bagit.AddFileToBag(bag, woPath); err != nil {
		panic(err)
	}
	fmt.Printf("OK\n")

	//get the transfer-info.txt
	time.Sleep(pause)
	fmt.Printf("  * Locating transfer-info.txt: ")
	transferInfoPath, err := go_bagit.FindFileInBag(bag, tiMatcher)
	if err != nil {
		panic(err)
	}
	fmt.Printf("OK\n")

	transferInfoPath = strings.ReplaceAll(transferInfoPath, bag, "")

	//create a tag set from transfer-info.txt
	time.Sleep(pause)
	fmt.Printf("  * Creating new tag set from %s: ", transferInfoPath)
	transferInfo, err := go_bagit.NewTagSet(transferInfoPath, bag)
	if err != nil {
		panic(err)
	}
	fmt.Printf("OK\n")

	//Update the hostname
	time.Sleep(pause)
	fmt.Printf("  * Adding hostname to tag set: ")
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	transferInfo.Tags["nyu-dl-hostname"] = hostname
	fmt.Printf("OK\n")

	//add pathname to the tag-set
	time.Sleep(pause)
	fmt.Printf("  * Adding bag's path to tag set: ")
	path, err := filepath.Abs(bag)
	if err != nil {
		log.Fatal(err)
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		log.Fatal(err)
	}
	transferInfo.Tags["nyu-dl-pathname"] = path
	fmt.Printf("OK\n")

	time.Sleep(pause)
	bagInfoLocation := filepath.Join(bag, "bag-info.txt")
	//backup bag-info
	backupLocation := filepath.Join("/var/archivematica/tmp", "bag-info.txt")
	backup, err := os.Create(backupLocation)
	if err != nil {
		log.Fatal(err)
	}
	defer backup.Close()

	source, err := os.Open(bagInfoLocation)
	if err != nil {
		log.Fatal(err)
	}
	defer source.Close()

	_, err = io.Copy(backup, source)
	if err != nil {
		log.Fatal(err)
	}

	//getting tagset from bag-info
	fmt.Printf("  * Creating new tag set from %s: ", bagInfoLocation)
	bagInfo, err := go_bagit.NewTagSet("bag-info.txt", bag)
	if err != nil {
		panic(err)
	}
	fmt.Printf("OK\n")

	//merge tagsets
	time.Sleep(pause)
	fmt.Printf("  * Merging Tag Sets: ")
	bagInfo.AddTags(transferInfo.Tags)
	fmt.Printf("OK\n")

	time.Sleep(pause)
	fmt.Printf("  * Rewriting bag-info.txt with updated tag set: ")
	//write the new baginfo file
	if err := bagInfo.Serialize(); err != nil {
		panic(err)
	}
	fmt.Printf("OK\n")

	//create new manifest object for tagmanifest-sha256.txt
	time.Sleep(pause)
	fmt.Printf("  * Creating new manifest for tagmanifest-sha256.txt: ")
	tagManifest, err := go_bagit.NewManifest(bag, "tagmanifest-sha256.txt")
	if err != nil {
		panic(err)
	}
	fmt.Printf("OK\n")

	//update the checksum for bag-info.txt
	time.Sleep(pause)
	fmt.Printf("  * Updating checksum for bag-info.txt in tagmanifest-sha256.txt: ")
	if err := tagManifest.UpdateManifest("bag-info.txt"); err != nil {
		panic(err)
	}
	fmt.Printf("OK\n")

	time.Sleep(pause)
	fmt.Printf("  * Rewriting tagmanifest-sha256.txt: ")
	if err := tagManifest.Serialize(); err != nil {
		panic(err)
	}
	fmt.Printf("OK\n")

	//validate the updated bag
	time.Sleep(pause)
	fmt.Printf("\nValidating the updated bag: ")
	if err := go_bagit.ValidateBag(bag, false, false); err != nil {
		panic(err)
	}
	fmt.Printf("OK\n")

	fmt.Println("\nPackage preparation complete")
	os.Exit(0)
}
