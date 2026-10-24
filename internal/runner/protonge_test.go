package runner

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestSelectArchiveAsset(t *testing.T) {
	release := githubRelease{TagName: "GE-Proton99-1", Assets: []githubAsset{
		{Name: "GE-Proton99-1.tar.gz"},
		{Name: "GE-Proton99-1-aarch64.tar.gz"},
		{Name: "GE-Proton99-1-x86_64.tar.gz"},
	}}
	for arch, want := range map[string]string{"amd64": "GE-Proton99-1-x86_64.tar.gz", "arm64": "GE-Proton99-1-aarch64.tar.gz"} {
		got, err := selectArchiveAsset(release, arch)
		if err != nil || got.Name != want {
			t.Fatalf("%s: got %q, %v; want %q", arch, got.Name, err, want)
		}
	}
	if _, err := selectArchiveAsset(githubRelease{TagName: "T", Assets: []githubAsset{{Name: "T.tar.xz"}}}, "amd64"); err == nil {
		t.Fatal("expected error for xz-only release")
	}
}

func protonGETestArchive(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: "GE-Proton99-1/files/bin/wine", Mode: 0o755, Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	tw.Close()
	gz.Close()
	return buf.Bytes()
}

func protonGETestServer(t *testing.T, checksum string) *httptest.Server {
	t.Helper()
	archive := protonGETestArchive(t)
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/release":
			assets := []githubAsset{{Name: "GE-Proton99-1-x86_64.tar.gz", BrowserDownloadURL: server.URL + "/archive"}}
			if checksum != "" {
				assets = append(assets, githubAsset{Name: "GE-Proton99-1-x86_64.sha512sum", BrowserDownloadURL: server.URL + "/sum"})
			}
			json.NewEncoder(w).Encode(githubRelease{TagName: "GE-Proton99-1", Assets: assets})
		case "/archive":
			w.Write(archive)
		case "/sum":
			w.Write([]byte(checksum + "  GE-Proton99-1-x86_64.tar.gz\n"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func TestInstallLatestProtonGE(t *testing.T) {
	sum := sha512.Sum512(protonGETestArchive(t))
	server := protonGETestServer(t, hex.EncodeToString(sum[:]))
	installDir := filepath.Join(t.TempDir(), "missing", "compatibilitytools.d")

	tag, err := installLatestProtonGE(context.Background(), server.Client(), server.URL+"/release", installDir, "amd64", nil)
	if err != nil {
		t.Fatal(err)
	}
	if tag != "GE-Proton99-1" {
		t.Fatalf("tag = %q", tag)
	}
	if _, err := os.Stat(filepath.Join(installDir, tag, "files", "bin", "wine")); err != nil {
		t.Fatal(err)
	}
}

func TestInstallLatestProtonGEChecksumMismatch(t *testing.T) {
	server := protonGETestServer(t, "THISISAVERYLONGANDVERYOBVIOUSTESTVALUE")
	installDir := filepath.Join(t.TempDir(), "compatibilitytools.d")

	if _, err := installLatestProtonGE(context.Background(), server.Client(), server.URL+"/release", installDir, "amd64", nil); err == nil {
		t.Fatal("expected checksum error")
	}
}

func TestCompatibilityToolsDirDefault(t *testing.T) {
	home := t.TempDir()
	want := filepath.Join(home, ".local", "share", "Steam", "compatibilitytools.d")
	if got := compatibilityToolsDir(home); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if err := os.MkdirAll(filepath.Join(home, ".steam", "root"), 0o755); err != nil {
		t.Fatal(err)
	}
	want = filepath.Join(home, ".steam", "root", "compatibilitytools.d")
	if got := compatibilityToolsDir(home); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
