package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const latestReleaseURL = "https://api.github.com/repos/Huijiro/YaYeet/releases/latest"

type Release struct {
	Version             string
	PageURL             string
	AppImageURL         string
	AppImageChecksumURL string
}

type githubRelease struct {
	TagName    string        `json:"tag_name"`
	HTMLURL    string        `json:"html_url"`
	Draft      bool          `json:"draft"`
	Prerelease bool          `json:"prerelease"`
	Assets     []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func Check(ctx context.Context, currentVersion string) (Release, bool, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	return check(ctx, client, latestReleaseURL, currentVersion)
}

func check(ctx context.Context, client *http.Client, endpoint, currentVersion string) (Release, bool, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Release{}, false, fmt.Errorf("create latest release request: %w", err)
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "YaYeet updater")

	response, err := client.Do(request)
	if err != nil {
		return Release{}, false, fmt.Errorf("fetch latest release: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return Release{}, false, fmt.Errorf("fetch latest release: unexpected HTTP status %s", response.Status)
	}

	var latest githubRelease
	if err := json.NewDecoder(response.Body).Decode(&latest); err != nil {
		return Release{}, false, fmt.Errorf("decode latest release: %w", err)
	}
	if latest.Draft || latest.Prerelease {
		return Release{}, false, nil
	}

	comparison, err := compareVersions(latest.TagName, currentVersion)
	if err != nil {
		return Release{}, false, err
	}

	release := Release{Version: strings.TrimPrefix(latest.TagName, "v"), PageURL: latest.HTMLURL}
	appImageName := "YaYeet-" + latest.TagName + "-x86_64.AppImage"
	for _, asset := range latest.Assets {
		switch asset.Name {
		case appImageName:
			release.AppImageURL = asset.BrowserDownloadURL
		case appImageName + ".sha256":
			release.AppImageChecksumURL = asset.BrowserDownloadURL
		}
	}
	return release, comparison > 0, nil
}

func compareVersions(left, right string) (int, error) {
	leftParts, err := parseVersion(left)
	if err != nil {
		return 0, fmt.Errorf("parse release version: %w", err)
	}
	rightParts, err := parseVersion(right)
	if err != nil {
		return 0, fmt.Errorf("parse current version: %w", err)
	}
	for index := range leftParts {
		if leftParts[index] < rightParts[index] {
			return -1, nil
		}
		if leftParts[index] > rightParts[index] {
			return 1, nil
		}
	}
	return 0, nil
}

func parseVersion(version string) ([3]int, error) {
	var parsed [3]int
	parts := strings.Split(strings.TrimPrefix(version, "v"), ".")
	if len(parts) != len(parsed) {
		return parsed, fmt.Errorf("invalid version %q", version)
	}
	for index, part := range parts {
		value, err := strconv.Atoi(part)
		if err != nil || value < 0 {
			return parsed, fmt.Errorf("invalid version %q", version)
		}
		parsed[index] = value
	}
	return parsed, nil
}
