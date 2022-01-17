package metadata

import (
	"fmt"
	"github.com/blang/semver/v4"
)

var (
	versionMain       = "1.0.2"
	versionPreRelease = "snapshot"
	versionBuild      = ""
	Version           = makeVersion(versionMain, versionPreRelease, versionBuild)
	VersionRaw        = Version.String()
)

func makeVersion(main string, prerelease string, build string) semver.Version {
	if prerelease != "" {
		prerelease = fmt.Sprintf("-%s", prerelease)
	}
	if build != "" {
		build = fmt.Sprintf("+%s", build)
	}
	return semver.MustParse(fmt.Sprintf("%s%s%s", main, prerelease, build))
}
