package main

import (
	"flag"
	"fmt"
	go_bagit "github.com/nyudlts/go-bagit"
	"os"
)

var bag string

func init() {
	flag.StringVar(&bag, "bag", "", "location of bag")
}
func main() {
	flag.Parse()
	go_bagit.Logger().SetOutput(os.Stderr)

	if err := go_bagit.ValidateBag(bag, false, false); err != nil {
		fmt.Println(err)
	}
}
