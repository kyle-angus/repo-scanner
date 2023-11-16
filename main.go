package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type Result struct {
	path      string
	status    string
	colorFunc func(a ...interface{}) string
}

var storedResults []Result

func storeResult(path, status string, colorFunc func(a ...interface{}) string) {
	result := Result{
		path:      path,
		status:    status,
		colorFunc: colorFunc,
	}
	storedResults = append(storedResults, result)
}

func parsePathArg() string {
	args := os.Args
	if len(args) <= 1 {
		return ""
	} else {
		return args[1]
	}
}

func main() {
	var path string
	showHelp := flag.Bool("help", false, "Show help information")
	showHelpShort := flag.Bool("h", false, "Show help information")
	path = parsePathArg()
	flag.Parse()

	if *showHelp || *showHelpShort || path == "" {
		fmt.Println("\nUsage: repo-scanner PATH")
		fmt.Println("\n  A tool to recurse through directories and output the status of any git repos.")
		fmt.Println("\nFlags:")
		fmt.Println("  -h, --h, -help, --help\t Show help information")
		os.Exit(0)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	walkFunction := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && info.Name() == "node_modules" {
			return filepath.SkipDir
		}

		if info.IsDir() {
			if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
				repo, err := git.PlainOpen(path)
				if err != nil {
					storeResult(path, "Error", color.New(color.FgRed).SprintFunc())
					return filepath.SkipDir
				}

				ref, err := repo.Head()
				if err != nil {
					storeResult(path, "Error", color.New(color.FgRed).SprintFunc())
					return filepath.SkipDir
				}

				localRef, err := repo.Reference(ref.Name(), true)
				if err != nil {
					storeResult(path, "Error", color.New(color.FgRed).SprintFunc())
					return filepath.SkipDir
				}

				remoteRef, err := repo.Reference(plumbing.NewRemoteReferenceName("origin", "master"), true)
				if err != nil {
					remoteRef, err = repo.Reference(plumbing.NewRemoteReferenceName("origin", "main"), true)
					if err != nil {
						storeResult(path, "Error", color.New(color.FgRed).SprintFunc())
						return filepath.SkipDir
					}
				}

				localCommit, err := repo.CommitObject(localRef.Hash())
				if err != nil {
					storeResult(path, "Error", color.New(color.FgRed).SprintFunc())
					return filepath.SkipDir
				}

				remoteCommit, err := repo.CommitObject(remoteRef.Hash())
				if err != nil {
					storeResult(path, "Error", color.New(color.FgRed).SprintFunc())
					return filepath.SkipDir
				}

				isLocalAhead := localCommit.Committer.When.After(remoteCommit.Committer.When)
				isLocalBehind := localCommit.Committer.When.Before(remoteCommit.Committer.When)

				if isLocalAhead {
					storeResult(path, "Not Synced", color.New(color.FgYellow).SprintFunc())
				} else if isLocalBehind {
					storeResult(path, "Not Synced", color.New(color.FgYellow).SprintFunc())
				} else {
					storeResult(path, "Synced", color.New(color.FgGreen).SprintFunc())
				}
			}
		}

		return nil
	}

	err = filepath.Walk(absPath, walkFunction)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	for _, result := range storedResults {
		fmt.Printf("%s: %s\n", color.New(color.FgBlue).SprintFunc()(result.path), result.colorFunc(result.status))
	}
}
