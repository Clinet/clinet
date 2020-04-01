package main

import (
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/mitchellh/go-ps"
)

var sig = syscall.SIGINT
var watchdogDelay = (5 * time.Second)

func doWatchdog() {
	log.Trace("--- doWatchdog() ---")

	log.Debug("Spawning initial bot process")
	botPID := spawnBot()
	log.Debug("Creating signal channel")
	sc := make(chan os.Signal, 1)
	log.Debug("Registering notification for ", sig, " on signal channel")
	signal.Notify(sc, sig)
	log.Debug("Creating watchdog ticker for ", watchdogDelay)
	watchdogTicker := time.Tick(watchdogDelay)

	for {
		select {
		case _, ok := <-sc:
			if ok {
				log.Debug("Finding process for bot PID ", botPID)
				botProcess, _ := os.FindProcess(botPID)
				log.Debug("Sending ", sig, " signal to bot process")
				_ = botProcess.Signal(sig)
				log.Debug("Waiting for bot process to exit gracefully")
				waitProcess(botPID)
				log.Debug("Exiting")
				os.Exit(0)
			}
		case <- watchdogTicker:
			if !isProcessRunning(botPID) {
				botPID = spawnBot()
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
					if err == nil {
						oldProcess.Signal(syscall.SIGKILL)
					}
				}
			}
		}
	}

	botProcess := exec.Command(os.Args[0], "--isBot", "--watchdogPID", strconv.Itoa(os.Getpid()), "--verbosity", strconv.Itoa(verbosity))
	botProcess.Stdout = os.Stdout
	botProcess.Stderr = os.Stderr

	log.Debug("Spawning bot process")
	err := botProcess.Start()
	if err != nil {
		panic(err)
	}
	log.Debug("Created bot process with PID ", botProcess.Process.Pid)

	return botProcess.Process.Pid
}

func isProcessRunning(pid int) bool {
	log.Trace("--- isProcessRunning(pid: ", pid, ") ---")

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	if runtime.GOOS != "windows" {
		return process.Signal(syscall.Signal(0)) == nil
	}

	processState, err := process.Wait()
	if err != nil {
		return false
	}
	if processState.Exited() {
		return false
	}

	return true
}

func waitProcess(pid int) {
	log.Trace("--- waitProcess(pid: ", pid, ") ---")

	process, err := os.FindProcess(pid)
	if err != nil {
		return
	}
	_, _ = process.Wait()
}