package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type GitClientMock struct {
	mock.Mock
}

func (_m *GitClientMock) ListTags() ([]string, error) {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

var Tags = []string{
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

func TestNewSemVer(t *testing.T) {
	v, err := NewSemVer("1.1.0")
	assert.NoError(t, err)

	assert.Equal(t, "1.1.0", v.String())
}

func TestNewSemVerNoPatch(t *testing.T) {
	v, err := NewSemVer("1.0")
	assert.NoError(t, err)

	assert.Equal(t, "1.0.0", v.String())
}

func TestNewSemVerPre(t *testing.T) {
	v, err := NewSemVer("1.0-SNAPSHOT")
	assert.NoError(t, err)

	assert.Equal(t, "1.0.0-SNAPSHOT", v.String())
}

func TestMajorMinorEqual(t *testing.T) {
	v1, err := NewSemVer("1.2.0")
	assert.NoError(t, err)

	v2, err := NewSemVer("1.2.3")
	assert.NoError(t, err)

	assert.True(t, MajorMinorEqual(v1, v2))
}

func TestMajorMinorNotEqual(t *testing.T) {
	v1, err := NewSemVer("1.2.0")
	assert.NoError(t, err)

	v2, err := NewSemVer("1.3.0")
	assert.NoError(t, err)

	assert.False(t, MajorMinorEqual(v1, v2))
}

func TestGetLatestVersion(t *testing.T) {
	r := NewRelVer{
		dir: "examples",
	}

	mockClient := &GitClientMock{}
	mockClient.On("ListTags").Return(Tags, nil)

	v, err := r.GetLatestVersion(mockClient)
	assert.NoError(t, err)

	assert.Equal(t, "99.0.17", v.String())
}

func TestGetLatestVersionNoTags(t *testing.T) {
	r := NewRelVer{
		dir: "examples",
	}

	mockClient := &GitClientMock{}
	mockClient.On("ListTags").Return([]string{}, nil)

	v, _ := r.GetLatestVersion(mockClient)
	assert.NotNil(t, v)

	assert.Equal(t, "0.0.0", v.String())
}

// The latest tag in GitHub and locally should be equal, granted the user hasn't added a new tag.
func TestGetLatestVersionGitHub(t *testing.T) {
	r := NewRelVer{
		dir: ".",
	}

	gitHubClient := NewGitHubClient("trendmicro", "new-release-version", r.debug)
	githubVersion, err := r.GetLatestVersion(gitHubClient)
	assert.NoError(t, err)

	localGitClient := NewLocalGitClient(".", true /*fetch*/, r.debug)
	localVersion, err := r.GetLatestVersion(localGitClient)
	assert.NoError(t, err)

	assert.Equal(t, githubVersion, localVersion)
}

func TestGetNewVersion(t *testing.T) {
	r := NewRelVer{
		dir: "examples",
	}

	mockClient := &GitClientMock{}
	mockClient.On("ListTags").Return(Tags, nil)

	v, err := r.GetNewVersion(mockClient)
	assert.NoError(t, err)

	assert.Equal(t, "99.0.18", v.String())
}

func TestGetNewVersionNoTags(t *testing.T) {
	r := NewRelVer{
		dir: "examples",
	}

	mockClient := &GitClientMock{}
	mockClient.On("ListTags").Return([]string{}, nil)

	v, _ := r.GetNewVersion(mockClient)
	assert.NotNil(t, v)

	assert.Equal(t, "0.0.1", v.String())
}

func TestGetNewMinorVersion(t *testing.T) {
	r := NewRelVer{
		dir:   "examples",
		minor: true,
	}

	mockClient := &GitClientMock{}
	mockClient.On("ListTags").Return(Tags, nil)

	v, err := r.GetNewVersion(mockClient)
	assert.NoError(t, err)

	assert.Equal(t, "99.1.0", v.String())
}

func TestGetBaseVersionNoVersionFile(t *testing.T) {
	r := NewRelVer{
		dir: "examples",
	}

	v, _ := r.GetBaseVersion()

	assert.Equal(t, "0.0.0", v.String())
}

func TestVersionsGradle(t *testing.T) {
	r := NewRelVer{
		dir: "examples/java/versions.gradle",
	}

	v, err := r.GetBaseVersion()
	assert.NoError(t, err)

	assert.Equal(t, "1.2.3", v.String())
}

func TestBuildGradle(t *testing.T) {
	r := NewRelVer{
		dir: "examples/java/build.gradle",
	}

	v, err := r.GetBaseVersion()
	assert.NoError(t, err)

	assert.Equal(t, "1.2.3-SNAPSHOT", v.String())
}

func TestPomXML(t *testing.T) {
	r := NewRelVer{
		dir: "examples/java/pom.xml",
	}

	v, err := r.GetBaseVersion()
	assert.NoError(t, err)

	assert.Equal(t, "1.0.0-SNAPSHOT", v.String())
}

func TestBuildGradleKTS(t *testing.T) {
	r := NewRelVer{
		dir: "examples/kotlin",
	}

	v, err := r.GetBaseVersion()
	assert.NoError(t, err)

	assert.Equal(t, "1.2.3", v.String())
}

func TestPackageJSON(t *testing.T) {
	r := NewRelVer{
		dir: "examples/nodejs",
	}

	v, err := r.GetBaseVersion()
	assert.NoError(t, err)

	assert.Equal(t, "1.2.3", v.String())
}

func TestSetupCfg(t *testing.T) {
	r := NewRelVer{
		dir: "examples/python/setup.cfg",
	}

	v, err := r.GetBaseVersion()
	assert.NoError(t, err)

	assert.Equal(t, "1.2.3", v.String())
}

func TestSetupPy(t *testing.T) {
	r := NewRelVer{
		dir: "examples/python/setup.py",
	}

	v, err := r.GetBaseVersion()
	assert.NoError(t, err)

	assert.Equal(t, "4.5.6", v.String())
}

func TestSetupPyNested(t *testing.T) {
	r := NewRelVer{
		dir: "examples/python/setup.py/nested",
	}

	v, err := r.GetBaseVersion()
	assert.NoError(t, err)

	assert.Equal(t, "4.5.6", v.String())
}

func TestSetupPyOneLine(t *testing.T) {
	r := NewRelVer{
		dir: "examples/python/setup.py/one_line",
	}

	v, err := r.GetBaseVersion()
	assert.NoError(t, err)

	assert.Equal(t, "4.5.6", v.String())
}

func TestMakefile(t *testing.T) {
	r := NewRelVer{
		dir: "examples/make",
	}

	v, err := r.GetBaseVersion()
	assert.NoError(t, err)

	assert.Equal(t, "99.0.0-SNAPSHOT", v.String())
}

func TestCMakefile(t *testing.T) {
	r := NewRelVer{
		dir: "examples/cmake",
	}

	v, err := r.GetBaseVersion()
	assert.NoError(t, err)

	assert.Equal(t, "1.2.0-SNAPSHOT", v.String())
}
