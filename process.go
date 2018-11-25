package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"

	"github.com/mitchellh/go-ps"
)

func spawnBot() int {
	if killOldBot == "true" {
		processList, err := ps.Processes()
		if err == nil {
			for _, process := range processList {
				if process.Pid() != os.Getpid() && process.Pid() != masterPID && process.Executable() == filepath.Base(os.Args[0]) {
					oldProcess, err := os.FindProcess(process.Pid())
					if err == nil {
						oldProcess.Signal(syscall.SIGKILL)
					}
				}
			}
		}
	}
	os.Remove(os.Args[0] + ".old")

	botProcess := exec.Command(os.Args[0], "-bot", "true", "-masterpid", strconv.Itoa(os.Getpid()), "-debug", debug)
	botProcess.Stdout = os.Stdout
	botProcess.Stderr = os.Stderr
	err := botProcess.Start()
	if err != nil {
		panic(err)
	}
	return botProcess.Process.Pid
}

func isProcessRunning(pid int) bool {
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
	process, err := os.FindProcess(pid)
	if err != nil {
		return
	}
	_, _ = process.Wait()
}
