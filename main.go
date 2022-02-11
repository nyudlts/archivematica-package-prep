package main

import (
	"fmt"
	go_bagit "github.com/nyudlts/go-bagit"
)

func main() {
	testBag := "./test-bag/"
	fmt.Println(testBag)
	if err := go_bagit.ValidateBag(testBag, false, false); err != nil {
		fmt.Println(err)
	}
}
