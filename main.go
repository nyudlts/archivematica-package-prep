package main

import (
	"bufio"
	"flag"
	"fmt"
	go_bagit "github.com/nyudlts/go-bagit"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"
)

var (
	bag          string
	bagFiles     = []string{}
	rstarID      string
	copyLocation string
	uuidMatcher  = regexp.MustCompile("\\b[0-9a-f]{8}\\b-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-\\b[0-9a-f]{12}\\b")
	woMatcher    = regexp.MustCompile("aspace_wo.tsv$")
	tiMatcher    = regexp.MustCompile("transfer-info.txt")
	version      = "0.1.0a"
	pause        = 500 * time.Millisecond
)

func init() {
	flag.StringVar(&bag, "bag", "", "location of bag")
	flag.StringVar(&rstarID, "rstar-id", "", "rstar id of the collection")
	flag.StringVar(&copyLocation, "copy-location", "", "location to copy the bag before processing")
}

func main() {

	fmt.Println("Running Archivematica Package Prep version", version)
	flag.Parse()

	time.Sleep(pause)
	//ensure that the bag exists and is a directory
	fmt.Print("  Checking that bag location exists: ")
	fi, err := os.Stat(bag)
	if err != nil {
		panic(err)
	}
	fmt.Print("OK\n")

	time.Sleep(pause)
	fmt.Print("  Checking that bag location is a directory: ")
	if fi.IsDir() != true {
		panic(fmt.Errorf("Location provided is not a directory"))
	}
	fmt.Print("OK\n")

	if copyLocation != "" {
		time.Sleep(pause)
		fmt.Printf("Copying %s to %s\n", bag, copyLocation)
		//if the copy location exists, delete it

		time.Sleep(pause)
		fmt.Printf("  Checking if %s exists: ", copyLocation)
		fi, err := os.Stat(copyLocation)
		if err != nil {
			fmt.Printf("OK\n")
		} else {
			fmt.Printf("OK\n")
			if fi.IsDir() {
				time.Sleep(pause)
				fmt.Printf("  Removing directory at %s: ", copyLocation)
				err := os.RemoveAll(copyLocation)
				if err != nil {
					panic(err)
				}
				fmt.Printf("OK\n")
			}
		}

		// resolve any symlinks
		time.Sleep(pause)
		fmt.Printf("  Resolving any symlinks: ")
		inputPath, err := filepath.EvalSymlinks(bag)
		if err != nil {
			panic(err)
		}
		fmt.Printf("OK\n")

		//copy the directory
		time.Sleep(pause)
		fmt.Printf("  Copying bag to copy location: ")
		cmd := exec.Command("cp", "-r", inputPath, copyLocation)
		_, err = cmd.Output()
		if err != nil {
			panic(err)
		}
		fmt.Printf("OK\n")
		//use the copy of the bag
		bag = copyLocation
	}

	time.Sleep(pause)
	fmt.Println("Updating bag at: ", bag)

	//validate the copied bag
	time.Sleep(pause)
	fmt.Printf("  Validating bag at %s: ", bag)
	if err := go_bagit.ValidateBag(bag, false, false); err != nil {
		panic(err)
	}
	fmt.Printf("OK\n")

	time.Sleep(pause)
	fmt.Printf("  Locating work order: ")
	//find the workorder
	woPath, err := go_bagit.FindFileInBag(bag, woMatcher)
	if err != nil {
		panic(err)
	}
	fmt.Printf("OK\n")

	time.Sleep(pause)
	fmt.Printf("  Moving work order to bag's root :")
	if err := go_bagit.AddFileToBag(bag, woPath); err != nil {
		panic(err)
	}
	fmt.Printf("OK\n")

	//get the transfer-info.txt
	time.Sleep(pause)
	fmt.Printf("  Locating transfer-info.txt: ")
	transferInfoPath, err := go_bagit.FindFileInBag(bag, tiMatcher)
	if err != nil {
		panic(err)
	}
	fmt.Printf("OK\n")

	time.Sleep(pause)
	fmt.Printf("  Creating new tag set from %s: ", transferInfoPath)
	transferInfoPath = transferInfoPath[len(bag)+1:]
	//Get the contents of transfer-info.txt
	transferInfo, err := go_bagit.NewTagSet(transferInfoPath, bag)
	if err != nil {
		panic(err)
	}
	fmt.Printf("OK\n")

	time.Sleep(pause)
	fmt.Printf("  Adding hostname to tag set: ")
	//append the hostname to bag-info.txt
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	transferInfo.Tags["nyu-dl-hostname"] = hostname
	fmt.Printf("OK\n")

	time.Sleep(pause)
	fmt.Printf("  Adding bag's path to tag set: ")
	//append the pathname
	path, err := filepath.Abs(bag)
	if err != nil {
		panic(err)
	}

	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		panic(err)
	}

	transferInfo.Tags["nyu-dl-pathname"] = path
	fmt.Printf("OK\n")

	time.Sleep(pause)
	fmt.Printf("  Checking for R* ID")
	//check for rstar collection id
	if transferInfo.HasTag("nyu-dl-rstar-collection-id") != true {
		//check if the rstarID is set
		if rstarID == "" {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("    Enter the collections rstar uuid: ")
			rstarID, err = reader.ReadString('\n')
			if err != nil {
				panic(err)
			}
		} else {
			fmt.Printf(": OK\n")
		}
	} else {
		fmt.Printf(": OK\n")
	}

	time.Sleep(pause)
	fmt.Printf("  Validating R* ID: ")
	if uuidMatcher.MatchString(rstarID) != true {
		panic(fmt.Sprintf("%s is not a valid uuid", rstarID))
	}
	fmt.Printf("OK\n")

	time.Sleep(pause)
	fmt.Printf("  Adding R* ID to Tag Set:")
	transferInfo.Tags["nyu-dl-rstar-collection-id"] = rstarID
	fmt.Printf("OK\n")
	time.Sleep(pause)
	fmt.Printf("  Updating Software Agent in Tag Set\n")
	//update the software agent
	transferInfo.Tags["Bag-Software-Agent"] = go_bagit.GetSoftwareAgent()

	bagInfoLocation := filepath.Join(bag, "bag-info.txt")
	//Get a tag set for bag-info.txt
	time.Sleep(pause)
	fmt.Printf("  Creating new tag set from %s: ", bagInfoLocation)
	bagInfo, err := go_bagit.NewTagSet("bag-info.txt", bag)
	if err != nil {
		panic(err)
	}
	fmt.Printf("OK\n")

	time.Sleep(pause)
	fmt.Printf("  Merging Tag Sets: ")
	//update the tagmap
	bagInfo.AddTags(transferInfo.Tags)
	fmt.Printf("OK\n")

	time.Sleep(pause)
	fmt.Printf("  Rewriting bag-info.txt with updated tag set: ")
	//write the new baginfo file
	if err := bagInfo.Serialize(); err != nil {
		panic(err)
	}
	fmt.Printf("OK\n")

	//update the tag manifest
	time.Sleep(pause)
	fmt.Printf("  Creating new manifest for tagmanifest-sha256.txt: ")
	tagManifest, err := go_bagit.NewManifest(bag, "tagmanifest-sha256.txt")
	if err != nil {
		panic(err)
	}
	fmt.Printf("OK\n")

	time.Sleep(pause)
	fmt.Printf("  Updating checksum for bag-info.txt in tagmanifest-sha256.txt: ")
	if err := tagManifest.UpdateManifest("bag-info.txt"); err != nil {
		panic(err)
	}
	fmt.Printf("OK\n")

	time.Sleep(pause)
	fmt.Printf("  Rewriting tagmanifest-sha256.txt: ")
	if err := tagManifest.Serialize(); err != nil {
		panic(err)
	}
	fmt.Printf("OK\n")

	//validate the bag
	time.Sleep(pause)
	fmt.Printf("  Validating the updated bag: ")
	if err := go_bagit.ValidateBag(bag, false, false); err != nil {
		panic(err)
	}
	fmt.Printf("OK\n")

	fmt.Println("Package preparation complete")
	os.Exit(0)
}
