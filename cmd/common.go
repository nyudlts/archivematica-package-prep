package cmd

import (
	"os"
)

const version string = "0.2.6"

var (
	logFileName = "ampp.log"
	logFile     *os.File
)
