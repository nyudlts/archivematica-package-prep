package cmd

import (
	"os"
	"regexp"
)

const version string = "0.2.6"

var (
	logFileName    = "ampp.log"
	logFile        *os.File
	partnerAndCode = regexp.MustCompile(`^[tamwag|fales|nyuarchives]`)
)
