package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/coreos/go-semver/semver"
	"github.com/hashicorp/go-version"
)

// NewRelVer is the release version config.
type NewRelVer struct {
	dir         string
	baseVersion string
	sameRelease bool
	minor       bool
	debug       bool
}

// NewSemVer converts a version string into a semver.Version struct through a go-version.Version struct
// This is done because go-version is more lenient with version strings while semver can actually increment version numbers.
func NewSemVer(v string) (*semver.Version, error) {
	ver, err := version.NewVersion(v)
	if err != nil {
		return nil, err
	}
	return semver.NewVersion(ver.String())
}

func (r NewRelVer) GetNewVersion(gitClient GitClient) (*semver.Version, error) {
	newVersion, err := r.GetLatestVersion(gitClient)
	if err != nil {
		return nil, err
	}

	// Increment version
	if r.minor {
		newVersion.BumpMinor()
	} else {
		newVersion.BumpPatch()
	}

	return newVersion, nil
}

func (r NewRelVer) GetLatestVersion(gitClient GitClient) (*semver.Version, error) {
	baseVersion, err := r.GetBaseVersion()
	if err != nil {
		return nil, err
	}

	// Get all tags from git
	tags, err := gitClient.ListTags()
	if err != nil {
		return nil, err
	}
	if r.debug {
		fmt.Printf("found tags: %v\n", tags)
	}
	if len(tags) == 0 {
		if r.debug {
			fmt.Println("No existing tags found")
		}
		return baseVersion, nil
	}

	// Find and sort the version tags
	var versions []*semver.Version
	for _, t := range tags {
		if v, _ := NewSemVer(t); v != nil {
			if r.sameRelease && !MajorMinorEqual(baseVersion, v) {
				continue
			}
			versions = append(versions, v)
		}
	}
	if r.debug {
		fmt.Printf("found versions: %v\n", versions)
	}
	if len(versions) == 0 {
		if r.debug {
			fmt.Println("No version tags found")
		}
		return baseVersion, nil
	}

	semver.Sort(versions)
	latestVersion := versions[len(versions)-1]

	// Return latest version unless base version is higher
	if baseVersion.Compare(*latestVersion) > 0 {
		return baseVersion, nil
	}
	return latestVersion, nil
}

func MajorMinorEqual(v1, v2 *semver.Version) bool {
	return v1.Major == v2.Major && v1.Minor == v2.Minor
}

func (r NewRelVer) GetBaseVersion() (*semver.Version, error) {
	if r.baseVersion != "" {
		return NewSemVer(r.baseVersion)
	}
	for verFile, verFunc := range versionFiles {
		if file, err := r.FindVersionFile(verFile); err == nil {
			if v, err := verFunc(file); err == nil {
				return NewSemVer(v)
			} else if r.debug {
				fmt.Printf("%v\n", err)
			}
		}
	}
	if r.debug {
		fmt.Println("No version file found")
	}
	return &semver.Version{}, nil
}

func (r NewRelVer) FindVersionFile(f string) ([]byte, error) {
	data, err := ioutil.ReadFile(filepath.Join(r.dir, f))
	if err == nil && r.debug {
		fmt.Printf("found %s\n", f)
	}
	return data, err
}
