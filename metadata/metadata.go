package metadata

import (
	"fmt"
	"github.com/blang/semver/v4"
)

var (
	versionMain       = "0.0.1"
	versionPreRelease = "-snapshot"
	versionBuild      = ""
	Version           = semver.MustParse(fmt.Sprintf("%s%s%s", versionMain, versionPreRelease, versionBuild))
	VersionRaw        = Version.String()
)
