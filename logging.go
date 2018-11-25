package main

import (
	"io"
	"log"
	"os"
)

var (
	//logFlags contains a list of flags to use when logging information
	logFlags = log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile

	//Debug logs debugging and tracing information
	Debug *log.Logger

	//Info logs information the reader can use to know what is happening
	Info *log.Logger

	//Warning logs ignoreable issues that the reader may wish to know about
	Warning *log.Logger

	//Error logs information that the reader should use to resolve breaking issues
	Error *log.Logger
)

func initLogging(logFile *os.File, processType string) {
	Debug = log.New(io.MultiWriter(logFile, os.Stdout), "["+processType+"] DEBUG: ", logFlags)
	Info = log.New(io.MultiWriter(logFile, os.Stdout), "["+processType+"] INFO: ", logFlags)
	Warning = log.New(io.MultiWriter(logFile, os.Stdout), "["+processType+"] WARNING: ", logFlags)
	Error = log.New(io.MultiWriter(logFile, os.Stderr), "["+processType+"] ERROR: ", logFlags)
}
