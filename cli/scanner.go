package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type Cli interface {
	storeResults() error
	parsePathArg() string
	hasGitRepo() bool
	childContainsGitRepos() bool
	printResults()
	walk() error
	Run()
}

type RepoScanner struct {
	startPath     string
	absPath       string
	showHelp      *bool
	showHelpShort *bool
	results       []Result
	err           error
}

type Result struct {
	path      string
	status    string
	colorFunc func(a ...interface{}) string
}

func NewRepoScanner() *RepoScanner {
	cli := &RepoScanner{}

	cli.showHelp = flag.Bool("help", false, "Show help information")
	cli.showHelpShort = flag.Bool("h", false, "Show help information")
	cli.startPath = cli.parsePathArg()
	flag.Parse()

	return cli
}

func (cli *RepoScanner) walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if info.IsDir() {
		if filepath.Dir(path) == cli.absPath && !cli.hasGitRepo(path) && !cli.childContainsGitRepos(path) {
			cli.storeResult(path, "No Repo", color.New(color.FgYellow).SprintFunc())
		}

		if info.IsDir() && info.Name() == "node_modules" {
			return filepath.SkipDir
		}

		if cli.hasGitRepo(path) {
			repo, err := git.PlainOpen(path)
			if err != nil {
				cli.storeResult(path, "Error", color.New(color.FgRed).SprintFunc())
				return filepath.SkipDir
			}

			ref, err := repo.Head()
			if err != nil {
				cli.storeResult(path, "No Commits", color.New(color.FgYellow).SprintFunc())
				return filepath.SkipDir
			}

			remotes, err := repo.Remotes()
			if err != nil || len(remotes) == 0 {
				cli.storeResult(path, "No Remote", color.New(color.FgYellow).SprintFunc())
				return filepath.SkipDir
			}

			localRef, err := repo.Reference(ref.Name(), true)
			if err != nil {
				cli.storeResult(path, "Error", color.New(color.FgRed).SprintFunc())
				return filepath.SkipDir
			}

			remoteRef, err := repo.Reference(plumbing.NewRemoteReferenceName("origin", "master"), true)
			if err != nil {
				remoteRef, err = repo.Reference(plumbing.NewRemoteReferenceName("origin", "main"), true)
				if err != nil {
					cli.storeResult(path, "Error", color.New(color.FgRed).SprintFunc())
					return filepath.SkipDir
				}
			}

			localCommit, err := repo.CommitObject(localRef.Hash())
			if err != nil {
				cli.storeResult(path, "Error", color.New(color.FgRed).SprintFunc())
				return filepath.SkipDir
			}

			remoteCommit, err := repo.CommitObject(remoteRef.Hash())
			if err != nil {
				cli.storeResult(path, "Error", color.New(color.FgRed).SprintFunc())
				return filepath.SkipDir
			}

			isLocalAhead := localCommit.Committer.When.After(remoteCommit.Committer.When)
			isLocalBehind := localCommit.Committer.When.Before(remoteCommit.Committer.When)

			if isLocalAhead {
				cli.storeResult(path, "Not Synced", color.New(color.FgYellow).SprintFunc())
			} else if isLocalBehind {
				cli.storeResult(path, "Not Synced", color.New(color.FgYellow).SprintFunc())
			} else {
				cli.storeResult(path, "Synced", color.New(color.FgGreen).SprintFunc())
			}
		}
	}

	return nil
}

func (cli *RepoScanner) Run() {
	if *cli.showHelp || *cli.showHelpShort || cli.startPath == "" {
		fmt.Println("\nUsage: repo-scanner PATH")
		fmt.Println("\n  A tool to recurse through directories and output the status of any git repos.")
		fmt.Println("\nFlags:")
		fmt.Println("  -h, --h, -help, --help\t Show help information")
		os.Exit(0)
	}

	cli.absPath, cli.err = filepath.Abs(cli.startPath)
	if cli.err != nil {
		fmt.Println("Error:", cli.err)
		os.Exit(1)
	}

	cli.err = filepath.Walk(cli.absPath, cli.walk)
	if cli.err != nil {
		fmt.Println("Error:", cli.err)
		os.Exit(1)
	}

	cli.printResults()
}

func (cli *RepoScanner) storeResult(path, status string, colorFunc func(a ...interface{}) string) {
	result := Result{
		path:      path,
		status:    status,
		colorFunc: colorFunc,
	}
	cli.results = append(cli.results, result)
}

func (cli *RepoScanner) parsePathArg() string {
	args := os.Args
	if len(args) <= 1 {
		return ""
	} else {
		return args[1]
	}
}

func (cli *RepoScanner) hasGitRepo(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}

func (cli *RepoScanner) childContainsGitRepos(path string) bool {
	found := false
	filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if found {
			return filepath.SkipAll
		}

		if info.IsDir() && cli.hasGitRepo(subpath) {
			found = true
			return filepath.SkipDir
		}

		return nil
	})

	return found
}

func (cli *RepoScanner) printResults() {
	for _, result := range cli.results {
		fmt.Printf("%s: %s\n", color.New(color.FgBlue).SprintFunc()(result.path), result.colorFunc(result.status))
	}
}
