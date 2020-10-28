package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trendmicro/new-release-version/adapters"
	"github.com/trendmicro/new-release-version/mocks"
)

func TestVersionsGradle(t *testing.T) {
	r := NewRelVer{
		dir: "test-resources/java/versions.gradle",
	}
	v, err := r.getVersion()

	assert.NoError(t, err)

	assert.Equal(t, "1.2.3", v, "error with getVersion for a versions.gradle")
}

func TestBuildGradle(t *testing.T) {
	r := NewRelVer{
		dir: "test-resources/java/build.gradle",
	}
	v, err := r.getVersion()

	assert.NoError(t, err)

	assert.Equal(t, "1.2.3-SNAPSHOT", v, "error with getVersion for a build.gradle")
}

func TestPomXML(t *testing.T) {
	r := NewRelVer{
		dir: "test-resources/java/pom.xml",
	}
	v, err := r.getVersion()

	assert.NoError(t, err)

	assert.Equal(t, "1.0-SNAPSHOT", v, "error with getVersion for a pom.xml")
}

func TestBuildGradleKTS(t *testing.T) {
	r := NewRelVer{
		dir: "test-resources/kotlin",
	}
	v, err := r.getVersion()

	assert.NoError(t, err)

	assert.Equal(t, "1.2.3", v, "error with getVersion for a build.gradle.kts")
}

func TestPackageJSON(t *testing.T) {
	r := NewRelVer{
		dir: "test-resources/nodejs",
	}
	v, err := r.getVersion()

	assert.NoError(t, err)

	assert.Equal(t, "1.2.3", v, "error with getVersion for a package.json")
}

func TestSetupCfg(t *testing.T) {

	r := NewRelVer{
		dir: "test-resources/python/setup.cfg",
	}
	v, err := r.getVersion()

	assert.NoError(t, err)

	assert.Equal(t, "1.2.3", v, "error with getVersion for a setup.cfg")
}

func TestSetupPy(t *testing.T) {

	r := NewRelVer{
		dir: "test-resources/python/setup.py",
	}
	v, err := r.getVersion()

	assert.NoError(t, err)

	assert.Equal(t, "4.5.6", v, "error with getVersion for a setup.py")
}

func TestSetupPyNested(t *testing.T) {

	r := NewRelVer{
		dir: "test-resources/python/setup.py/nested",
	}
	v, err := r.getVersion()

	assert.NoError(t, err)

	assert.Equal(t, "4.5.6", v, "error with getVersion for a nested setup.py")
}

func TestSetupPyOneLine(t *testing.T) {

	r := NewRelVer{
		dir: "test-resources/python/setup.py/one_line",
	}
	v, err := r.getVersion()

	assert.NoError(t, err)

	assert.Equal(t, "4.5.6", v, "error with getVersion for a oneliner setup.py")
}

func TestMakefile(t *testing.T) {
	r := NewRelVer{
		dir: "test-resources/make",
	}

	v, err := r.getVersion()

	assert.NoError(t, err)

	assert.Equal(t, "99.0.0-SNAPSHOT", v, "error with getVersion for a Makefile")
}

func TestCMakefile(t *testing.T) {

	r := NewRelVer{
		dir: "test-resources/cmake",
	}

	v, err := r.getVersion()

	assert.NoError(t, err)

	assert.Equal(t, "1.2.0-SNAPSHOT", v, "error with getVersion for a CMakeLists.txt")
}

func TestGetNewPatchVersion(t *testing.T) {

	r := NewRelVer{}

	tags := createTags()

	mockClient := &mocks.GitClient{}
	mockClient.On("ListTags").Return(tags, nil)

	v, err := r.getNewVersionFromTag(mockClient)

	assert.NoError(t, err)
	assert.Equal(t, "99.0.18", v, "error bumping a patch version")
}

func TestGetNewMinorVersion(t *testing.T) {

	r := NewRelVer{
		minor: true,
	}

	tags := createTags()

	mockClient := &mocks.GitClient{}
	mockClient.On("ListTags").Return(tags, nil)

	v, err := r.getNewVersionFromTag(mockClient)

	assert.NoError(t, err)
	assert.Equal(t, "99.1.0", v, "error bumping a minor version")
}

// The latest tag in GitHub and locally should be equal, granted the user hasn't added a new tag.
func TestGetLatestTag(t *testing.T) {
	r := NewRelVer{}

	gitHubClient := adapters.NewGitHubClient("trendmicro", "new-release-version", r.debug)
	githubVersion, err := r.getLatestTag(gitHubClient)
	assert.NoError(t, err)

	localGitClient := adapters.NewLocalGitClient(".", r.debug)
	localVersion, err := r.getLatestTag(localGitClient)
	assert.NoError(t, err)
	assert.Equal(t, githubVersion, localVersion, "error with getLatestTag")
}

func createTags() []string {
	return []string{
		"v99.0.0",
		"v99.0.1",
		"v99.0.10",
		"v99.0.11",
		"v99.0.12",
		"v99.0.13",
		"v99.0.14",
		"v99.0.15",
		"v99.0.16",
		"v99.0.17",
		"v99.0.2",
		"v99.0.3",
		"v99.0.4",
		"v99.0.5",
		"v99.0.6",
		"v99.0.7",
		"v99.0.8",
		"v99.0.9",
	}
}
