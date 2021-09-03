package main

import (
	//Process management
	"github.com/mitchellh/go-ps"

	//std necessities
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

var sig1 = syscall.SIGINT
var sig2 = syscall.SIGKILL
var watchdogDelay = (5 * time.Second)

func doWatchdog() {
	log.Trace("--- doWatchdog() ---")

	log.Debug("Spawning initial bot process")
	botPID := spawnBot()
	log.Debug("Creating signal channel")
	sc := make(chan os.Signal, 1)
	log.Debug("Registering notification for ", sig1, " on signal channel")
	signal.Notify(sc, sig1)
	log.Debug("Registering notification for ", sig2, " on signal channel")
	signal.Notify(sc, sig2)
	log.Debug("Creating watchdog ticker for ", watchdogDelay)
	watchdogTicker := time.Tick(watchdogDelay)

	for {
		select {
		case <-watchdogTicker:
			if !isProcessRunning(botPID) {
				log.Debug("Failed to find process for bot PID ", botPID, ", spawning new bot process")
				botPID = spawnBot()
			}
		case sig, ok := <-sc:
			if ok {
				log.Trace("SIGNAL: ", sig)
				if isProcessRunning(botPID) {
					log.Debug("Finding process for bot PID ", botPID)
					botProcess, err := os.FindProcess(botPID)
					if err == nil {
						log.Debug("Sending ", sig, " signal to bot process")
						_ = botProcess.Signal(sig)
						log.Debug("Waiting for bot process to exit gracefully")
						waitProcess(botPID)
					}
				}
				log.Debug("Exiting")
				os.Exit(0)
			}
		}
	}
}

func spawnBot() int {
	log.Trace("--- spawnBot() ---")

	if killOldBot {
		processList, err := ps.Processes()
		if err == nil {
			for _, process := range processList {
				if process.Pid() != os.Getpid() && process.Pid() != watchdogPID && process.Executable() == filepath.Base(os.Args[0]) {
					oldProcess, err := os.FindProcess(process.Pid())
					if oldProcess != nil && err == nil {
						oldProcess.Signal(syscall.SIGKILL)
					}
				}
			}
		}
	}

	botProcess := exec.Command(os.Args[0], "--isBot", "--watchdogPID", strconv.Itoa(os.Getpid()), "--verbosity", strconv.Itoa(verbosity), "--config", config)
	botProcess.Stdout = os.Stdout
	botProcess.Stderr = os.Stderr
	botProcess.Stdin = os.Stdin

	log.Debug("Spawning bot process")
	err := botProcess.Start()
	if err != nil {
		panic(err)
	}
	log.Debug("Created bot process with PID ", botProcess.Process.Pid)

	return botProcess.Process.Pid
}

func isProcessRunning(pid int) bool {
	//log.Trace("--- isProcessRunning(pid: ", pid, ") ---")

	process, err := ps.FindProcess(pid)
	if err != nil {
		log.Error(err)
		return false
	}

	return process != nil
}

func waitProcess(pid int) {
	log.Trace("--- waitProcess(pid: ", pid, ") ---")

	process, err := os.FindProcess(pid)
	if err != nil {
		return
	}
	_, _ = process.Wait()
}
