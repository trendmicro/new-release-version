package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/coreos/go-semver/semver"
	goVersion "github.com/hashicorp/go-version"
)

// VersionNumberRegex is the regex used to find a version number.
const VersionNumberRegex = `[\.\d]+(-\w+)?`

// Version identifier regex strings for version files.
// The %s is replaced with VersionNumberRegex.
const (
	VersionsGradleRegexf = `(?m)project\.version\s*=\s*['"](%s)['"]$`
	BuildGradleRegexf    = `(?m)^version\s*=\s*['"](%s)['"]$`
	SetupCfgRegexf       = `(?m)^version\s*=\s*(%s)$`
	SetupPyRegexf        = `(?ms)setup\(.*\s+version\s*=\s*['"](%s)['"].*\)$`
	CMakeListsTxtRegexf  = `(?ms)^project\s*\(.*\s+VERSION\s+(%s).*\)$`
	MakefileRegexf       = `(?m)^VERSION\s*:=\s*(%s)$`
)

type findVersion func([]byte) (string, error)

var versionFiles = map[string]findVersion{
	"versions.gradle":  versionMatcher(VersionsGradleRegexf, 1),
	"build.gradle":     versionMatcher(BuildGradleRegexf, 1),
	"build.gradle.kts": versionMatcher(BuildGradleRegexf, 1),
	"pom.xml":          unmarshalXMLVersion,
	"package.json":     unmarshalJSONVersion,
	"setup.cfg":        versionMatcher(SetupCfgRegexf, 1),
	"setup.py":         versionMatcher(SetupPyRegexf, 1),
	"CMakeLists.txt":   versionMatcher(CMakeListsTxtRegexf, 1),
	"Makefile":         versionMatcher(MakefileRegexf, 1),
}

func versionMatcher(regexf string, group int) findVersion {
	return func(file []byte) (string, error) {
		regex := fmt.Sprintf(regexf, VersionNumberRegex)
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

// MajorMinorEqual returns true if v1 and v2 share the same major and minor version numbers; false otherwise.
func MajorMinorEqual(v1, v2 *semver.Version) bool {
	return v1.Major == v2.Major && v1.Minor == v2.Minor
}

// NewSemVer converts a version string into a semver.Version struct through a go-version.Version struct.
//
// This is done because go-version is more lenient with version strings while semver can actually increment version numbers.
func NewSemVer(v string) (*semver.Version, error) {
	ver, err := goVersion.NewVersion(v)
	if err != nil {
		return nil, err
	}
	return semver.NewVersion(ver.String())
}

// NewRelVer is the release version config.
type NewRelVer struct {
	Dir         string
	BaseVersion string
	SameRelease bool
	Minor       bool
	Debug       bool
}

// GetNewVersion returns an incremented version number based on the current latest version.
//
// E.g.
//
// - If the latest version is 1.2.0 then 1.2.1 will be returned (or 1.3.0 if NewRelVer.minor is set to true).
//
// - If a project has no previous versions but has set a base version of 1.0, then 1.0.0 is returned.
//
// - For projects that have no previous versions or base version, then 0.0.1 is returned (or 0.1.0 if NewRelVer.minor is set to true).
func (r NewRelVer) GetNewVersion(gitClient GitClient) (*semver.Version, error) {
	newVersion, baseVersion, err := r.GetLatestVersion(gitClient)
	if err != nil {
		return nil, err
	}

	if newVersion == nil {
		// Return the new base version as is unless it is 0.0.0, in which case we should increment to 0.0.1
		if !baseVersion.Equal(semver.Version{}) {
			return baseVersion, nil
		}
		newVersion = baseVersion
	}

	// Increment version
	if r.Minor {
		newVersion.BumpMinor()
	} else {
		newVersion.BumpPatch()
	}

	return newVersion, nil
}

// GetLatestVersion returns the project's latest known version and base version.
//
// The latest version is found by looking at the project's base version and git tags and returning the highest version number from those.
//
// E.g.
//
// - If the base version is 1.0 and the highest git tag is 1.1.0, then 1.1.0 will be returned.
//
// - Vice versa, if the base version is 1.2 and the highest git tag is 1.1.0, then 1.2.0 will be returned.
//
// - If there are no git tags and no base version, then 0.0.0 will be returned.
//
// Note the base version is always returned (even if it is 0.0.0) unless there is an error.
func (r NewRelVer) GetLatestVersion(gitClient GitClient) (latest, base *semver.Version, err error) {
	baseVersion, err := r.GetBaseVersion()
	if err != nil {
		return nil, nil, err
	}

	// Get all tags from git
	tags, err := gitClient.ListTags()
	if err != nil {
		return nil, nil, err
	}
	if r.Debug {
		fmt.Printf("found tags: %v\n", tags)
	}
	if len(tags) == 0 {
		return nil, baseVersion, nil
	}

	// Find and sort the version tags
	var versions []*semver.Version
	for _, t := range tags {
		if v, _ := NewSemVer(t); v != nil {
			if r.SameRelease && !MajorMinorEqual(baseVersion, v) {
				continue
			}
			versions = append(versions, v)
		}
	}
	if r.Debug {
		fmt.Printf("found versions: %v\n", versions)
	}
	if len(versions) == 0 {
		return nil, baseVersion, nil
	}
	semver.Sort(versions)
	latestVersion := versions[len(versions)-1]

	// Return latest version unless base version is higher
	if baseVersion.Compare(*latestVersion) > 0 {
		return nil, baseVersion, nil
	}
	return latestVersion, baseVersion, nil
}

// GetBaseVersion returns the project's base version.
//
// The base version is found by searching a known set of project config files for a known version identifier.
//
// E.g.
//
// - If the project config file sets a version 1.0, then 1.0.0 is returned.
//
// - If no project config file is found then 0.0.0 is returned.
//
// - If NewRelVer.baseVersion is set, then that version is returned.
//
// WARNING: GetBaseVersion does not search for project config files in a deterministic order, so if you have more than one supported project config file in your
// repo, make sure only one has a version identifier.
func (r NewRelVer) GetBaseVersion() (*semver.Version, error) {
	if r.BaseVersion != "" {
		return NewSemVer(r.BaseVersion)
	}
	for verFile, verFunc := range versionFiles {
		if file, err := r.FindVersionFile(verFile); err == nil {
			if v, err := verFunc(file); err == nil {
				return NewSemVer(v)
			} else if r.Debug {
				fmt.Printf("%v\n", err)
			}
		}
	}
	if r.Debug {
		fmt.Println("No version file found")
	}
	return &semver.Version{}, nil
}

// FindVersionFile returns the contents of the given file from NewRelVer.dir directory.
func (r NewRelVer) FindVersionFile(f string) ([]byte, error) {
	data, err := ioutil.ReadFile(filepath.Join(r.Dir, f))
	if err == nil && r.Debug {
		fmt.Printf("found %s\n", f)
	}
	return data, err
}
