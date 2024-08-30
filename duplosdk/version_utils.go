package duplosdk

import (
	"strconv"
	"strings"
)

type AppVersion struct {
	Major int
	Minor int
}

func ParseAppVersion(version string) AppVersion {
	parts := strings.Split(version, ".")
	if len(parts) < 1 {
		return AppVersion{Major: 0, Minor: 0}
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		major = 0
	}

	minor := 0
	if len(parts) > 1 {
		minor, err = strconv.Atoi(parts[1])
		if err != nil {
			minor = 0
		}
	}

	return AppVersion{Major: major, Minor: minor}
}

func IsAppVersionEqualOrGreater(versionCurrent, versionCompareTo string) bool {
	if len(versionCurrent) == 0 || len(versionCompareTo) == 0 {
		return false
	}
	verCurrent := ParseAppVersion(versionCurrent)
	verCompareTo := ParseAppVersion(versionCompareTo)
	if verCurrent.Major > verCompareTo.Major {
		return true
	} else if verCurrent.Major == verCompareTo.Major && verCurrent.Minor >= verCompareTo.Minor {
		return true
	}
	return false
}
