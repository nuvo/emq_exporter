package main

import "fmt"

var (
	// GitTag stands for a git tag, populated at build time
	GitTag string
	// GitCommit stands for a git commit hash populated at build time
	GitCommit string
)

func printVersion() string {
	return fmt.Sprintf("Version %s (git-%s)", GitTag, GitCommit)
}
