package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/hashicorp/go-version"

	"github.com/trendmicro/new-release-version/adapters"
)

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

// NewRelVer is the release version config.
type NewRelVer struct {
	debug       bool
	dir         string
	samerelease bool
	baseVersion string
	minor       bool
}

// Version is the version of new-release-version.  This is set by goreleaser.
var Version = "dev"

func main() {

	dir := flag.String("directory", ".", "Directory of git project.")
	baseVersion := flag.String("base-version", "", "Version to use instead of version file.")
	samerelease := flag.Bool("same-release", false, "Support older releases: for example 7.0.x and tag for new release 7.1.x already exist, with `-same-release` argument next version from 7.0.x will be returned.")
	minor := flag.Bool("minor", false, "Increment minor version instead of patch.")
	owner := flag.String("gh-owner", "", "GitHub repository owner to fetch tags from instead of the local git repo.")
	repo := flag.String("gh-repository", "", "GitHub repository to fetch tags from instead of the local git repo.")
	debug := flag.Bool("debug", false, "Prints debug into to console")
	ver := flag.Bool("version", false, "Prints the version")
	flag.Parse()

	if *ver {
		fmt.Printf("new-release-version %s\n", Version)
		os.Exit(0)
	}

	r := NewRelVer{
		debug:       *debug,
		dir:         *dir,
		samerelease: *samerelease,
		baseVersion: *baseVersion,
		minor:       *minor,
	}

	if r.debug {
		fmt.Println("available environment:")
		for _, e := range os.Environ() {
			fmt.Println(e)
		}
	}

	var gitClient adapters.GitClient
	if *owner != "" && *repo != "" {
		gitClient = adapters.NewGitHubClient(*owner, *repo, r.debug)
	} else {
		gitClient = adapters.NewLocalGitClient(r.dir, r.debug)
	}

	v, err := r.getNewVersionFromTag(gitClient)
	if err != nil {
		fmt.Println("failed to get new version", err)
		os.Exit(-1)
	}
	fmt.Print(v)
}

func (r NewRelVer) getNewVersionFromTag(gitClient adapters.GitClient) (string, error) {

	tag, err := r.getLatestTag(gitClient)
	if err != nil && tag == "" {
		return "", err
	}
	sv, err := semver.NewVersion(tag)
	if err != nil {
		return "", err
	}

	if r.minor {
		sv.BumpMinor()
	} else {
		sv.BumpPatch()
	}

	majorVersion := sv.Major
	minorVersion := sv.Minor
	patchVersion := sv.Patch

	// check if major or minor version has been changed
	baseVersion, err := r.getVersion()
	if err != nil {
		return fmt.Sprintf("%d.%d.%d", majorVersion, minorVersion, patchVersion), nil
	}

	// first use go-version to turn into a proper version, this handles 1.0-SNAPSHOT which semver doesn't
	tmpVersion, err := version.NewVersion(baseVersion)
	if err != nil {
		return fmt.Sprintf("%d.%d.%d", majorVersion, minorVersion, patchVersion), nil
	}
	bsv, err := semver.NewVersion(tmpVersion.String())
	if err != nil {
		return "", err
	}
	baseMajorVersion := bsv.Major
	baseMinorVersion := bsv.Minor
	basePatchVersion := bsv.Patch

	if baseMajorVersion > majorVersion ||
		(baseMajorVersion == majorVersion &&
			(baseMinorVersion > minorVersion) || (baseMinorVersion == minorVersion && basePatchVersion > patchVersion)) {
		majorVersion = baseMajorVersion
		minorVersion = baseMinorVersion
		patchVersion = basePatchVersion
	}
	return fmt.Sprintf("%d.%d.%d", majorVersion, minorVersion, patchVersion), nil
}

func (r NewRelVer) getLatestTag(gitClient adapters.GitClient) (string, error) {
	// Get base version from file, will fallback to 0.0.0 if not found.
	baseVersion, err := r.getVersion()
	if err != nil && r.debug {
		fmt.Printf("%v\n", err)
	}
	if r.debug {
		fmt.Printf("base version: %s\n", baseVersion)
	}

	tags, err := gitClient.ListTags()
	if err != nil {
		return "", err
	}
	if len(tags) == 0 {
		// if no tags exist then lets start at base version
		return baseVersion, errors.New("No existing tags found")
	}
	if r.debug {
		fmt.Printf("found tags: %v\n", tags)
	}

	// turn tags into a new collection of versions that we can sort
	var versions []*version.Version
	for _, t := range tags {
		// if same-release argument is set work only with versions which Major and Minor versions are the same
		if r.samerelease {
			same, _ := isMajorMinorTheSame(baseVersion, t)
			if same {
				v, _ := version.NewVersion(t)
				if v != nil {
					versions = append(versions, v)
				}
			}
		} else {
			v, _ := version.NewVersion(t)
			if v != nil {
				versions = append(versions, v)
			}
		}
	}

	if len(versions) == 0 {
		// if no version tags exist then lets start at base version
		return baseVersion, errors.New("No version tags found")
	}

	// return the latest tag
	col := version.Collection(versions)
	if r.debug {
		fmt.Printf("found versions: %v\n", col)
	}

	sort.Sort(col)
	latest := len(versions)
	if versions[latest-1] == nil {
		return baseVersion, errors.New("No latest version found")
	}
	return versions[latest-1].String(), nil
}

func (r NewRelVer) getVersion() (string, error) {
	if r.baseVersion != "" {
		return r.baseVersion, nil
	}
	for verFile, verFunc := range versionFiles {
		if file, err := r.findVersionFile(verFile); err == nil {
			if v, err := verFunc(file); err == nil {
				return v, nil
			} else if r.debug {
				fmt.Printf("%v\n", err)
			}
		}
	}
	return "0.0.0", errors.New("No version file found")
}

func (r NewRelVer) findVersionFile(f string) ([]byte, error) {
	data, err := ioutil.ReadFile(filepath.Join(r.dir, f))
	if err == nil && r.debug {
		fmt.Printf("found %s\n", f)
	}
	return data, err
}

func isMajorMinorTheSame(v1 string, v2 string) (bool, error) {
	sv1, err1 := semver.NewVersion(v1)
	if err1 != nil {
		return false, err1
	}
	sv2, err2 := semver.NewVersion(v2)
	if err2 != nil {
		return false, err2
	}
	if sv1.Major != sv2.Major {
		return false, nil
	}
	if sv1.Minor != sv2.Minor {
		return false, nil
	}
	return true, nil
}
