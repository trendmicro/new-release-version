package main

import (
	"errors"
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
	samerelease bool
	minor       bool
	debug       bool
}

// Converts a version string into a semver.Version struct through a go-version.Version struct
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
	if newVersion == nil {
		return nil, err
	}

	// Increment version
	if r.minor {
		newVersion.BumpMinor()
	} else {
		newVersion.BumpPatch()
	}

	// Check if major or minor version has been changed
	// If there is no base version to check, the just return the new version
	baseVersion, err := r.GetBaseVersion()
	if err != nil {
		return newVersion, nil
	}

	if baseVersion.Compare(*newVersion) > 0 {
		return baseVersion, nil
	} else {
		return newVersion, nil
	}
}

func (r NewRelVer) GetLatestVersion(gitClient GitClient) (*semver.Version, error) {
	// Get base version from file, will fallback to 0.0.0 if not found.
	baseVersion, err := r.GetBaseVersion()
	if err != nil && r.debug {
		fmt.Printf("%v\n", err)
	}
	if r.debug {
		fmt.Printf("base version: %v\n", baseVersion)
	}

	// Get all tags from git repo
	tags, err := gitClient.ListTags()
	if err != nil {
		return baseVersion, err
	}
	if r.debug {
		fmt.Printf("found tags: %v\n", tags)
	}
	// if no tags exist then lets start at base version
	if len(tags) == 0 {
		return baseVersion, errors.New("No existing tags found")
	}

	// Find and sort the version tags
	var versions []*semver.Version
	for _, t := range tags {
		if v, _ := NewSemVer(t); v != nil {
			// if same-release argument is set work only with versions which Major and Minor versions are the same
			if r.samerelease && !MajorMinorEqual(baseVersion, v) {
				continue
			}
			versions = append(versions, v)
		}
	}
	if r.debug {
		fmt.Printf("found versions: %v\n", versions)
	}
	// if no version tags exist then lets start at base version
	if len(versions) == 0 {
		return baseVersion, errors.New("No version tags found")
	}
	semver.Sort(versions)

	return versions[len(versions)-1], nil
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
	return &semver.Version{}, errors.New("No version file found")
}

func (r NewRelVer) FindVersionFile(f string) ([]byte, error) {
	data, err := ioutil.ReadFile(filepath.Join(r.dir, f))
	if err == nil && r.debug {
		fmt.Printf("found %s\n", f)
	}
	return data, err
}
