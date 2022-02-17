package main

import (
	"bufio"
	"flag"
	"fmt"
	go_bagit "github.com/nyudlts/go-bagit"
	cp "github.com/otiai10/copy"
	"os"
	"path/filepath"
	"regexp"
)

var (
	input       string
	bag         = "test-bag-copy"
	bagFiles    = []string{}
	uuidMatcher = regexp.MustCompile("\\b[0-9a-f]{8}\\b-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-\\b[0-9a-f]{12}\\b")
	woMatcher   = regexp.MustCompile("aspace_wo.tsv$")
	tiMatcher   = regexp.MustCompile("transfer-info.txt")
	version     = "0.1.0a"
	rstarID     string
)

func init() {
	flag.StringVar(&input, "input", "", "location of bag")
	flag.StringVar(&rstarID, "rstar-id", "", "rstar id of the collection")
}

func main() {
	fmt.Println("Running Archivematica Package Prep version", version)
	fi, err := os.Stat(bag)
	if err != nil {
		//do nothing for now
	} else {
		if fi.IsDir() {
			err := os.RemoveAll(bag)
			if err != nil {
				panic(err)
			}
		}
	}

	flag.Parse()
	// resolve any symlinks
	inputPath, err := filepath.EvalSymlinks(input)
	if err != nil {
		panic(err)
	}

	//copy the bag (for dev purposes)
	err = cp.Copy(inputPath, "test-bag-copy")

	//ensure that the bag exists and is a directory
	fi, err = os.Stat(bag)
	if err != nil {
		panic(err)
	}

	if fi.IsDir() != true {
		panic(fmt.Errorf("Location provided is not a directory"))
	}

	//validate the copied bag
	if err := go_bagit.ValidateBag(bag, false, false); err != nil {
		panic(err)
	}

	//find the workorder
	woPath, err := go_bagit.FindFileInBag(bag, woMatcher)
	if err != nil {
		panic(err)
	}

	if err := go_bagit.AddFileToBag(bag, woPath); err != nil {
		panic(err)
	}

	//get the transfer-info.txt
	transferInfoPath, err := go_bagit.FindFileInBag(bag, tiMatcher)
	if err != nil {
		panic(err)
	}

	transferInfoPath = transferInfoPath[len(bag)+1:]
	//Get the contents of transfer-info.txt
	transferInfo, err := go_bagit.NewTagSet(transferInfoPath, bag)
	if err != nil {
		panic(err)
	}

	//append the hostname to bag-info.txt
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	transferInfo.Tags["nyu-dl-hostname"] = hostname

	//append the pathname
	path, err := filepath.Abs(bag)
	if err != nil {
		panic(err)
	}
	transferInfo.Tags["nyu-dl-pathname"] = path

	bagInfo, err := go_bagit.NewTagSet("bag-info.txt", bag)

	//check for rstar collection id
	if bagInfo.HasTag("nyu-dl-rstar-collection-id") != true {
		//check if the rstarID is set
		if rstarID == "" {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter the rstar uuid: ")
			rstarID, err = reader.ReadString('\n')
			if err != nil {
				panic(err)
			}
		}

		//make sure there is a valid uuid
		if uuidMatcher.MatchString(rstarID) != true {
			panic(fmt.Sprintf("%s is not a valid uuid", rstarID))
		}

		transferInfo.Tags["nyu-dl-rstar-collection-id"] = rstarID
	}

	//update the software agent
	transferInfo.Tags["Bag-Software-Agent"] = go_bagit.GetSoftwareAgent()

	//update the tagmap
	bagInfo.AddTags(transferInfo.Tags)

	//write the new baginfo file
	if err := bagInfo.Serialize(); err != nil {
		panic(err)
	}

	//update the tag manifest
	tagManifest, err := go_bagit.NewManifest(bag, "tagmanifest-sha256.txt")
	if err != nil {
		panic(err)
	}

	if err := tagManifest.UpdateManifest("bag-info.txt"); err != nil {
		panic(err)
	}

	if err := tagManifest.Serialize(); err != nil {
		panic(err)
	}

	//validate the bag
	if err := go_bagit.ValidateBag(bag, false, false); err != nil {
		panic(err)
	}

	fmt.Println("Package preparation complete")
	os.Exit(0)
}
