package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Version is the version of new-release-version.  This is set by goreleaser.
var Version = "latest"

func main() {

	dir := flag.String("directory", ".", "Directory of git project.")
	baseVersion := flag.String("base-version", "", "Version to use instead of version file.")
	sameRelease := flag.Bool("same-release", false, "Support older releases: for example 7.0.x and tag for new release 7.1.x already exist, with `-same-release` argument next version from 7.0.x will be returned.")
	minor := flag.Bool("minor", false, "Increment minor version instead of patch.")
	fetch := flag.Bool("git-fetch", true, "Fetch tags from remote.")
	owner := flag.String("gh-owner", "", "GitHub repository owner to fetch tags from instead of the local git repo.")
	repo := flag.String("gh-repository", "", "GitHub repository to fetch tags from instead of the local git repo.")
	debug := flag.Bool("debug", false, "Prints debug into to console.")
	ver := flag.Bool("version", false, "Prints the version.")
	flag.Parse()

	if *ver {
		fmt.Println("new-release-version", Version)
		os.Exit(0)
	}

	r := NewRelVer{
		dir:         *dir,
		baseVersion: *baseVersion,
		sameRelease: *sameRelease,
		minor:       *minor,
		debug:       *debug,
	}

	if r.debug {
		fmt.Println("environment:")
		for _, e := range os.Environ() {
			fmt.Println(e)
		}
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

type findVersion func([]byte) (string, error)

const versionRegex = `[\.\d]+(-\w+)?`

var versionFiles = map[string]findVersion{
	"versions.gradle":  versionMatcher(fmt.Sprintf(`(?m)project\.version\s*=\s*['"](%s)['"]$`, versionRegex), 1),
	"build.gradle":     versionMatcher(fmt.Sprintf(`(?m)^version\s*=\s*['"](%s)['"]$`, versionRegex), 1),
	"build.gradle.kts": versionMatcher(fmt.Sprintf(`(?m)^version\s*=\s*['"](%s)['"]$`, versionRegex), 1),
	"pom.xml":          unmarshalXMLVersion,
	"package.json":     unmarshalJSONVersion,
	"setup.cfg":        versionMatcher(fmt.Sprintf(`(?m)^version\s*=\s*(%s)$`, versionRegex), 1),
	"setup.py":         versionMatcher(fmt.Sprintf(`(?ms)setup\(.*\s+version\s*=\s*['"](%s)['"].*\)$`, versionRegex), 1),
	"CMakeLists.txt":   versionMatcher(fmt.Sprintf(`(?ms)^project\s*\(.*\s+VERSION\s+(%s).*\)$`, versionRegex), 1),
	"Makefile":         versionMatcher(fmt.Sprintf(`(?m)^VERSION\s*:=\s*(%s)$`, versionRegex), 1),
}

func versionMatcher(regex string, group int) findVersion {
	return func(file []byte) (string, error) {
		return matchVersion(file, regex, group)
	}
}

func matchVersion(data []byte, regex string, group int) (string, error) {
	re := regexp.MustCompile(regex)
	matched := re.FindSubmatch(data)
	if len(matched) > 0 {
		version := strings.TrimSpace(string(matched[group]))
		return version, nil
	}
	return "0.0.0", errors.New("No version found")
}

func unmarshalJSONVersion(data []byte) (string, error) {
	var project struct {
		Version string `json:"version"`
	}
	json.Unmarshal(data, &project)
	if project.Version != "" {
		return project.Version, nil
	}
	return "0.0.0", errors.New("No version found")
}

func unmarshalXMLVersion(data []byte) (string, error) {
	var project struct {
		Version string `xml:"version"`
	}
	xml.Unmarshal(data, &project)
	if project.Version != "" {
		return project.Version, nil
	}
	return "0.0.0", errors.New("No version found")
}
