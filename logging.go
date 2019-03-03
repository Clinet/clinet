package main

import (
	"io"
	"io/ioutil"
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

	//Same as above, but for the API
	DebugAPI   *log.Logger
	InfoAPI    *log.Logger
	WarningAPI *log.Logger
	ErrorAPI   *log.Logger
)

func initLogging(logFile *os.File, processType, debug string) {
	Debug = log.New(ioutil.Discard, "["+processType+"] DEBUG: ", logFlags)
	if debug == "true" {
		Debug = log.New(io.MultiWriter(logFile, os.Stdout), "["+processType+"] DEBUG: ", logFlags)
	}
	Info = log.New(io.MultiWriter(logFile, os.Stdout), "["+processType+"] INFO: ", logFlags)
	Warning = log.New(io.MultiWriter(logFile, os.Stdout), "["+processType+"] WARNING: ", logFlags)
	Error = log.New(io.MultiWriter(logFile, os.Stderr), "["+processType+"] ERROR: ", logFlags)

	DebugAPI = log.New(ioutil.Discard, "[API] DEBUG: ", logFlags)
	if debug == "true" {
		DebugAPI = log.New(io.MultiWriter(logFile, os.Stdout), "[API] DEBUG: ", logFlags)
	}
	InfoAPI = log.New(io.MultiWriter(logFile, os.Stdout), "[API] INFO: ", logFlags)
	WarningAPI = log.New(io.MultiWriter(logFile, os.Stdout), "[API] WARNING: ", logFlags)
	ErrorAPI = log.New(io.MultiWriter(logFile, os.Stderr), "[API] ERROR: ", logFlags)
}
