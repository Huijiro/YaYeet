package update

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func InstallAppImage(ctx context.Context, release Release, appImagePath string) error {
	client := &http.Client{Timeout: 10 * time.Minute}
	return installAppImage(ctx, client, release, appImagePath)
}

func installAppImage(ctx context.Context, client *http.Client, release Release, appImagePath string) error {
	if release.AppImageURL == "" || release.AppImageChecksumURL == "" {
		return fmt.Errorf("release does not contain an AppImage and checksum")
	}
	if !filepath.IsAbs(appImagePath) {
		return fmt.Errorf("AppImage path is not absolute")
	}
	info, err := os.Lstat(appImagePath)
	if err != nil {
		return fmt.Errorf("inspect current AppImage: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("current AppImage is not a regular file")
	}

	expectedHash, err := fetchChecksum(ctx, client, release.AppImageChecksumURL)
	if err != nil {
		return err
	}

	temporary, err := os.CreateTemp(filepath.Dir(appImagePath), ".yayeet-update-*")
	if err != nil {
		return fmt.Errorf("create temporary AppImage: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)

	hash := sha256.New()
	if err := download(ctx, client, release.AppImageURL, io.MultiWriter(temporary, hash)); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return fmt.Errorf("sync downloaded AppImage: %w", err)
	}
	if err := temporary.Chmod(info.Mode().Perm() | 0o111); err != nil {
		temporary.Close()
		return fmt.Errorf("make downloaded AppImage executable: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close downloaded AppImage: %w", err)
	}

	actualHash := hash.Sum(nil)
	if subtle.ConstantTimeCompare(actualHash, expectedHash) != 1 {
		return fmt.Errorf("downloaded AppImage checksum does not match release checksum")
	}
	if err := os.Rename(temporaryPath, appImagePath); err != nil {
		return fmt.Errorf("replace current AppImage: %w", err)
	}
	return nil
}

func fetchChecksum(ctx context.Context, client *http.Client, checksumURL string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, checksumURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create checksum request: %w", err)
	}
	request.Header.Set("User-Agent", "YaYeet updater")
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("download AppImage checksum: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download AppImage checksum: unexpected HTTP status %s", response.Status)
	}
	content, err := io.ReadAll(io.LimitReader(response.Body, 4097))
	if err != nil {
		return nil, fmt.Errorf("read AppImage checksum: %w", err)
	}
	if len(content) > 4096 {
		return nil, fmt.Errorf("AppImage checksum file is too large")
	}
	fields := strings.Fields(string(content))
	if len(fields) == 0 {
		return nil, fmt.Errorf("AppImage checksum file is empty")
	}
	hash, err := hex.DecodeString(fields[0])
	if err != nil || len(hash) != sha256.Size {
		return nil, fmt.Errorf("AppImage checksum file is invalid")
	}
	return hash, nil
}

func download(ctx context.Context, client *http.Client, sourceURL string, destination io.Writer) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return fmt.Errorf("create download request: %w", err)
	}
	request.Header.Set("User-Agent", "YaYeet updater")
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("download release asset: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("download release asset: unexpected HTTP status %s", response.Status)
	}
	if _, err := io.Copy(destination, response.Body); err != nil {
		return fmt.Errorf("write release asset: %w", err)
	}
	return nil
}
