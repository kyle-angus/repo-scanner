package main

import "github.com/kyle-angus/local-repo-scanner/scanner"

func main() {
	cli := scanner.NewRepoScanner()
	cli.Run()
}
