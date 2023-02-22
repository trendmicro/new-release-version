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
	"v1.0.0",
	"v1.0.1",
	"v1.0.2",
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
		Dir: "examples",
	}

	mockClient := &GitClientMock{}
	mockClient.On("ListTags").Return(Tags, nil)

	v, b, err := r.GetLatestVersion(mockClient)
	assert.NoError(t, err)

	assert.Equal(t, "0.0.0", b.String())
	assert.Equal(t, "99.0.17", v.String())
}

func TestGetLatestVersionNoTags(t *testing.T) {
	r := NewRelVer{
		Dir: "examples",
	}

	mockClient := &GitClientMock{}
	mockClient.On("ListTags").Return([]string{}, nil)

	v, b, err := r.GetLatestVersion(mockClient)
	assert.NoError(t, err)

	assert.Equal(t, "0.0.0", b.String())
	assert.Nil(t, v)
}

func TestGetLatestVersionInitBase(t *testing.T) {
	r := NewRelVer{
		Dir:         "examples",
		BaseVersion: "1.0",
	}

	mockClient := &GitClientMock{}
	mockClient.On("ListTags").Return([]string{}, nil)

	v, b, err := r.GetLatestVersion(mockClient)
	assert.NoError(t, err)

	assert.Equal(t, "1.0.0", b.String())
	assert.Nil(t, v)
}

// The latest tag in GitHub and locally should be equal, granted the user hasn't added a new tag.
func TestGetLatestVersionGitHub(t *testing.T) {
	r := NewRelVer{
		Dir: ".",
	}

	gitHubClient := NewGitHubClient("trendmicro", "new-release-version", r.Debug)
	ghv, ghb, err := r.GetLatestVersion(gitHubClient)
	assert.NoError(t, err)

	localGitClient := NewLocalGitClient(".", true /*fetch*/, r.Debug)
	v, b, err := r.GetLatestVersion(localGitClient)
	assert.NoError(t, err)

	assert.Equal(t, ghb, b)
	assert.Equal(t, ghv, v)
}

func TestGetNewVersion(t *testing.T) {
	r := NewRelVer{
		Dir: "examples",
	}

	mockClient := &GitClientMock{}
	mockClient.On("ListTags").Return(Tags, nil)

	v, err := r.GetNewVersion(mockClient)
	assert.NoError(t, err)

	assert.Equal(t, "99.0.18", v.String())
}

func TestGetNewVersionNoTags(t *testing.T) {
	r := NewRelVer{
		Dir: "examples",
	}

	mockClient := &GitClientMock{}
	mockClient.On("ListTags").Return([]string{}, nil)

	v, err := r.GetNewVersion(mockClient)
	assert.NoError(t, err)

	assert.Equal(t, "0.0.1", v.String())
}

func TestGetNewVersionInitBaseVersion(t *testing.T) {
	r := NewRelVer{
		Dir:         "examples",
		BaseVersion: "1.0",
	}

	mockClient := &GitClientMock{}
	mockClient.On("ListTags").Return([]string{}, nil)

	v, err := r.GetNewVersion(mockClient)
	assert.NoError(t, err)

	assert.Equal(t, "1.0.0", v.String())
}

func TestGetNewVersionBumpBaseVersion(t *testing.T) {
	r := NewRelVer{
		Dir:         "examples",
		BaseVersion: "100.0.0",
	}

	mockClient := &GitClientMock{}
	mockClient.On("ListTags").Return(Tags, nil)

	v, err := r.GetNewVersion(mockClient)
	assert.NoError(t, err)

	assert.Equal(t, "100.0.0", v.String())
}

func TestGetNewVersionSameRelease(t *testing.T) {
	r := NewRelVer{
		Dir:         "examples",
		BaseVersion: "1.0",
		SameRelease: true,
	}

	mockClient := &GitClientMock{}
	mockClient.On("ListTags").Return(Tags, nil)

	v, err := r.GetNewVersion(mockClient)
	assert.NoError(t, err)

	assert.Equal(t, "1.0.3", v.String())
}

func TestGetNewMinorVersion(t *testing.T) {
	r := NewRelVer{
		Dir:   "examples",
		Minor: true,
	}

	mockClient := &GitClientMock{}
	mockClient.On("ListTags").Return(Tags, nil)

	v, err := r.GetNewVersion(mockClient)
	assert.NoError(t, err)

	assert.Equal(t, "99.1.0", v.String())
}

func TestGetNewMinorVersionNoTags(t *testing.T) {
	r := NewRelVer{
		Dir:   "examples",
		Minor: true,
	}

	mockClient := &GitClientMock{}
	mockClient.On("ListTags").Return([]string{}, nil)

	v, err := r.GetNewVersion(mockClient)
	assert.NoError(t, err)

	assert.Equal(t, "0.1.0", v.String())
}

func TestGetNewMinorVersionInitBase(t *testing.T) {
	r := NewRelVer{
		Dir:         "examples",
		BaseVersion: "100.0.0",
		Minor:       true,
	}

	mockClient := &GitClientMock{}
	mockClient.On("ListTags").Return([]string{}, nil)

	v, err := r.GetNewVersion(mockClient)
	assert.NoError(t, err)

	assert.Equal(t, "100.0.0", v.String())
}

func TestGetNewMinorVersionBumpBase(t *testing.T) {
	r := NewRelVer{
		Dir:         "examples",
		BaseVersion: "100.0.0",
		Minor:       true,
	}

	mockClient := &GitClientMock{}
	mockClient.On("ListTags").Return(Tags, nil)

	v, err := r.GetNewVersion(mockClient)
	assert.NoError(t, err)

	assert.Equal(t, "100.0.0", v.String())
}

func TestGetNewMinorVersionSameRelease(t *testing.T) {
	r := NewRelVer{
		Dir:         "examples",
		BaseVersion: "1.0.0",
		SameRelease: true,
		Minor:       true,
	}

	mockClient := &GitClientMock{}
	mockClient.On("ListTags").Return(Tags, nil)

	v, err := r.GetNewVersion(mockClient)
	assert.NoError(t, err)

	assert.Equal(t, "1.1.0", v.String())
}

func TestGetBaseVersionNoVersionFile(t *testing.T) {
	r := NewRelVer{
		Dir: "examples",
	}

	v, err := r.GetBaseVersion()
	assert.NoError(t, err)

	assert.Equal(t, "0.0.0", v.String())
}

func TestVersionsGradle(t *testing.T) {
	r := NewRelVer{
		Dir: "examples/java/versions.gradle",
	}

	v, err := r.GetBaseVersion()
	assert.NoError(t, err)

	assert.Equal(t, "1.2.3", v.String())
}

func TestGradleProperties(t *testing.T) {
	r := NewRelVer{
		Dir: "examples/java/gradle.properties",
	}

	v, err := r.GetBaseVersion()
	assert.NoError(t, err)

	assert.Equal(t, "1.3.2", v.String())
}

func TestBuildGradle(t *testing.T) {
	r := NewRelVer{
		Dir: "examples/java/build.gradle",
	}

	v, err := r.GetBaseVersion()
	assert.NoError(t, err)

	assert.Equal(t, "1.2.3-SNAPSHOT", v.String())
}

func TestPomXML(t *testing.T) {
	r := NewRelVer{
		Dir: "examples/java/pom.xml",
	}

	v, err := r.GetBaseVersion()
	assert.NoError(t, err)

	assert.Equal(t, "1.0.0-SNAPSHOT", v.String())
}

func TestBuildGradleKTS(t *testing.T) {
	r := NewRelVer{
		Dir: "examples/kotlin",
	}

	v, err := r.GetBaseVersion()
	assert.NoError(t, err)

	assert.Equal(t, "1.2.3", v.String())
}

func TestPackageJSON(t *testing.T) {
	r := NewRelVer{
		Dir: "examples/nodejs",
	}

	v, err := r.GetBaseVersion()
	assert.NoError(t, err)

	assert.Equal(t, "1.2.3", v.String())
}

func TestSetupCfg(t *testing.T) {
	r := NewRelVer{
		Dir: "examples/python/setup.cfg",
	}

	v, err := r.GetBaseVersion()
	assert.NoError(t, err)

	assert.Equal(t, "1.2.3", v.String())
}

func TestSetupPy(t *testing.T) {
	r := NewRelVer{
		Dir: "examples/python/setup.py",
	}

	v, err := r.GetBaseVersion()
	assert.NoError(t, err)

	assert.Equal(t, "4.5.6", v.String())
}

func TestSetupPyNested(t *testing.T) {
	r := NewRelVer{
		Dir: "examples/python/setup.py/nested",
	}

	v, err := r.GetBaseVersion()
	assert.NoError(t, err)

	assert.Equal(t, "4.5.6", v.String())
}

func TestSetupPyOneLine(t *testing.T) {
	r := NewRelVer{
		Dir: "examples/python/setup.py/one_line",
	}

	v, err := r.GetBaseVersion()
	assert.NoError(t, err)

	assert.Equal(t, "4.5.6", v.String())
}

func TestMakefile(t *testing.T) {
	r := NewRelVer{
		Dir: "examples/make",
	}

	v, err := r.GetBaseVersion()
	assert.NoError(t, err)

	assert.Equal(t, "99.0.0-SNAPSHOT", v.String())
}

func TestCMakefile(t *testing.T) {
	r := NewRelVer{
		Dir: "examples/cmake",
	}

	v, err := r.GetBaseVersion()
	assert.NoError(t, err)

	assert.Equal(t, "1.2.0-SNAPSHOT", v.String())
}
