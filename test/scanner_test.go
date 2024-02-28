package main

import (
	"github.com/kyle-angus/local-repo-scanner/cli"
)

type MockCli struct {
}

func NewMockCli() *cli.RepoScanner {
	return &cli.RepoScanner{}
}

func (cli *MockCli) storeResult() {}

func (cli *MockCli) parsePathArg() string {
	return ""
}

func (cli *MockCli) hasGitRepo() bool {
	return false
}

func (cli *MockCli) childContainsGitRepos() bool {
	return false
}

func (cli *MockCli) printResults() {}

func (cli *MockCli) walk() error {
	return nil
}

func (cli *MockCli) Run() {}
