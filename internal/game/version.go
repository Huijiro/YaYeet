package game

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
)

const (
	manifestURL = "https://votv.dev/patcher_assets/patch_manifest.json"
	catalogURL  = "https://votv.dev/patcher_assets/index_manifest.json"
)

type VersionOption struct {
	Name              string
	Label             string
	RevisionlessLabel string
	URL               string
	Unstable          bool
	Test              bool
}

type Status struct {
	ExecutablePath string
	PakPath        string
	Hash           string
	Installed      string
	Latest         string
	UpdateReady    bool
}

type manifest struct {
	FileHashMap      map[string]string `json:"FileHashMap"`
	LowerHashMap     map[string]string `json:"fileHashMap"`
	LatestStable     string            `json:"LatestStable"`
	LatestUnstable   string            `json:"LatestUnstable"`
	UnstableVersions []string          `json:"UnstableVersions"`
	Latest           string            `json:"latest"`
	Patches          map[string]patch  `json:"Patches"`
	LowerPatches     map[string]patch  `json:"patches"`
}

type patch struct {
	URL    string `json:"Url"`
	SHA256 string `json:"Sha256"`
}

type catalogEntry struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
	Link string `json:"link"`
}

func AvailableVersions(ctx context.Context) ([]VersionOption, string, error) {
	currentManifest, err := fetchManifest(ctx)
	if err != nil {
		return nil, "", err
	}

	catalog, err := fetchCatalog(ctx)
	if err != nil {
		return nil, "", err
	}

	unstableVersions := make(map[string]struct{}, len(currentManifest.UnstableVersions))
	for _, version := range currentManifest.UnstableVersions {
		unstableVersions[version] = struct{}{}
	}

	versions := make([]VersionOption, 0, len(catalog))
	selected := ""
	for _, entry := range catalog {
		if entry.Name == "" {
			continue
		}

		versionName, catalogTag, tagged := strings.Cut(entry.Name, " [")
		_, unstable := unstableVersions[versionName]
		unstable = unstable || strings.HasPrefix(catalogTag, "UNSTABLE]")
		test := strings.HasPrefix(catalogTag, "TEST]")
		label := displayVersionName(versionName)
		if versionName == currentManifest.LatestStable || versionName == currentManifest.LatestUnstable {
			label += " (latest)"
		}
		if unstable {
			label += " (unstable)"
		} else if tagged {
			label += " (" + strings.ToLower(strings.TrimSuffix(catalogTag, "]")) + ")"
		} else {
			label += " (stable)"
		}
		if versionName == currentManifest.LatestStable {
			selected = label
		}
		revisionlessLabel := label
		if base, _, ok := revisionParts(versionName); ok {
			revisionlessLabel = base + strings.TrimPrefix(label, displayVersionName(versionName))
		}
		versions = append(versions, VersionOption{
			Name:              versionName,
			Label:             label,
			RevisionlessLabel: revisionlessLabel,
			URL:               entry.Link,
			Unstable:          unstable,
			Test:              test,
		})
	}
	sort.Slice(versions, func(i, j int) bool { return versions[i].Label > versions[j].Label })
	return versions, selected, nil
}

func LatestRevisions(versions []VersionOption) []VersionOption {
	latestByBase := make(map[string][]int)
	for _, version := range versions {
		base, revision, ok := revisionParts(version.Name)
		if !ok {
			continue
		}
		if latest, exists := latestByBase[base]; !exists || newerRevision(revision, latest) {
			latestByBase[base] = revision
		}
	}

	filtered := make([]VersionOption, 0, len(versions))
	for _, version := range versions {
		base, revision, ok := revisionParts(version.Name)
		if !ok || slices.Equal(revision, latestByBase[base]) {
			filtered = append(filtered, version)
		}
	}
	return filtered
}

func revisionParts(version string) (string, []int, bool) {
	parts := strings.Split(version, "_")
	if len(parts) < 2 || parts[0] == "" {
		return "", nil, false
	}

	revision := make([]int, 0, len(parts)-1)
	for _, part := range parts[1:] {
		value, err := strconv.Atoi(part)
		if err != nil {
			return "", nil, false
		}
		revision = append(revision, value)
	}
	return parts[0], revision, true
}

func newerRevision(candidate, current []int) bool {
	for index := 0; index < max(len(candidate), len(current)); index++ {
		var candidatePart, currentPart int
		if index < len(candidate) {
			candidatePart = candidate[index]
		}
		if index < len(current) {
			currentPart = current[index]
		}
		if candidatePart != currentPart {
			return candidatePart > currentPart
		}
	}
	return false
}

func displayVersionName(version string) string {
	base, revision, ok := revisionParts(version)
	if !ok {
		return version
	}

	parts := make([]string, len(revision))
	for index, value := range revision {
		parts[index] = strconv.Itoa(value)
	}
	return base + "-" + strings.Join(parts, ".")
}

func Detect(ctx context.Context, installationPath string) (Status, error) {
	executablePath := filepath.Join(installationPath, "WindowsNoEditor", "VotV.exe")
	if _, err := os.Stat(executablePath); err != nil {
		return Status{}, fmt.Errorf("find VotV.exe: %w", err)
	}

	pakPath := filepath.Join(installationPath, "WindowsNoEditor", "VotV", "Content", "Paks", "VotV-WindowsNoEditor.pak")
	if _, err := os.Stat(pakPath); err != nil {
		return Status{}, fmt.Errorf("find game pak: %w", err)
	}

	hash, err := pakHash(ctx, pakPath)
	if err != nil {
		return Status{}, err
	}

	currentManifest, err := fetchManifest(ctx)
	if err != nil {
		return Status{}, err
	}
	installed := currentManifest.FileHashMap[hash]
	if installed == "" {
		return Status{ExecutablePath: executablePath, PakPath: pakPath, Hash: hash}, fmt.Errorf("pak hash is not recognized: %s", hash)
	}

	return Status{
		ExecutablePath: executablePath,
		PakPath:        pakPath,
		Hash:           hash,
		Installed:      installed,
		Latest:         currentManifest.LatestStable,
		UpdateReady:    installed != currentManifest.LatestStable,
	}, nil
}

func pakHash(ctx context.Context, path string) (string, error) {
	output, err := exec.CommandContext(ctx, "xxhsum", path).Output()
	if err != nil {
		return "", fmt.Errorf("hash pak: %w", err)
	}
	fields := strings.Fields(string(output))
	if len(fields) == 0 {
		return "", errors.New("xxhsum returned no hash")
	}
	return strings.ToUpper(fields[0]), nil
}

func cacheJSON(name string, data []byte) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return
	}
	cacheDir = filepath.Join(cacheDir, "yayeet", "manifests")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(cacheDir, name), data, 0o644)
}

func fetchCatalog(ctx context.Context) ([]catalogEntry, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, catalogURL, nil)
	if err != nil {
		return nil, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fetch install catalog: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("fetch install catalog: HTTP %s", response.Status)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	cacheJSON("catalog.json", data)
	var catalog []catalogEntry
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("decode install catalog: %w", err)
	}
	return catalog, nil
}

func fetchManifest(ctx context.Context) (manifest, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, manifestURL, nil)
	if err != nil {
		return manifest{}, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return manifest{}, fmt.Errorf("fetch patch manifest: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return manifest{}, fmt.Errorf("fetch patch manifest: HTTP %s", response.Status)
	}

	data, err := io.ReadAll(io.LimitReader(response.Body, 4<<20))
	if err != nil {
		return manifest{}, err
	}
	cacheJSON("manifest.json", data)
	var current manifest
	if err := json.Unmarshal(data, &current); err != nil {
		return manifest{}, fmt.Errorf("decode patch manifest: %w", err)
	}
	if current.FileHashMap == nil {
		current.FileHashMap = current.LowerHashMap
	}
	if current.LatestStable == "" {
		current.LatestStable = current.Latest
	}
	if current.Patches == nil {
		current.Patches = current.LowerPatches
	}
	if current.FileHashMap == nil || current.LatestStable == "" {
		return manifest{}, errors.New("patch manifest is missing required fields")
	}
	return current, nil
}
