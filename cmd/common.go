package cmd

import (
	"os"
)

const version string = "0.2.5"

var (
	logFileName = "ampp.log"
	logFile     *os.File
)
