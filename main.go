/*
new-release-version is a simple command that will print a new patch (or minor) version of a release.

This helps in continuous delivery if you want an automatic release when a change is merged to your main branch.  Traditional approaches mean the version is
stored in a file that is checked and updated after each release.  If you want automatic releases this means you will get another release triggered from the
version update resulting in a cyclic release sitiation.
*/
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

	var gitClient GitClient
	if *owner != "" && *repo != "" {
		gitClient = NewGitHubClient(*owner, *repo, *debug)
	} else {
		gitClient = NewLocalGitClient(*dir, *fetch, *debug)
	}

	r := NewRelVer{
		Dir:         *dir,
		BaseVersion: *baseVersion,
		SameRelease: *sameRelease,
		Minor:       *minor,
		Debug:       *debug,
	}

	v, err := r.GetNewVersion(gitClient)
	if err != nil {
		fmt.Printf("failed to get new version: %v\n", err)
		os.Exit(-1)
	}
	fmt.Print(v.String())
}
