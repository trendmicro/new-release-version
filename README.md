# New Release Version

Forked from https://github.com/jenkins-x-plugins/jx-release-version

![Build](https://github.com/trendmicro/new-release-version/workflows/Build/badge.svg)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/trendmicro/new-release-version/)](https://pkg.go.dev/github.com/trendmicro/new-release-version/)

new-release-version is a simple command that will print a new patch (or minor) version of a release.

This helps in continuous delivery if you want an automatic release when a change is merged to your main branch.  Traditional approaches mean the version is stored in a file that is checked and updated after each release.  If you want automatic releases this means you will get another release triggered from the version update resulting in a cyclic release sitiation.  

Using a git tag to work out the next release version is better than traditional approaches of storing it in a VERSION file or updating a project's config file.

The major and minor version of the release is determined by searching for a version identifier in a project config file, like package.json or build.gradle.
See [examples](examples) for examples of supported version files.

If you need to bump the major or minor version of your project, simply increment the version in your project's config file and commit that to your main branch.
You can also use the `-base-version` option to set a base version.

See `new-release-version -help` or the [go docs](https://pkg.go.dev/github.com/trendmicro/new-release-version/) for more information.

## Install

You must have [Git](https://git-scm.com/) installed on your system in order for `new-release-version` to work.

You can install the latest from the `main` branch

    go get github.com/trendmicro/new-release-version

Or install a specific version from [releases](https://github.com/trendmicro/new-release-version/releases/)

## Examples

```sh
    ➜ RELEASE_VERSION=$(new-release-version)
    ➜ echo "New release version ${RELEASE_VERSION}
    ➜ git tag -fa v${RELEASE_VERSION} -m 'Release version ${RELEASE_VERSION}'
    ➜ git push origin v${RELEASE_VERSION}
```

- If your project is new or has no existing git tags then running `new-release-version` will return a default version of `0.0.1`

- If your latest git tag is `1.2.3` and your version file is `1.2.0-SNAPSHOT` then `new-release-version` will return `1.2.4`

- If your latest git tag is `1.2.3` and your version file is `2.0.0` then `new-release-version` will return `2.0.0`

- If your latest git tag is `7.1.0` but you want to increment an older release, say `7.0.5`, use `new-release-version -base-version 7.0 -same-release` to return the next version in the `7.0` release.

## Development

### Prereqs

- [Go 1.15+](https://go.dev/)
- [Git](https://git-scm.com/)

### Build

    make build

### Test

    make test
