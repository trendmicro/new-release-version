package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/hashicorp/go-version"

	"github.com/trendmicro/new-release-version/adapters"
	"github.com/trendmicro/new-release-version/domain"
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

// NewRelVer is the release version config
type NewRelVer struct {
	dryrun       bool
	debug        bool
	dir          string
	ghOwner      string
	ghRepository string
	samerelease  bool
	baseVersion  string
	minor        bool
}

var Version = "dev"

func main() {

	debug := flag.Bool("debug", false, "prints debug into to console")
	dir := flag.String("directory", ".", "the directory to look for version files with the project version to bump")
	owner := flag.String("gh-owner", "", "a github repository owner if not running from within a git project")
	repo := flag.String("gh-repository", "", "a github repository if not running from within a git project")
	baseVersion := flag.String("base-version", "", "version to use instead of version file")
	samerelease := flag.Bool("same-release", false, "for support old releases: for example 7.0.x and tag for new release 7.1.x already exist, with `-same-release` argument next version from 7.0.x will be returned ")
	ver := flag.Bool("version", false, "prints the version")
	minor := flag.Bool("minor", false, "increment minor version instead of patch")
	flag.Parse()

	if *ver {
		fmt.Printf("new-release-version %s\n", Version)
		os.Exit(0)
	}

	r := NewRelVer{
		debug:        *debug,
		dir:          *dir,
		ghOwner:      *owner,
		ghRepository: *repo,
		samerelease:  *samerelease,
		baseVersion:  *baseVersion,
		minor:        *minor,
	}

	if r.debug {
		fmt.Println("available environment:")
		for _, e := range os.Environ() {
			fmt.Println(e)
		}
	}

	gitHubClient := adapters.NewGitHubClient(r.debug)
	v, err := r.getNewVersionFromTag(gitHubClient)
	if err != nil {
		fmt.Println("failed to get new version", err)
		os.Exit(-1)
	}
	fmt.Print(v)
}

func (r NewRelVer) getNewVersionFromTag(gitClient domain.GitClient) (string, error) {

	// get the latest github tag
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

func (r NewRelVer) getLatestTag(gitClient domain.GitClient) (string, error) {
	// Get base version from file, will fallback to 0.0.0 if not found.
	baseVersion, _ := r.getVersion()

	if r.debug {
		fmt.Printf("base version: %s\n", baseVersion)
	}

	// if repo isn't provided by flags fall back to using current repo if run from a git project
	var versionsRaw []string
	if r.ghOwner != "" && r.ghRepository != "" {
		ctx := context.Background()

		tags, err := gitClient.ListTags(ctx, r.ghOwner, r.ghRepository)

		if err != nil {
			return "", err
		}
		if len(tags) == 0 {
			// if no current flags exist then lets start at base version
			return baseVersion, errors.New("No existing tags found")
		}

		// build an array of all the tags
		versionsRaw = make([]string, len(tags))
		for i, tag := range tags {
			if r.debug {
				fmt.Printf("found remote tag %s\n", tag.Name)
			}
			versionsRaw[i] = tag.Name
		}
	} else {
		_, err := exec.LookPath("git")
		if err != nil {
			return "", fmt.Errorf("error running git: %v", err)
		}
		cmd := exec.Command("git", "fetch", "--tags", "-v")
		cmd.Env = append(cmd.Env, os.Environ()...)
		cmd.Dir = r.dir
		err = cmd.Run()
		if err != nil {
			return "", fmt.Errorf("error fetching tags: %v", err)
		}

		cmd = exec.Command("git", "tag")
		cmd.Dir = r.dir
		out, err := cmd.Output()
		if err != nil {
			return "", err
		}
		str := strings.TrimSuffix(string(out), "\n")
		tags := strings.Split(str, "\n")

		if len(tags) == 0 {
			// if no current flags exist then lets start at base version
			return baseVersion, errors.New("No existing tags found")
		}

		// build an array of all the tags
		versionsRaw = make([]string, len(tags))
		for i, tag := range tags {
			if r.debug {
				fmt.Printf("found tag %s\n", tag)
			}
			tag = strings.TrimPrefix(tag, "v")
			if tag != "" {
				versionsRaw[i] = tag
			}
		}
	}

	// turn the array into a new collection of versions that we can sort
	var versions []*version.Version
	for _, raw := range versionsRaw {
		// if same-release argument is set work only with versions which Major and Minor versions are the same
		if r.samerelease {
			same, _ := isMajorMinorTheSame(baseVersion, raw)
			if same {
				v, _ := version.NewVersion(raw)
				if v != nil {
					versions = append(versions, v)
				}
			}
		} else {
			v, _ := version.NewVersion(raw)
			if v != nil {
				versions = append(versions, v)
			}
		}
	}

	if len(versions) == 0 {
		// if no current flags exist then lets start at base version
		return baseVersion, errors.New("No existing tags found")
	}

	// return the latest tag
	col := version.Collection(versions)
	if r.debug {
		fmt.Printf("version collection %v\n", col)
	}

	sort.Sort(col)
	latest := len(versions)
	if versions[latest-1] == nil {
		return baseVersion, errors.New("No existing tags found")
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
