package update

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		left  string
		right string
		want  int
	}{
		{left: "v1.10.0", right: "1.9.0", want: 1},
		{left: "v1.3.2", right: "1.3.2", want: 0},
		{left: "v1.2.9", right: "1.3.0", want: -1},
	}
	for _, test := range tests {
		got, err := compareVersions(test.left, test.right)
		if err != nil {
			t.Fatalf("compareVersions(%q, %q): %v", test.left, test.right, err)
		}
		if got != test.want {
			t.Fatalf("compareVersions(%q, %q) = %d, want %d", test.left, test.right, got, test.want)
		}
	}
}

func TestCheckFindsStableUpdateAssets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		fmt.Fprint(writer, `{
			"tag_name":"v1.4.0",
			"html_url":"https://example.com/release",
			"draft":false,
			"prerelease":false,
			"assets":[
				{"name":"YaYeet-v1.4.0-x86_64.AppImage","browser_download_url":"https://example.com/yayeet"},
				{"name":"YaYeet-v1.4.0-x86_64.AppImage.sha256","browser_download_url":"https://example.com/checksum"}
			]
		}`)
	}))
	defer server.Close()

	release, available, err := check(context.Background(), server.Client(), server.URL, "1.3.2")
	if err != nil {
		t.Fatal(err)
	}
	if !available {
		t.Fatal("expected update to be available")
	}
	if release.Version != "1.4.0" || release.AppImageURL == "" || release.AppImageChecksumURL == "" {
		t.Fatalf("unexpected release: %+v", release)
	}
}

func TestCheckIgnoresPrerelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		fmt.Fprint(writer, `{"tag_name":"v2.0.0","html_url":"https://example.com/release","prerelease":true}`)
	}))
	defer server.Close()

	_, available, err := check(context.Background(), server.Client(), server.URL, "1.3.2")
	if err != nil {
		t.Fatal(err)
	}
	if available {
		t.Fatal("prerelease must not be offered as an update")
	}
}

func TestInstallAppImageRejectsChecksumMismatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/appimage":
			fmt.Fprint(writer, "unverified update")
		case "/checksum":
			fmt.Fprintf(writer, "%064d  YaYeet.AppImage\n", 0)
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	appImagePath := filepath.Join(t.TempDir(), "YaYeet.AppImage")
	if err := os.WriteFile(appImagePath, []byte("current AppImage"), 0o755); err != nil {
		t.Fatal(err)
	}
	release := Release{
		AppImageURL:         server.URL + "/appimage",
		AppImageChecksumURL: server.URL + "/checksum",
	}
	if err := installAppImage(context.Background(), server.Client(), release, appImagePath); err == nil {
		t.Fatal("expected checksum mismatch")
	}
	content, err := os.ReadFile(appImagePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "current AppImage" {
		t.Fatalf("current AppImage was modified: %q", content)
	}
}

func TestInstallAppImageReplacesVerifiedFile(t *testing.T) {
	updatedContent := []byte("updated AppImage")
	expectedHash := sha256.Sum256(updatedContent)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/appimage":
			writer.Write(updatedContent)
		case "/checksum":
			fmt.Fprintf(writer, "%x  YaYeet.AppImage\n", expectedHash)
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	appImagePath := filepath.Join(t.TempDir(), "YaYeet.AppImage")
	if err := os.WriteFile(appImagePath, []byte("old AppImage"), 0o755); err != nil {
		t.Fatal(err)
	}
	release := Release{
		AppImageURL:         server.URL + "/appimage",
		AppImageChecksumURL: server.URL + "/checksum",
	}
	if err := installAppImage(context.Background(), server.Client(), release, appImagePath); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(appImagePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != string(updatedContent) {
		t.Fatalf("installed content = %q, want %q", content, updatedContent)
	}
	info, err := os.Stat(appImagePath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatal("installed AppImage is not executable")
	}
}
