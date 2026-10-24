package runner

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const protonGELatestReleaseURL = "https://api.github.com/repos/GloriousEggroll/proton-ge-custom/releases/latest"

type ProtonGEProgress struct {
	Phase   string
	Current int64
	Total   int64
}

type ProtonGEProgressFunc func(ProtonGEProgress)

type githubRelease struct {
	TagName    string        `json:"tag_name"`
	Draft      bool          `json:"draft"`
	Prerelease bool          `json:"prerelease"`
	Assets     []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

var protonGEInstallParents = [4]string{
	filepath.Join(".steam", "root"),
	filepath.Join(".steam", "steam"),
	filepath.Join(".local", "share", "Steam"),
	filepath.Join(".var", "app", "com.valvesoftware.Steam", "data", "Steam"),
}

func InstallLatestProtonGE(ctx context.Context, logger *slog.Logger, report ProtonGEProgressFunc) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find home directory: %w", err)
	}
	installDir := compatibilityToolsDir(homeDir)
	logger.Info("installing ProtonGE", slog.String("path", installDir))
	client := &http.Client{}
	return installLatestProtonGE(ctx, client, protonGELatestReleaseURL, installDir, runtime.GOARCH, report)
}

func compatibilityToolsDir(homeDir string) string {
	for _, parent := range protonGEInstallParents {
		info, err := os.Stat(filepath.Join(homeDir, parent))
		if err == nil && info.IsDir() {
			return filepath.Join(homeDir, parent, "compatibilitytools.d")
		}
	}
	return filepath.Join(homeDir, ".local", "share", "Steam", "compatibilitytools.d")
}

func selectArchiveAsset(release githubRelease, arch string) (githubAsset, error) {
	var names []string
	switch arch {
	case "amd64":
		names = []string{release.TagName + "-x86_64.tar.gz"}
	case "arm64":
		names = []string{release.TagName + "-aarch64.tar.gz"}
	default:
		return githubAsset{}, fmt.Errorf("unsupported architecture %q for ProtonGE", arch)
	}
	names = append(names, release.TagName+".tar.gz")

	for _, name := range names {
		for _, asset := range release.Assets {
			if asset.Name == name {
				return asset, nil
			}
		}
	}
	for _, asset := range release.Assets {
		if asset.Name == release.TagName+".tar.xz" {
			return githubAsset{}, fmt.Errorf("release %s only provides a .tar.xz archive, which is not supported", release.TagName)
		}
	}
	return githubAsset{}, fmt.Errorf("release %s has no ProtonGE archive for %s", release.TagName, arch)
}

func installLatestProtonGE(ctx context.Context, client *http.Client, releaseURL, installDir, arch string, report ProtonGEProgressFunc) (string, error) {
	if report == nil {
		report = func(ProtonGEProgress) {}
	}
	release, err := fetchProtonGERelease(ctx, client, releaseURL)
	if err != nil {
		return "", err
	}
	asset, err := selectArchiveAsset(release, arch)
	if err != nil {
		return "", err
	}
	var checksumURL string
	for _, candidate := range release.Assets {
		if candidate.Name == asset.Name[:strings.Index(asset.Name, ".tar.gz")]+".sha512sum" {
			checksumURL = candidate.BrowserDownloadURL
		}
	}

	if err := os.MkdirAll(installDir, 0o755); err != nil {
		return "", fmt.Errorf("create compatibility tools directory: %w", err)
	}

	archive, err := os.CreateTemp("", "yayeet-protonge-*.tar.gz")
	if err != nil {
		return "", fmt.Errorf("create temporary archive: %w", err)
	}
	defer os.Remove(archive.Name())
	defer archive.Close()

	sum, err := downloadProtonGE(ctx, client, asset.BrowserDownloadURL, archive, report)
	if err != nil {
		return "", err
	}
	if checksumURL != "" {
		if err := verifyProtonGEChecksum(ctx, client, checksumURL, asset.Name, sum); err != nil {
			return "", err
		}
	}

	if _, err := archive.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("rewind archive: %w", err)
	}
	topLevel, err := extractProtonGE(archive, installDir, report)
	if err != nil {
		return "", err
	}

	for _, dir := range append([]string{release.TagName}, topLevel...) {
		if _, err := os.Stat(filepath.Join(installDir, dir, "files", "bin", "wine")); err == nil {
			return release.TagName, nil
		}
	}
	return "", fmt.Errorf("installed ProtonGE %s does not contain files/bin/wine", release.TagName)
}

func fetchProtonGERelease(ctx context.Context, client *http.Client, releaseURL string) (githubRelease, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, releaseURL, nil)
	if err != nil {
		return githubRelease{}, fmt.Errorf("create ProtonGE release request: %w", err)
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "YaYeet")
	response, err := client.Do(request)
	if err != nil {
		return githubRelease{}, fmt.Errorf("fetch ProtonGE release: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return githubRelease{}, fmt.Errorf("fetch ProtonGE release: unexpected HTTP status %s", response.Status)
	}
	var release githubRelease
	if err := json.NewDecoder(response.Body).Decode(&release); err != nil {
		return githubRelease{}, fmt.Errorf("decode ProtonGE release: %w", err)
	}
	if release.Draft || release.Prerelease || release.TagName == "" {
		return githubRelease{}, errors.New("no stable ProtonGE release found")
	}
	return release, nil
}

func downloadProtonGE(ctx context.Context, client *http.Client, url string, dst io.Writer, report ProtonGEProgressFunc) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create ProtonGE download request: %w", err)
	}
	request.Header.Set("User-Agent", "YaYeet")
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("download ProtonGE: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download ProtonGE: unexpected HTTP status %s", response.Status)
	}

	hash := sha512.New()
	progress := &progressReader{
		reader: response.Body,
		phase:  "download",
		total:  response.ContentLength,
		report: report,
	}
	if _, err := io.Copy(io.MultiWriter(dst, hash), progress); err != nil {
		return "", fmt.Errorf("download ProtonGE: %w", err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

type progressReader struct {
	reader  io.Reader
	phase   string
	current int64
	total   int64
	report  ProtonGEProgressFunc
	last    time.Time
}

func (p *progressReader) Read(buf []byte) (int, error) {
	n, err := p.reader.Read(buf)
	p.current += int64(n)
	if err != nil || time.Since(p.last) > 50*time.Millisecond {
		p.last = time.Now()
		p.report(ProtonGEProgress{Phase: p.phase, Current: p.current, Total: p.total})
	}
	return n, err
}

func verifyProtonGEChecksum(ctx context.Context, client *http.Client, url, archiveName, actual string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create ProtonGE checksum request: %w", err)
	}
	request.Header.Set("User-Agent", "YaYeet")
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("fetch ProtonGE checksum: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch ProtonGE checksum: unexpected HTTP status %s", response.Status)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read ProtonGE checksum: %w", err)
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return errors.New("ProtonGE checksum file is empty")
	}
	if !strings.EqualFold(fields[0], actual) {
		return fmt.Errorf("ProtonGE checksum mismatch for %s", archiveName)
	}
	return nil
}

// extractProtonGE returns the top-level directory names found in the archive.
func extractProtonGE(archive io.Reader, installDir string, report ProtonGEProgressFunc) ([]string, error) {
	gz, err := gzip.NewReader(archive)
	if err != nil {
		return nil, fmt.Errorf("open ProtonGE archive: %w", err)
	}
	defer gz.Close()

	root := filepath.Clean(installDir)
	reader := tar.NewReader(gz)
	seen := map[string]bool{}
	var topLevel []string
	var count int64
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read ProtonGE archive: %w", err)
		}
		target := filepath.Join(root, header.Name)
		if target != root && !strings.HasPrefix(target, root+string(os.PathSeparator)) {
			return nil, fmt.Errorf("ProtonGE archive entry %q escapes install directory", header.Name)
		}
		if rel, err := filepath.Rel(root, target); err == nil && rel != "." {
			top := strings.SplitN(filepath.ToSlash(rel), "/", 2)[0]
			if !seen[top] {
				seen[top] = true
				topLevel = append(topLevel, top)
			}
		}

		mode := header.FileInfo().Mode().Perm()
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return nil, fmt.Errorf("create directory %s: %w", target, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return nil, fmt.Errorf("create directory for %s: %w", target, err)
			}
			file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode|0o200)
			if err != nil {
				return nil, fmt.Errorf("create %s: %w", target, err)
			}
			_, copyErr := io.Copy(file, reader)
			closeErr := file.Close()
			if copyErr != nil {
				return nil, fmt.Errorf("extract %s: %w", target, copyErr)
			}
			if closeErr != nil {
				return nil, fmt.Errorf("close %s: %w", target, closeErr)
			}
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return nil, fmt.Errorf("create directory for %s: %w", target, err)
			}
			resolved := header.Linkname
			if !filepath.IsAbs(resolved) {
				resolved = filepath.Join(filepath.Dir(target), resolved)
			}
			if resolved != root && !strings.HasPrefix(filepath.Clean(resolved), root+string(os.PathSeparator)) {
				return nil, fmt.Errorf("ProtonGE archive symlink %q escapes install directory", header.Name)
			}
			_ = os.Remove(target)
			if err := os.Symlink(header.Linkname, target); err != nil {
				return nil, fmt.Errorf("create symlink %s: %w", target, err)
			}
		}
		count++
		report(ProtonGEProgress{Phase: "extract", Current: count})
	}
	return topLevel, nil
}
