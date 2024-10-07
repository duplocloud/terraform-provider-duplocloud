package duplocloud

import (
	"regexp"
	"strconv"
	"strings"
)

func extractVersion(input string) string {
	// Regular expression to match version-like patterns (e.g., 3.07.1, 4.01.2)
	re := regexp.MustCompile(`\d+\.\d+\.\d+`)

	// Find the first version number match
	version := re.FindString(input)

	return version
}

func compareEngineVersion(version1, version2 string) int {

	v1Parts := strings.Split(extractVersion(version1), ".")
	v2Parts := strings.Split(version2, ".")
	// Get the maximum length to compare
	maxLen := max(len(v1Parts), len(v2Parts))

	// Iterate over each part of the version (major, minor, patch, etc.)
	for i := 0; i < maxLen; i++ {
		v1 := getVersionPart(v1Parts, i)
		v2 := getVersionPart(v2Parts, i)

		if v1 > v2 {
			return 1
		} else if v1 < v2 {
			return -1
		}
	}

	// If all parts are equal, return 0
	return 0
}

// getVersionPart retrieves the version part at a specific index, returns 0 if the part doesn't exist.
func getVersionPart(versionParts []string, index int) int {
	if index < len(versionParts) {
		part, err := strconv.Atoi(versionParts[index])
		if err != nil {
			return 0 // Fallback to 0 if conversion fails
		}
		return part
	}
	return 0 // Return 0 if the index is out of bounds (e.g., comparing "1.2" with "1.2.0")
}

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
