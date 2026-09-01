package game

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestLatestRevisionsKeepsHighestNumericRevision(t *testing.T) {
	versions := []VersionOption{
		{Name: "0.6.0"},
		{Name: "0.5.1_0010"},
		{Name: "0.5.1_0009"},
		{Name: "0.5.1_0008"},
		{Name: "0.5.0_0002"},
		{Name: "0.5.0_0001"},
		{Name: "0.4.0_0015_1"},
		{Name: "0.4.0_0015"},
		{Name: "0.4.0_0014_9"},
		{Name: "0.3.0_preview"},
	}

	filtered := LatestRevisions(versions)
	names := make([]string, 0, len(filtered))
	for _, version := range filtered {
		names = append(names, version.Name)
	}

	expected := []string{"0.6.0", "0.5.1_0010", "0.5.0_0002", "0.4.0_0015_1", "0.3.0_preview"}
	if !slices.Equal(names, expected) {
		t.Fatalf("LatestRevisions() = %v, want %v", names, expected)
	}
}

func TestPakHash(t *testing.T) {
	path := filepath.Join(t.TempDir(), "game.pak")
	if err := os.WriteFile(path, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	hash, err := pakHash(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if hash != "26C7827D889F6DA3" {
		t.Fatalf("pakHash() = %q, want %q", hash, "26C7827D889F6DA3")
	}
}

func TestDisplayVersionNameFormatsManifestRevision(t *testing.T) {
	tests := map[string]string{
		"0.9.0_0015_1": "0.9.0-15.1",
		"0.5.1_0009":   "0.5.1-9",
		"0.9.0n":       "0.9.0n",
	}

	for version, expected := range tests {
		if actual := displayVersionName(version); actual != expected {
			t.Errorf("displayVersionName(%q) = %q, want %q", version, actual, expected)
		}
	}
}
