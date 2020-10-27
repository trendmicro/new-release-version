package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trendmicro/new-release-version/adapters"
	"github.com/trendmicro/new-release-version/domain"
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

	assert.Equal(t, "1.0-SNAPSHOT", v, "error with getVersion for a build.gradle")
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

	assert.Equal(t, "4.5.6", v, "error with getVersion for a setup.py")
}

func TestSetupPyOneLine(t *testing.T) {

	r := NewRelVer{
		dir: "test-resources/python/setup.py/one_line",
	}
	v, err := r.getVersion()

	assert.NoError(t, err)

	assert.Equal(t, "4.5.6", v, "error with getVersion for a setup.py")
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

func TestGetNewVersionFromTagCurrentRepo(t *testing.T) {
	r := NewRelVer{
		dryrun: false,
		dir:    "test-resources/make",
	}

	tags := createTags()

	mockClient := &mocks.GitClient{}
	mockClient.On("ListTags", context.Background(), r.ghOwner, r.ghRepository).Return(tags, nil)
	v, err := r.getNewVersionFromTag(mockClient)

	assert.NoError(t, err)
	assert.Equal(t, "99.0.0", v, "error bumping a patch version")
}

func TestGetGitTag(t *testing.T) {
	r := NewRelVer{
		ghOwner:      "trendmicro",
		ghRepository: "new-release-version",
	}

	gitHubClient := adapters.NewGitHubClient(r.debug)

	expectedVersion, err := r.getLatestTag(gitHubClient)
	assert.NoError(t, err)

	r = NewRelVer{}

	v, err := r.getLatestTag(gitHubClient)

	assert.NoError(t, err)

	assert.Equal(t, expectedVersion, v, "error with getLatestTag for a Makefile")
}

func TestGetNewMinorVersionFromGitHubTag(t *testing.T) {

	r := NewRelVer{
		ghOwner:      "trendmicro",
		ghRepository: "new-release-version",
		minor:        true,
	}

	tags := createTags()

	mockClient := &mocks.GitClient{}
	mockClient.On("ListTags", context.Background(), r.ghOwner, r.ghRepository).Return(tags, nil)

	v, err := r.getNewVersionFromTag(mockClient)

	assert.NoError(t, err)
	assert.Equal(t, "99.1.0", v, "error bumping a minor version")
}

func TestGetNewPatchVersionFromGitHubTag(t *testing.T) {

	r := NewRelVer{
		ghOwner:      "trendmicro",
		ghRepository: "new-release-version",
	}

	tags := createTags()

	mockClient := &mocks.GitClient{}
	mockClient.On("ListTags", context.Background(), r.ghOwner, r.ghRepository).Return(tags, nil)

	v, err := r.getNewVersionFromTag(mockClient)

	assert.NoError(t, err)
	assert.Equal(t, "99.0.18", v, "error bumping a patch version")
}

func createTags() []domain.Tag {
	var tags []domain.Tag
	tags = append(tags, domain.Tag{Name: "v99.0.0"})
	tags = append(tags, domain.Tag{Name: "v99.0.1"})
	tags = append(tags, domain.Tag{Name: "v99.0.2"})
	tags = append(tags, domain.Tag{Name: "v99.0.3"})
	tags = append(tags, domain.Tag{Name: "v99.0.4"})
	tags = append(tags, domain.Tag{Name: "v99.0.5"})
	tags = append(tags, domain.Tag{Name: "v99.0.6"})
	tags = append(tags, domain.Tag{Name: "v99.0.7"})
	tags = append(tags, domain.Tag{Name: "v99.0.8"})
	tags = append(tags, domain.Tag{Name: "v99.0.9"})
	tags = append(tags, domain.Tag{Name: "v99.0.10"})
	tags = append(tags, domain.Tag{Name: "v99.0.11"})
	tags = append(tags, domain.Tag{Name: "v99.0.12"})
	tags = append(tags, domain.Tag{Name: "v99.0.13"})
	tags = append(tags, domain.Tag{Name: "v99.0.14"})
	tags = append(tags, domain.Tag{Name: "v99.0.15"})
	tags = append(tags, domain.Tag{Name: "v99.0.16"})
	tags = append(tags, domain.Tag{Name: "v99.0.17"})

	return tags
}
