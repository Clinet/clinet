package main

import (
	"runtime"
)

var (
	//Variables filled in on compile time using github.com/JoshuaDoes/govvv
	GitBranch     string
	GitCommit     string
	GitCommitFull string
	GitCommitMsg  string
	GitState      string
	BuildDate     string

	//A unique build ID inspired by the Android Open Source Project
	BuildID string = "clinet-" + GitState + " " + GitBranch + "-" + GitCommit

	//The URL to the current commit
	GitHubCommitURL string = "https://github.com/JoshuaDoes/clinet/commit/" + GitCommitFull

	//The version of Go used to build this release
	GolangVersion string = runtime.Version()
)
