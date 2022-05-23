package actions

import (
	"encoding/json"
	"log"
	"os"
	"runtime"

	"github.com/blang/semver/v4"
	"github.com/chuxel/devpacks/internal/common/utils"
)

type VersionManifestFile struct {
	Filename        string
	Arch            string
	Platform        string
	PlatformVersion string `json:"platform_version"`
	DownloadUrl     string `json:"download_url"`
}

type VersionManifestEntry struct {
	Version    string
	Stable     bool
	ReleaseUrl string `json:"release_url"`
	Files      []VersionManifestFile
}

type VersionManifest struct {
	Entries []VersionManifestEntry
}

func (manifest VersionManifest) Load(manifestPath string) {
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		log.Fatal("Failed to read manifest. ", err)
	}
	if err := json.Unmarshal(content, &manifest.Entries); err != nil {
		log.Fatal("Failed to unmarshal manifest contents. ", err)
	}
}

func (manifest VersionManifest) FindEntry(version string) VersionManifestEntry {
	for _, entry := range manifest.Entries {
		if entry.Version == version {
			return entry
		}
	}
	log.Fatal("Unable to find entry for version ", version)
	return VersionManifestEntry{}
}

func (manifest VersionManifest) FindVersion(semverRange string, stableOnly bool) string {
	versions := make([]semver.Version, len(manifest.Entries))
	for _, entry := range manifest.Entries {
		if (entry.Stable && stableOnly) || !stableOnly {
			version, err := semver.ParseTolerant(entry.Version)
			if err != nil {
				log.Fatal(err)
			}
			versions = append(versions, version)
		}
	}
	semver.Sort(versions)

	if semverRange != "latest" {
		expectedRange := utils.NewSemverRange(semverRange)
		// Sorted in ascending order, so run through in reverse order to get the latest matching
		for i := len(versions) - 1; i >= 0; i-- {
			nodeVersion := versions[i]
			if expectedRange(nodeVersion) {
				return nodeVersion.FinalizeVersion()
			}
		}
		log.Fatal("Unable to match node version", semverRange)
	}

	return versions[len(versions)-1].FinalizeVersion()
}

func (manifest VersionManifest) FindDownloadUrl(version string) string {
	dlArch := runtime.GOARCH
	if dlArch == "amd64" {
		dlArch = "x64"
	}
	entry := manifest.FindEntry(version)
	for _, file := range entry.Files {
		if file.Arch == dlArch && file.Platform == "linux" {
			// If a PlatformVersion value is set, then the download is specific to a distro version.
			// Since not all are, verify the distro only if PlatformVersion is actually set.
			if file.PlatformVersion != "" {
				osRelease := utils.ReadLinuxDistroInfo()
				if osRelease.Version == file.PlatformVersion || osRelease.VersionId == file.PlatformVersion {
					return file.DownloadUrl
				}
			} else {
				return file.DownloadUrl
			}
		}
	}
	return ""
}

func NewVersionManifest(manifestPath string) VersionManifest {
	manifest := VersionManifest{}
	manifest.Load(manifestPath)
	return manifest
}

func NewVersionManifestFromUrl(url string) VersionManifest {
	content := utils.DownloadBytesFromUrl(url)
	manifest := VersionManifest{}
	if err := json.Unmarshal(content, &manifest.Entries); err != nil {
		log.Fatal("Failed to unmarshal manifest contents. ", err)
	}
	return manifest
}
