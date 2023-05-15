package procfs

import (
	"os"
	"regexp"
	"strings"
)

type versionParts struct {
	Major       string `json:"major"`
	Minor       string `json:"minor"`
	BuildNumber string `json:"build_number"`
}

type Distribution struct {
	Name         string       `json:"name"`
	Id           string       `json:"id"`
	Version      string       `json:"version"`
	VersionParts versionParts `json:"version_parts"`
	Like         string       `json:"like"`
	Codename     string       `json:"codename"`
}

var (
	prettyNameRegex      = regexp.MustCompile(`(?m)^PRETTY_NAME="?(.+?)"?$`)
	idRegex              = regexp.MustCompile(`(?m)^ID="?(.+?)"?$`)
	versionIdRegex       = regexp.MustCompile(`(?m)^VERSION_ID="?(.+?)"?$`)
	idLikeRegex          = regexp.MustCompile(`(?m)^ID_LIKE="?(.+?)"?$`)
	versionCodenameRegex = regexp.MustCompile(`(?m)^VERSION_CODENAME="?(.+?)"?$`)
)

func GetDistribution() (*Distribution, error) {
	return getDistributionImpl("/etc/os-release")
}

func getDistributionImpl(osReleasePath string) (*Distribution, error) {

	info := &Distribution{}

	osReleaseBytes, err := os.ReadFile(osReleasePath)
	if err != nil {
		return info, err
	}
	osRelease := string(osReleaseBytes)

	if match := prettyNameRegex.FindStringSubmatch(osRelease); match != nil {
		info.Name = match[1]
	}

	if match := idRegex.FindStringSubmatch(osRelease); match != nil {
		info.Id = match[1]
	}

	if match := versionIdRegex.FindStringSubmatch(osRelease); match != nil {
		info.Version = match[1]
		parts := strings.Split(info.Version, ".")
		info.VersionParts.Major = parts[0]
		if len(parts) > 1 {
			info.VersionParts.Minor = parts[1]
		}
		if len(parts) > 2 {
			info.VersionParts.BuildNumber = parts[2]
		}
	}

	if match := idLikeRegex.FindStringSubmatch(osRelease); match != nil {
		info.Like = match[1]
	}

	if match := versionCodenameRegex.FindStringSubmatch(osRelease); match != nil {
		info.Codename = match[1]
	}

	return info, nil
}
