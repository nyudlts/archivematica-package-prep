package cmd

import (
	"bufio"
	"log"
	"os"
)

var (
	logFileName = "ampp.log"
	logFile     *os.File
)

func init() {
	var err error
	logFile, err = os.Create(logFileName)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	logWriter := bufio.NewWriter(logFile)
	log.SetOutput(logWriter)
}
