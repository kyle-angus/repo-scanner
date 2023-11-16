# Repo Scanner

A tool to recurse through directories and output the status of any git repos.

## Install

`$ git clone git@github.com:kyle-angus/repo-scanner.git`

`$ cd repo-scanner`

`$ go install`

## Build

`$ go build -o ./bin/repo-scanner .`

## Todo

- Handle case where the repo doesn't have a remote, which currently results in an error
- Handle case where the repo doesn't have any commits yet, which currently results in an error
- Support submodules, which currently will result in an error for the parent repo
