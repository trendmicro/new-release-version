package main

import (
	"flag"
	"fmt"
	"os"
)

// These are set by goreleaser during release.
var (
	version = "latest"
	commit  = "main"
	date    = "???"
	builtBy = "you"
)

func main() {

	dir := flag.String("directory", ".", "Directory of git project.")
	baseVersion := flag.String("base-version", "", "Version to use instead of version file.")
	sameRelease := flag.Bool("same-release", false, "Increment the latest base version release ignoring any releases higher than the base version release.")
	minor := flag.Bool("minor", false, "Increment minor version instead of patch.")
	fetch := flag.Bool("git-fetch", true, "Fetch tags from remote.")
	owner := flag.String("gh-owner", "", "GitHub repository owner to fetch tags from instead of the local git repo.")
	repo := flag.String("gh-repository", "", "GitHub repository to fetch tags from instead of the local git repo.")
	debug := flag.Bool("debug", false, "Prints debug into to console.")
	ver := flag.Bool("version", false, "Prints the version.")
	flag.Parse()

	if *debug {
		fmt.Println("version:", version)
		fmt.Println("commit:", commit)
		fmt.Println("date:", date)
		fmt.Println("built by:", builtBy)
		fmt.Println("environment:")
		for _, e := range os.Environ() {
			fmt.Println(e)
		}
	}

	if *ver {
		fmt.Println("new-release-version", version)
		os.Exit(0)
	}

	r := NewRelVer{
		dir:         *dir,
		baseVersion: *baseVersion,
		sameRelease: *sameRelease,
		minor:       *minor,
		debug:       *debug,
	}

	var gitClient GitClient
	if *owner != "" && *repo != "" {
		gitClient = NewGitHubClient(*owner, *repo, r.debug)
	} else {
		gitClient = NewLocalGitClient(r.dir, *fetch, r.debug)
	}

	v, err := r.GetNewVersion(gitClient)
	if err != nil {
		fmt.Printf("failed to get new version: %v\n", err)
		os.Exit(-1)
	}
	fmt.Print(v.String())
}
