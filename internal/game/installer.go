package game

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/folbricht/desync"
)

const storeURL = "https://votv.dev/patcher_assets/256-1024-4096-store"

type InstallProgress struct {
	Current int
	Total   int
}

type ProgressFunc func(InstallProgress)

func Install(ctx context.Context, indexURL, installationPath, cachePath string, report ProgressFunc) error {
	indexPath, err := downloadIndex(ctx, indexURL)
	if err != nil {
		return err
	}
	defer os.Remove(indexPath)

	indexFile, err := os.Open(indexPath)
	if err != nil {
		return fmt.Errorf("open install index: %w", err)
	}
	index, err := desync.IndexFromReader(indexFile)
	indexFile.Close()
	if err != nil {
		return fmt.Errorf("parse install index: %w", err)
	}

	if err := os.MkdirAll(cachePath, 0o755); err != nil {
		return fmt.Errorf("create download cache: %w", err)
	}
	if err := os.MkdirAll(installationPath, 0o755); err != nil {
		return fmt.Errorf("create installation directory: %w", err)
	}

	options := desync.NewStoreOptionsWithDefaults()
	options.N = 16
	remoteURL, err := url.Parse(storeURL)
	if err != nil {
		return err
	}
	remoteStore, err := desync.NewRemoteHTTPStore(remoteURL, options)
	if err != nil {
		return fmt.Errorf("create remote store: %w", err)
	}
	cacheStore, err := desync.NewLocalStore(cachePath, options)
	if err != nil {
		return fmt.Errorf("create local cache store: %w", err)
	}
	store := desync.NewStoreRouter(remoteStore, cacheStore)

	filesystem := desync.NewLocalFS(installationPath, desync.LocalFSOptions{
		NoSameOwner:       true,
		NoSamePermissions: true,
	})
	progress := &progressBar{report: report}
	if err := desync.UnTarIndex(ctx, filesystem, index, store, 16, progress); err != nil {
		return fmt.Errorf("install game files: %w", err)
	}
	return nil
}

func downloadIndex(ctx context.Context, indexURL string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, indexURL, nil)
	if err != nil {
		return "", err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("download install index: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("download install index: HTTP %s", response.Status)
	}

	file, err := os.CreateTemp("", "jirolauncher-*.caidx")
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(file, response.Body); err != nil {
		file.Close()
		os.Remove(file.Name())
		return "", err
	}
	if err := file.Close(); err != nil {
		os.Remove(file.Name())
		return "", err
	}
	return filepath.Clean(file.Name()), nil
}

type progressBar struct {
	report  ProgressFunc
	current int
	total   int
}

func (p *progressBar) SetTotal(total int)             { p.total = total; p.reportProgress() }
func (p *progressBar) Start()                         { p.reportProgress() }
func (p *progressBar) Finish()                        { p.current = p.total; p.reportProgress() }
func (p *progressBar) Increment() int                 { return p.Add(1) }
func (p *progressBar) Add(value int) int              { p.current += value; p.reportProgress(); return p.current }
func (p *progressBar) Set(value int)                  { p.current = value; p.reportProgress() }
func (p *progressBar) Write(data []byte) (int, error) { return len(data), nil }

func (p *progressBar) reportProgress() {
	if p.report != nil {
		p.report(InstallProgress{Current: p.current, Total: p.total})
	}
}
